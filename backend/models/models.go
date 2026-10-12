package models

import "time"

type MediaType string

const (
	MediaTypeMovie  MediaType = "movie"
	MediaTypeSeries MediaType = "series"
	MediaTypeBook   MediaType = "book"
)

type QueueStatus string

const (
	StatusPlanned    QueueStatus = "planned"
	StatusInProgress QueueStatus = "in_progress"
	StatusCompleted  QueueStatus = "completed"
	StatusDropped    QueueStatus = "dropped"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type MediaItem struct {
	ID                 int64     `json:"id"`
	UserID             int64     `json:"user_id"`
	Type               MediaType `json:"type"`
	Title              string    `json:"title"`
	ReleaseYear        int       `json:"release_year"`
	DurationMinutes    int       `json:"duration_minutes"`     // For movies
	TotalEpisodes      int       `json:"total_episodes"`       // For series
	WatchedEpisodes    int       `json:"watched_episodes"`     // For series
	AvgEpisodeDuration int       `json:"avg_episode_duration"` // For series (minutes)
	TotalPages         int       `json:"total_pages"`          // For books
	ReadPages          int       `json:"read_pages"`           // For books
	AvgPageReadTime    int       `json:"avg_page_read_time"`   // For books (minutes per 10 pages, or per page)
	Description        string    `json:"description"`
	CreatedAt          time.Time `json:"created_at"`
}

type QueueItem struct {
	ID        int64       `json:"id"`
	UserID    int64       `json:"user_id"`
	MediaID   int64       `json:"media_id"`
	Status    QueueStatus `json:"status"`
	Priority  int         `json:"priority"`
	CreatedAt time.Time   `json:"created_at"`

	Media *MediaItem `json:"media,omitempty"`
}

// PlanRequest input from user
type PlanRequest struct {
	AvailableMinutes int `json:"available_minutes"`
}

// PlannedAction detail for free time planner output
type PlannedAction struct {
	MediaID          int64     `json:"media_id"`
	Title            string    `json:"title"`
	Type             MediaType `json:"type"`
	UnitsToConsume   int       `json:"units_to_consume"`   // count of episodes, pages, or 1 movie
	UnitName         string    `json:"unit_name"`          // "эпизод", "страниц", "фильм"
	DurationMinutes  int       `json:"duration_minutes"`  // consumed time
	Description      string    `json:"description"`
}

type PlanResponse struct {
	AvailableMinutes int             `json:"available_minutes"`
	UsedMinutes      int             `json:"used_minutes"`
	RemainingMinutes int             `json:"remaining_minutes"`
	Actions          []PlannedAction `json:"actions"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type HTTPError struct {
	Error APIError `json:"error"`
}
