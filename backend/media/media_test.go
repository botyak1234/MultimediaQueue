package media

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"media-planner/backend/auth"
	"media-planner/backend/db"
	"media-planner/backend/models"
)

func TestMediaAndQueue(t *testing.T) {
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer database.Close()

	userID := createUser(t, database, "mediatest", "password")
	mediaH := &MediaHandler{DB: database}

	// 1. Create Movie
	movie := models.MediaItem{
		Type:            models.MediaTypeMovie,
		Title:           "Inception",
		ReleaseYear:     2010,
		DurationMinutes: 148,
	}
	body, _ := json.Marshal(movie)
	req := httptest.NewRequest("POST", "/api/media", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	rec := httptest.NewRecorder()

	mediaH.CreateMedia(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var createdMovie models.MediaItem
	json.Unmarshal(rec.Body.Bytes(), &createdMovie)

	// 2. Create Series
	series := models.MediaItem{
		Type:               models.MediaTypeSeries,
		Title:              "Breaking Bad",
		ReleaseYear:        2008,
		TotalEpisodes:      62,
		AvgEpisodeDuration: 45,
	}
	body, _ = json.Marshal(series)
	req = httptest.NewRequest("POST", "/api/media", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	rec = httptest.NewRecorder()

	mediaH.CreateMedia(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for series, got %d", rec.Code)
	}

	// 3. List Media
	req = httptest.NewRequest("GET", "/api/media", nil)
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	rec = httptest.NewRecorder()

	mediaH.ListMedia(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for list media, got %d", rec.Code)
	}

	var mediaList []models.MediaItem
	json.Unmarshal(rec.Body.Bytes(), &mediaList)
	if len(mediaList) != 2 {
		t.Fatalf("Expected 2 items in media list, got %d", len(mediaList))
	}
}

func createUser(t *testing.T, db *sql.DB, username, password string) int64 {
	res, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, 'hash')", username)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}
