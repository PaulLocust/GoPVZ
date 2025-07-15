package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/helpers"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type ProductRequest struct {
	Type  string `json:"type" example:"электроника" validate:"required,oneof=электроника одежда обувь"`
	PVZId string `json:"pvzId" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
}

type ProductResponse struct {
	Id          string `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	DateTime    string `json:"dateTime" example:"2025-07-15T18:55:28.164Z"`
	Type        string `json:"type" example:"in_progress" validate:"required,oneof=электроника одежда обувь"`
	ReceptionId string `json:"receptionId" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
}

// ProductHandler обрабатывает запросы на добавление товаров
// @Summary Добавление товара в текущую приемку (только для сотрудников ПВЗ)
// @Description Добавляет новый товар в текущую открытую приемку для указанного ПВЗ
// @Tags Default
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body ProductRequest true "Данные товара"
// @Success 201 {object} ProductResponse "Товар успешно добавлен"
// @Failure 400 {object} helpers.ErrorResponse "Неверный запрос"
// @Failure 401 {object} helpers.ErrorResponse "Нет открытой приемки для указанного ПВЗ"
// @Failure 403 {object} helpers.ErrorResponse "Доступ запрещен"
// @Failure 405 {object} helpers.ErrorResponse "Метод не разрешен"
// @Failure 500 {object} helpers.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /products [post]
func ProductHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			helpers.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ProductRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		allowedTypes := map[string]bool{
			electronics: true,
			clothes:     true,
			shoes:       true,
		}
		if !allowedTypes[strings.ToLower(req.Type)] {
			helpers.WriteJSONError(w, "invalid product type", http.StatusBadRequest)
			return
		}

		// Проверка наличия нужной приёмки
		var exists bool
		err = DBConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM receptions WHERE pvz_id=$1 AND status=$2)`, req.PVZId, in_progress).Scan(&exists)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		if !exists {
			helpers.WriteJSONError(w, "no open reception found for this PVZ", http.StatusUnauthorized)
			return
		}

		// Берем id приёмки
		var receptionId string
		err = DBConn.QueryRow(`SELECT id FROM receptions WHERE pvz_id=$1 AND status=$2`, req.PVZId, in_progress).Scan(&receptionId)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "database error", http.StatusInternalServerError)
			return
		}

		// Создаём продукт
		var productID string
		var dateTime string
		err = DBConn.QueryRow(`
            INSERT INTO products (reception_id, date_time, type)
            VALUES ($1, $2, $3)
            RETURNING id, date_time
        `, receptionId, time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), req.Type).Scan(&productID, &dateTime)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "failed to create product", http.StatusInternalServerError)
			return
		}

		resp := ProductResponse{Id: productID, DateTime: dateTime, Type: req.Type, ReceptionId: receptionId}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(resp)

	}
}
