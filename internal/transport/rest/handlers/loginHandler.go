package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/jwt_gen"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}


func LoginHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req loginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" || req.Password == "" {
			http.Error(w, "email and password are required", http.StatusBadRequest)
			return
		}

		var exists bool
		err = DBConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1 AND deleted_at IS NULL)`, req.Email).Scan(&exists)
		if err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "database error1", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "email is not registered", http.StatusConflict)
			return
		}

		var id string
		var password_hash string
		var role string
		err = DBConn.QueryRow(`SELECT id, password_hash, role FROM users WHERE email=$1 AND deleted_at IS NULL`, req.Email).Scan(&id, &password_hash, &role)
		if err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "database error2", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(password_hash), []byte(req.Password))
		if err != nil {
			log.Error("error message", sl.Err(err))
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				http.Error(w, "invalid password", http.StatusUnauthorized) // 401
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		token, err := jwt_gen.GenerateJWT(role, id)
		if err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "cannot generate token", http.StatusInternalServerError)
			return
		}

		resp := loginResponse{Token: token}
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(resp)
	}
}
