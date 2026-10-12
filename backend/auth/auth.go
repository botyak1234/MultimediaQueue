package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"media-planner/backend/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type ContextKey string

const UserIDKey ContextKey = "userID"

var jwtSecret = []byte("media-planner-secret-key-portable")

type AuthHandler struct {
	DB *sql.DB
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, http.StatusBadRequest, "INVALID_INPUT", "Неверный формат запроса")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		RenderError(w, http.StatusBadRequest, "INVALID_USERNAME", ErrUsernameRequired.Error())
		return
	}

	if len(req.Password) < 4 {
		RenderError(w, http.StatusBadRequest, "INVALID_PASSWORD", ErrPasswordLength.Error())
		return
	}

	var existingID int64
	err := h.DB.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingID)
	if err == nil {
		RenderError(w, http.StatusConflict, "USER_EXISTS", ErrUserExists.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		RenderError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Ошибка при обработке пароля")
		return
	}

	res, err := h.DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", req.Username, string(hashedPassword))
	if err != nil {
		RenderError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Не удалось создать пользователя")
		return
	}

	userID, _ := res.LastInsertId()

	token, err := generateJWT(userID, req.Username)
	if err != nil {
		RenderError(w, http.StatusInternalServerError, "TOKEN_ERROR", "Не удалось сгенерировать токен")
		return
	}

	user := models.User{
		ID:        userID,
		Username:  req.Username,
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RenderError(w, http.StatusBadRequest, "INVALID_INPUT", "Неверный формат запроса")
		return
	}

	req.Username = strings.TrimSpace(req.Username)

	var user models.User
	err := h.DB.QueryRow("SELECT id, username, password_hash, created_at FROM users WHERE username = ?", req.Username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		RenderError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", ErrInvalidCreds.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		RenderError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", ErrInvalidCreds.Error())
		return
	}

	token, err := generateJWT(user.ID, user.Username)
	if err != nil {
		RenderError(w, http.StatusInternalServerError, "TOKEN_ERROR", "Не удалось сгенерировать токен")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		RenderError(w, http.StatusUnauthorized, "UNAUTHORIZED", ErrUnauthorized.Error())
		return
	}

	var user models.User
	err := h.DB.QueryRow("SELECT id, username, created_at FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		RenderError(w, http.StatusNotFound, "USER_NOT_FOUND", "Пользователь не найден")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			RenderError(w, http.StatusUnauthorized, "UNAUTHORIZED", ErrUnauthorized.Error())
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			RenderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Сессия истекла или недействительна")
			return
		}

		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			RenderError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Неверные данные сессии")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func generateJWT(userID int64, username string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
