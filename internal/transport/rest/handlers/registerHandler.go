package handlers

import (
	"GoPVZ/internal/lib/sl"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // employee или moderator
}

func RegisterHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" || req.Password == "" || req.Role == "" {
			http.Error(w, "email, password and role are required", http.StatusBadRequest)
			return
		}

		allowedRoles := map[string]bool{"employee": true, "moderator": true}
		if !allowedRoles[req.Role] {
			http.Error(w, "invalid role", http.StatusBadRequest)
			return
		}

		// Проверяем, что пользователь с таким email не существует
		var exists bool
		err := DBConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1 AND deleted_at IS NULL)`, req.Email).Scan(&exists)
		if err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "email already registered", http.StatusConflict)
			return
		}

		// Хешируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}

		// Вставляем пользователя в БД
		var userID string
		err = DBConn.QueryRow(`
            INSERT INTO users (email, password_hash, role)
            VALUES ($1, $2, $3)
            RETURNING id
        `, req.Email, string(hashedPassword), req.Role).Scan(&userID)
		if err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}
		

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(map[string]string{
			"id":    userID,
			"email": req.Email,
			"role":  req.Role,
		})
	}
}
