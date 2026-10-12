package planner

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"media-planner/backend/auth"
	"media-planner/backend/models"
)

type PlannerHandler struct {
	DB *sql.DB
}

func (h *PlannerHandler) GeneratePlan(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int64)

	var req models.PlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_INPUT", "Некорректные параметры свободного времени")
		return
	}

	if req.AvailableMinutes <= 0 {
		auth.RenderError(w, http.StatusBadRequest, "INVALID_MINUTES", "Свободное время должно быть больше 0 минут")
		return
	}

	// Fetch items from user's planned queue ordered by priority
	rows, err := h.DB.Query(`
		SELECT m.id, m.type, m.title, m.release_year, m.duration_minutes,
		       m.total_episodes, m.watched_episodes, m.avg_episode_duration,
			   m.total_pages, m.read_pages, m.avg_page_read_time, m.description
		FROM queue_items q
		JOIN media_items m ON q.media_id = m.id
		WHERE q.user_id = ? AND q.status IN ('planned', 'in_progress')
		ORDER BY q.priority ASC, q.created_at ASC`, userID)
	if err != nil {
		auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Не удалось загрузить элементы очереди для расчета плана")
		return
	}
	defer rows.Close()

	items := []models.MediaItem{}
	for rows.Next() {
		var item models.MediaItem
		err := rows.Scan(
			&item.ID, &item.Type, &item.Title, &item.ReleaseYear, &item.DurationMinutes,
			&item.TotalEpisodes, &item.WatchedEpisodes, &item.AvgEpisodeDuration,
			&item.TotalPages, &item.ReadPages, &item.AvgPageReadTime, &item.Description,
		)
		if err != nil {
			auth.RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Ошибка считывания медиаданных")
			return
		}
		items = append(items, item)
	}

	plan := CalculatePlan(items, req.AvailableMinutes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func CalculatePlan(items []models.MediaItem, availableMins int) models.PlanResponse {
	remaining := availableMins
	used := 0
	actions := []models.PlannedAction{}

	for _, item := range items {
		if remaining <= 0 {
			break
		}

		switch item.Type {
		case models.MediaTypeMovie:
			if item.DurationMinutes <= remaining {
				actions = append(actions, models.PlannedAction{
					MediaID:         item.ID,
					Title:           item.Title,
					Type:            item.Type,
					UnitsToConsume:  1,
					UnitName:        "фильм целиком",
					DurationMinutes: item.DurationMinutes,
					Description:     fmt.Sprintf("Посмотреть фильм «%s» (%d мин)", item.Title, item.DurationMinutes),
				})
				remaining -= item.DurationMinutes
				used += item.DurationMinutes
			}

		case models.MediaTypeSeries:
			remainingEpisodes := item.TotalEpisodes - item.WatchedEpisodes
			if remainingEpisodes <= 0 || item.AvgEpisodeDuration <= 0 {
				continue
			}

			episodesCanFit := remaining / item.AvgEpisodeDuration
			if episodesCanFit > remainingEpisodes {
				episodesCanFit = remainingEpisodes
			}

			if episodesCanFit > 0 {
				timeSpent := episodesCanFit * item.AvgEpisodeDuration
				actions = append(actions, models.PlannedAction{
					MediaID:         item.ID,
					Title:           item.Title,
					Type:            item.Type,
					UnitsToConsume:  episodesCanFit,
					UnitName:        formatEpisodesUnit(episodesCanFit),
					DurationMinutes: timeSpent,
					Description:     fmt.Sprintf("Посмотреть %d %s сериала «%s» (%d мин)", episodesCanFit, formatEpisodesUnit(episodesCanFit), item.Title, timeSpent),
				})
				remaining -= timeSpent
				used += timeSpent
			}

		case models.MediaTypeBook:
			remainingPages := item.TotalPages - item.ReadPages
			if remainingPages <= 0 || item.AvgPageReadTime <= 0 {
				continue
			}

			// AvgPageReadTime is minutes per page (e.g. 2 min/page)
			pagesCanFit := remaining / item.AvgPageReadTime
			if pagesCanFit > remainingPages {
				pagesCanFit = remainingPages
			}

			if pagesCanFit > 0 {
				timeSpent := pagesCanFit * item.AvgPageReadTime
				actions = append(actions, models.PlannedAction{
					MediaID:         item.ID,
					Title:           item.Title,
					Type:            item.Type,
					UnitsToConsume:  pagesCanFit,
					UnitName:        formatPagesUnit(pagesCanFit),
					DurationMinutes: timeSpent,
					Description:     fmt.Sprintf("Прочитать %d %s книги «%s» (%d мин)", pagesCanFit, formatPagesUnit(pagesCanFit), item.Title, timeSpent),
				})
				remaining -= timeSpent
				used += timeSpent
			}
		}
	}

	return models.PlanResponse{
		AvailableMinutes: availableMins,
		UsedMinutes:      used,
		RemainingMinutes: remaining,
		Actions:          actions,
	}
}

func formatEpisodesUnit(n int) string {
	if n%10 == 1 && n%100 != 11 {
		return "серию"
	}
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return "серии"
	}
	return "серий"
}

func formatPagesUnit(n int) string {
	if n%10 == 1 && n%100 != 11 {
		return "страницу"
	}
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return "страницы"
	}
	return "страниц"
}
