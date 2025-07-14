package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/helpers"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"strongpassword123"`
	Role     string `json:"role" example:"employee"`
}

type RegisterResponse struct {
	Id    string `json:"id" example:"uuid-or-id"`
	Email string `json:"email" example:"user@example.com"`
	Role  string `json:"role" example:"employee"`
}

// RegisterHandler godoc
// @Summary Регистрация нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} RegisterResponse "Пользователь успешно зарегистрирован"
// @Failure 400 {object} helpers.ErrorResponse "Некорректный запрос или валидация"
// @Failure 405 {object} helpers.ErrorResponse "Метод не разрешён"
// @Failure 409 {object} helpers.ErrorResponse "Email уже зарегистрирован"
// @Failure 500 {object} helpers.ErrorResponse "Ошибка сервера"
// @Router /register [post]
func RegisterHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			helpers.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" || req.Password == "" || req.Role == "" {
			helpers.WriteJSONError(w, "email, password and role are required", http.StatusBadRequest)
			return
		}

		allowedRoles := map[string]bool{"employee": true, "moderator": true}
		if !allowedRoles[req.Role] {
			helpers.WriteJSONError(w, "invalid role", http.StatusBadRequest)
			return
		}

		// Проверяем, что пользователь с таким email не существует
		var exists bool
		err = DBConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1 AND deleted_at IS NULL)`, req.Email).Scan(&exists)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		if exists {
			helpers.WriteJSONError(w, "email already registered", http.StatusConflict)
			return
		}

		// Хешируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "failed to hash password", http.StatusInternalServerError)
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
			helpers.WriteJSONError(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		resp := RegisterResponse{Id: userID, Email: req.Email, Role: req.Role}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201

		json.NewEncoder(w).Encode(resp)
	}
}
