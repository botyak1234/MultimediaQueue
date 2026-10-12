package media

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"media-planner/backend/auth"
	"media-planner/backend/models"
)

type MediaHandler struct {
	DB *sql.DB
}

func (h *MediaHandler) ListMedia(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int64)

	rows, err := h.DB.Query(`
		SELECT id, user_id, type, title, release_year, duration_minutes,
		       total_episodes, watched_episodes, avg_episode_duration,
			   total_pages, read_pages, avg_page_read_time, description, created_at
		FROM media_items
		WHERE user_id = ?
		ORDER BY created_at DESC`, userID)
	if err != nil {
		auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Ошибка при получении списка контента")
		return
	}
	defer rows.Close()

	items := []models.MediaItem{}
	for rows.Next() {
		var item models.MediaItem
		err := rows.Scan(
			&item.ID, &item.UserID, &item.Type, &item.Title, &item.ReleaseYear,
			&item.DurationMinutes, &item.TotalEpisodes, &item.WatchedEpisodes, &item.AvgEpisodeDuration,
			&item.TotalPages, &item.ReadPages, &item.AvgPageReadTime, &item.Description, &item.CreatedAt,
		)
		if err != nil {
			auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Ошибка считывания данных")
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *MediaHandler) CreateMedia(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int64)

	var item models.MediaItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_INPUT", "Некорректный JSON в запросе")
		return
	}

	item.Title = strings.TrimSpace(item.Title)
	if item.Title == "" {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_TITLE", "Название не может быть пустым")
		return
	}

	if item.Type != models.MediaTypeMovie && item.Type != models.MediaTypeSeries && item.Type != models.MediaTypeBook {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_TYPE", "Неподдерживаемый тип медиа. Допустимые: movie, series, book")
		return
	}

	// Validation per type
	switch item.Type {
	case models.MediaTypeMovie:
		if item.DurationMinutes <= 0 {
			auth.RenderError(w, http.StatusBadRequest, "INVALID_DURATION", "Продолжительность фильма должна быть больше 0 минут")
			return
		}
	case models.MediaTypeSeries:
		if item.TotalEpisodes <= 0 {
			auth.RenderError(w, http.StatusBadRequest, "INVALID_EPISODES", "Количество серий должно быть больше 0")
			return
		}
		if item.AvgEpisodeDuration <= 0 {
			auth.RenderError(w, http.StatusBadRequest, "INVALID_DURATION", "Средняя продолжительность серии должна быть больше 0 минут")
			return
		}
	case models.MediaTypeBook:
		if item.TotalPages <= 0 {
			auth.RenderError(w, http.StatusBadRequest, "INVALID_PAGES", "Количество страниц должно быть больше 0")
			return
		}
		if item.AvgPageReadTime <= 0 {
			item.AvgPageReadTime = 2 // Default 2 min per page if not provided
		}
	}

	res, err := h.DB.Exec(`
		INSERT INTO media_items (
			user_id, type, title, release_year, duration_minutes,
			total_episodes, watched_episodes, avg_episode_duration,
			total_pages, read_pages, avg_page_read_time, description
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, item.Type, item.Title, item.ReleaseYear, item.DurationMinutes,
		item.TotalEpisodes, item.WatchedEpisodes, item.AvgEpisodeDuration,
		item.TotalPages, item.ReadPages, item.AvgPageReadTime, item.Description,
	)
	if err != nil {
		auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Не удалось сохранить медиа-контент")
		return
	}

	item.ID, _ = res.LastInsertId()
	item.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *MediaHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int64)

	pathParts := strings.Split(r.URL.Path, "/")
	idStr := pathParts[len(pathParts)-1]
	mediaID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID медиа")
		return
	}

	res, err := h.DB.Exec("DELETE FROM media_items WHERE id = ? AND user_id = ?", mediaID, userID)
	if err != nil {
		auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Не удалось удалить элемент")
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		auth.RenderError(w, http.StatusNotFound, "NOT_FOUND", "Контент не найден")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MediaHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int64)

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_PATH", "Некорректный URL")
		return
	}
	mediaID, err := strconv.ParseInt(pathParts[len(pathParts)-2], 10, 64)
	if err != nil {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_ID", "Некорректный ID медиа")
		return
	}

	var req struct {
		WatchedEpisodes *int `json:"watched_episodes"`
		ReadPages       *int `json:"read_pages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_INPUT", "Некорректные данные прогресса")
		return
	}

	if req.WatchedEpisodes != nil {
		_, err = h.DB.Exec("UPDATE media_items SET watched_episodes = ? WHERE id = ? AND user_id = ?", *req.WatchedEpisodes, mediaID, userID)
	} else if req.ReadPages != nil {
		_, err = h.DB.Exec("UPDATE media_items SET read_pages = ? WHERE id = ? AND user_id = ?", *req.ReadPages, mediaID, userID)
	}

	if err != nil {
		auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Не удалось обновить прогресс")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}
