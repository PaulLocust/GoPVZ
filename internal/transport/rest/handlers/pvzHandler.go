package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/helpers"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type PVZRequest struct {
	City string `json:"city" example:"Москва" validate:"required,oneof=Москва Санкт-Петербург Казань"`
}

type PVZResponse struct {
	Id      string `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	RegDate string `json:"registrationDate" example:"2025-07-15T13:39:10.268Z"`
	City    string `json:"city" example:"Москва"`
}

// PVZHandler обрабатывает запросы для работы с ПВЗ
// @Summary Создание ПВЗ (только для модераторов)
// @Description Создает новый пункт выдачи заказов в указанном городе
// @Tags Default
// @Accept json
// @Produce json
// @Param input body PVZRequest true "Данные для создания ПВЗ"
// @Success 201 {object} PVZResponse "ПВЗ успешно создан"
// @Failure 400 {object} helpers.ErrorResponse "Неверный запрос"
// @Failure 401 {object} helpers.ErrorResponse "Ошибка авторизации"
// @Failure 403 {object} helpers.ErrorResponse "Доступ запрещен"
// @Failure 405 {object} helpers.ErrorResponse "Метод не разрешен"
// @Failure 500 {object} helpers.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /pvz [post]
func PVZHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			helpers.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req PVZRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		allowedCities := map[string]bool{
			moscow:           true,
			saint_petersburg: true,
			kazan:            true,
		}
		if !allowedCities[req.City] {
			helpers.WriteJSONError(w, "invalid city", http.StatusBadRequest)
			return
		}

		// Вставляем пользователя в БД
		var pvzID string
		var regDate string
		err = DBConn.QueryRow(`
            INSERT INTO pvz (city, registration_date)
            VALUES ($1, $2)
            RETURNING id, registration_date
        `, req.City, time.Now().UTC().Format("2006-01-02T15:04:05.000Z")).Scan(&pvzID, &regDate)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "failed to create pvz", http.StatusInternalServerError)
			return
		}
		resp := PVZResponse{Id: pvzID, RegDate: regDate, City: req.City}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(resp)
	}

}

// TODO: 4.Добавление информации о приёмке товаров

// TODO: 5.Добавление товаров в рамках одной приёмки

// TODO: 6.Удаление товаров в рамках незакрытой приёмки

// TODO: 7.Закрытие приёмки

// TODO: 8.Получение данных
