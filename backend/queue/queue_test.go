package queue

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

func TestQueueOperations(t *testing.T) {
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer database.Close()

	userID := createUser(t, database, "queuetest")

	// Insert test media item
	res, err := database.Exec(`
		INSERT INTO media_items (user_id, type, title, release_year, duration_minutes)
		VALUES (?, 'movie', 'Test Movie', 2022, 120)`, userID)
	if err != nil {
		t.Fatalf("Failed to insert media item: %v", err)
	}
	mediaID, _ := res.LastInsertId()

	qHandler := &QueueHandler{DB: database}

	// Add to queue
	addReq := map[string]interface{}{
		"media_id": mediaID,
		"status":   models.StatusPlanned,
		"priority": 1,
	}
	body, _ := json.Marshal(addReq)
	req := httptest.NewRequest("POST", "/api/queue", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	rec := httptest.NewRecorder()

	qHandler.AddToQueue(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for AddToQueue, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// List Queue
	req = httptest.NewRequest("GET", "/api/queue", nil)
	req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, userID))
	rec = httptest.NewRecorder()

	qHandler.ListQueue(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for ListQueue, got %d", rec.Code)
	}

	var queueList []models.QueueItem
	json.Unmarshal(rec.Body.Bytes(), &queueList)
	if len(queueList) != 1 {
		t.Fatalf("Expected 1 queue item, got %d", len(queueList))
	}
	if queueList[0].Media == nil || queueList[0].Media.Title != "Test Movie" {
		t.Fatalf("Media item not populated properly in queue item")
	}
}

func createUser(t *testing.T, db *sql.DB, username string) int64 {
	res, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, 'hash')", username)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}
