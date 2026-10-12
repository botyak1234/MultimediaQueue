package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"media-planner/backend/db"
)

func TestAuth(t *testing.T) {
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer database.Close()

	handler := &AuthHandler{DB: database}

	// 1. Test Register
	regReq := AuthRequest{Username: "testuser", Password: "password123"}
	body, _ := json.Marshal(regReq)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var authResp AuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &authResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if authResp.Token == "" {
		t.Fatal("Expected JWT token in response, got empty")
	}

	// 2. Test Duplicate Register
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	handler.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("Expected 409 Conflict for duplicate user, got %d", rec.Code)
	}

	// 3. Test Login
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for login, got %d", rec.Code)
	}

	// 4. Test Middleware
	protectedReq := httptest.NewRequest("GET", "/api/auth/me", nil)
	protectedReq.Header.Set("Authorization", "Bearer "+authResp.Token)
	rec = httptest.NewRecorder()

	handler.Middleware(handler.Me)(rec, protectedReq)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for protected route, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}
