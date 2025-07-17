package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/helpers"
	"GoPVZ/internal/transport/rest/jwt_gen"
	"encoding/json"
	"log/slog"
	"net/http"
)

type DummyLoginRequest struct {
	Role string `json:"role" example:"moderator"`
}

type DummyLoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// DummyLoginHandler godoc
// @Summary Получение тестового токена
// @Tags Public
// @Accept json
// @Produce json
// @Param dummyLoginRequest body DummyLoginRequest true "Данные для входа (role и user_id)"
// @Success 200 {object} DummyLoginResponse "Успешная генерация токена"
// @Failure 400 {object} helpers.ErrorResponse "Некорректный запрос"
// @Failure 405 {object} helpers.ErrorResponse "Метод не разрешён"
// @Failure 500 {object} helpers.ErrorResponse "Внутренняя ошибка сервера"
// @Router /dummyLogin [post]
func DummyLoginHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			helpers.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req DummyLoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("error decoding request body", sl.Err(err))
			helpers.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		allowedRoles := map[string]bool{
			"moderator": true,
			"employee":  true,
		}
		if !allowedRoles[req.Role] {
			helpers.WriteJSONError(w, "invalid role", http.StatusBadRequest)
			return
		}

		token, err := jwt_gen.GenerateJWT(req.Role)
		if err != nil {
			log.Error("error generating token", sl.Err(err))
			helpers.WriteJSONError(w, "cannot generate token", http.StatusInternalServerError)
			return
		}

		resp := DummyLoginResponse{Token: token}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
