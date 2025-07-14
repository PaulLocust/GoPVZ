package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/jwt_gen"
	"encoding/json"
	"log/slog"
	"net/http"
)

// TODO: 1.Авторизация пользователей /dummyLogin
type loginRequest struct {
	Role   string `json:"role"`    // client, moderator и т.п.
	UserID int    `json:"user_id"` // можно передать ID или сгенерировать
}

type loginResponse struct {
	Token string `json:"token"`
}

func DummyLoginHandler(log *slog.Logger) http.HandlerFunc {

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

		// Можно добавить валидацию роли, чтобы разрешать только нужные
		allowedRoles := map[string]bool{
			"moderator": true,
			"employee":  true,
		}
		if !allowedRoles[req.Role] {
			http.Error(w, "invalid role", http.StatusBadRequest)
			return
		}

		// Если user_id не передан, можно сгенерировать, например, 1
		if req.UserID == 0 {
			req.UserID = 1
		}

		token, err := jwt_gen.GenerateJWT(req.Role, req.UserID)
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

// TODO: 2.Регистрация и авторизация пользователей по почте и паролю
