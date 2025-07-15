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

type ReceptionRequest struct {
	PVZId string `json:"pvzId" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
}

type ReceptionResponse struct {
	Id       string `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	DateTime string `json:"dateTime" example:"2025-07-15T18:55:28.164Z"`
	PVZId    string `json:"pvzId" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Status   string `json:"status" example:"in_progress" validate:"required,oneof=in_progress close"`
}

const (
	in_progress = "in_progress"
	close       = "close"
)


// ReceptionHandler обрабатывает запросы для работы с приемками товаров
// @Summary Создание новой приемки товаров (только для сотрудников ПВЗ)
// @Description Создает новую приемку товаров для указанного ПВЗ. Требуется, чтобы предыдущая приемка была закрыта.
// @Tags Default
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body ReceptionRequest true "Данные для создания приемки"
// @Success 201 {object} ReceptionResponse "Приемка успешно создана"
// @Failure 400 {object} helpers.ErrorResponse "Неверный формат запроса"
// @Failure 401 {object} helpers.ErrorResponse "Неверный PVZ ID или предыдущая приемка не закрыта"
// @Failure 403 {object} helpers.ErrorResponse "Доступ запрещен"
// @Failure 405 {object} helpers.ErrorResponse "Метод не разрешен"
// @Failure 500 {object} helpers.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /receptions [post]
func ReceptionHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			helpers.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ReceptionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Проверка существования id pvz
		var exists bool
		err = DBConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM pvz WHERE id=$1)`, req.PVZId).Scan(&exists)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		if !exists {
			helpers.WriteJSONError(w, "invalid pvz_id", http.StatusUnauthorized)
			return
		}

		// Проверка статуста прошлой приемки
		err = DBConn.QueryRow(`SELECT EXISTS(SELECT 1 FROM receptions WHERE status=$1)`, in_progress).Scan(&exists)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "database error", http.StatusInternalServerError)
			return
		}
		if exists {
			helpers.WriteJSONError(w, "previous reception is not closed", http.StatusUnauthorized)
			return
		}

		var receptionId string
		var dateTime string
		var status string
		err = DBConn.QueryRow(`
            INSERT INTO receptions (date_time, pvz_id, status)
            VALUES ($1, $2, $3)
            RETURNING id, date_time, status
        `, time.Now().UTC().Format("2006-01-02T15:04:05.000Z"), req.PVZId, in_progress).Scan(&receptionId, &dateTime, &status)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "failed to create reception", http.StatusInternalServerError)
			return
		}

		resp := ReceptionResponse{Id: receptionId, DateTime: dateTime, PVZId: req.PVZId, Status: status}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated) // 201

		json.NewEncoder(w).Encode(resp)
	}
}
