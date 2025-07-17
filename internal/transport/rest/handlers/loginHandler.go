package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/helpers"
	"GoPVZ/internal/transport/rest/jwt_gen"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret"`
}

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// LoginHandler godoc
// @Summary Авторизация пользователя
// @Tags Public
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Данные для входа"
// @Success 200 {object} LoginResponse "Успешная авторизация"
// @Failure 400 {object} helpers.ErrorResponse "Неверный запрос"
// @Failure 401 {object} helpers.ErrorResponse "Неверный email или пароль"
// @Failure 405 {object} helpers.ErrorResponse "Метод не разрешён"
// @Failure 500 {object} helpers.ErrorResponse "Внутренняя ошибка сервера"
// @Router /login [post]
func LoginHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			helpers.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" || req.Password == "" {
			helpers.WriteJSONError(w, "email and password are required", http.StatusBadRequest)
			return
		}

		var exists bool
		err = DBConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, req.Email).Scan(&exists)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		if !exists {
			helpers.WriteJSONError(w, "invalid email", http.StatusUnauthorized)
			return
		}

		var password_hash string
		var role string
		err = DBConn.QueryRow(`SELECT id, password_hash, role FROM users WHERE email=$1`, req.Email).Scan(&password_hash, &role)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(req.Password))
		if err != nil {
			log.Error("error message", sl.Err(err))
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				helpers.WriteJSONError(w, "invalid password", http.StatusUnauthorized) // 401
				return
			}
			helpers.WriteJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		token, err := jwt_gen.GenerateJWT(role)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "cannot generate token", http.StatusInternalServerError)
			return
		}

		resp := LoginResponse{Token: token}
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(resp)
	}
}
