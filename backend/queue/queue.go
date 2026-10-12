package queue

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"media-planner/backend/auth"
	"media-planner/backend/models"
)

type QueueHandler struct {
	DB *sql.DB
}

func (h *QueueHandler) ListQueue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int64)

	rows, err := h.DB.Query(`
		SELECT q.id, q.user_id, q.media_id, q.status, q.priority, q.created_at,
		       m.id, m.user_id, m.type, m.title, m.release_year, m.duration_minutes,
			   m.total_episodes, m.watched_episodes, m.avg_episode_duration,
			   m.total_pages, m.read_pages, m.avg_page_read_time, m.description, m.created_at
		FROM queue_items q
		JOIN media_items m ON q.media_id = m.id
		WHERE q.user_id = ?
		ORDER BY q.priority ASC, q.created_at ASC`, userID)
	if err != nil {
		auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Ошибка при загрузке очереди")
		return
	}
	defer rows.Close()

	items := []models.QueueItem{}
	for rows.Next() {
		var qItem models.QueueItem
		var mItem models.MediaItem
		err := rows.Scan(
			&qItem.ID, &qItem.UserID, &qItem.MediaID, &qItem.Status, &qItem.Priority, &qItem.CreatedAt,
			&mItem.ID, &mItem.UserID, &mItem.Type, &mItem.Title, &mItem.ReleaseYear, &mItem.DurationMinutes,
			&mItem.TotalEpisodes, &mItem.WatchedEpisodes, &mItem.AvgEpisodeDuration,
			&mItem.TotalPages, &mItem.ReadPages, &mItem.AvgPageReadTime, &mItem.Description, &mItem.CreatedAt,
		)
		if err != nil {
			auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Ошибка считывания элементов очереди")
			return
		}
		qItem.Media = &mItem
		items = append(items, qItem)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *QueueHandler) AddToQueue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int64)

	var req struct {
		MediaID  int64              `json:"media_id"`
		Status   models.QueueStatus `json:"status"`
		Priority int                `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_INPUT", "Некорректный JSON в запросе")
		return
	}

	if req.MediaID <= 0 {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_MEDIA_ID", "Не указан контент для добавления в очередь")
		return
	}

	if req.Status == "" {
		req.Status = models.StatusPlanned
	}

	var exists bool
	err := h.DB.QueryRow("SELECT 1 FROM media_items WHERE id = ? AND user_id = ?", req.MediaID, userID).Scan(&exists)
	if err != nil || !exists {
		auth.RenderError(w, http.StatusNotFound, "MEDIA_NOT_FOUND", "Указанный контент не найден в вашей библиотеке")
		return
	}

	res, err := h.DB.Exec(`
		INSERT INTO queue_items (user_id, media_id, status, priority)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, media_id) DO UPDATE SET
			status = excluded.status,
			priority = excluded.priority`,
		userID, req.MediaID, req.Status, req.Priority,
	)
	if err != nil {
		auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Не удалось добавить в очередь")
		return
	}

	id, _ := res.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       id,
		"status":   req.Status,
		"priority": req.Priority,
		"message":  "Элемент добавлен в очередь",
	})
}

func (h *QueueHandler) UpdateQueueStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int64)

	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	queueID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID очереди")
		return
	}

	var req struct {
		Status   *models.QueueStatus `json:"status"`
		Priority *int                `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_INPUT", "Некорректный запрос")
		return
	}

	if req.Status != nil {
		_, err = h.DB.Exec("UPDATE queue_items SET status = ? WHERE id = ? AND user_id = ?", *req.Status, queueID, userID)
	}
	if req.Priority != nil {
		_, err = h.DB.Exec("UPDATE queue_items SET priority = ? WHERE id = ? AND user_id = ?", *req.Priority, queueID, userID)
	}

	if err != nil {
		auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Не удалось обновить элемент очереди")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"updated"}`))
}

func (h *QueueHandler) DeleteFromQueue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int64)

	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	queueID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID очереди")
		return
	}

	_, err = h.DB.Exec("DELETE FROM queue_items WHERE id = ? AND user_id = ?", queueID, userID)
	if err != nil {
		auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Не удалось удалить элемент из очереди")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
