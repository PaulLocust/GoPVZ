package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/jwt_gen"
	"encoding/json"
	"log/slog"
	"net/http"
)

// TODO: 1.Авторизация пользователей /dummyLogin
type dummyAuthRequest struct {
	Role   string `json:"role"`
	UserID string `json:"user_id"`
}

type dummyAuthResponse struct {
	Token string `json:"token"`
}

func DummyLoginHandler(log *slog.Logger) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req dummyAuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// валидация роли
		allowedRoles := map[string]bool{
			"moderator": true,
			"employee":  true,
		}
		if !allowedRoles[req.Role] {
			http.Error(w, "invalid role", http.StatusBadRequest)
			return
		}

		token, err := jwt_gen.GenerateJWT(req.Role, req.UserID)
		if err != nil {
			log.Error("error message", sl.Err(err))
			http.Error(w, "cannot generate token", http.StatusInternalServerError)
			return
		}

		resp := dummyAuthResponse{Token: token}
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(resp)
	}
}
