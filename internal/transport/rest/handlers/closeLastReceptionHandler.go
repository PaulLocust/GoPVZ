package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/helpers"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

type CloseLastReceptionResponse struct {
	Id       string `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	PVZId    string `json:"pvzId" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	DateTime string `json:"dateTime" example:"2025-07-17T12:15:49.386Z"`
	Status   string `json:"status" example:"close"`
}


// CloseLastReceptionHandler godoc
// @Summary Закрытие последней открытой приемки товаров в рамках ПВЗ (только для сотрудников ПВЗ)
// @Description Закрывает последнюю открытую приемку для указанного ПВЗ (меняет статус на "closed")
// @Tags Protected
// @Accept json
// @Produce json
// @Param pvzId path string true "ID пункта выдачи заказов (ПВЗ)"
// @Success 200 {object} CloseLastReceptionResponse "Приемка успешно закрыта"
// @Failure 400 {object} helpers.ErrorResponse "Некорректный путь запроса"
// @Failure 403 {object} helpers.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} helpers.ErrorResponse "Не найдено открытой приемки для данного ПВЗ"
// @Failure 405 {object} helpers.ErrorResponse "Метод не разрешен (разрешен только POST)"
// @Failure 500 {object} helpers.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /pvz/{pvzId}/close_last_reception [post]
func CloseLastReceptionHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			helpers.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Извлекаем pvzId из пути
		urlPath := r.URL.Path

		// Путь должен быть вида /pvz/{pvzId}/close_last_reception
		parts := strings.Split(urlPath, "/")

		// Если путь корректный, то извлекаем pvzId
		var pvzId string
		if len(parts) >= 4 {
			pvzId = parts[2]
		} else {
			helpers.WriteJSONError(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Проверяем существование открытой приемки
		var exists bool
		err := DBConn.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM receptions WHERE pvz_id = $1 AND status = $2)`, pvzId, in_progress).Scan(&exists)

		if err != nil {
			log.Error("Database error", sl.Err(err))
			helpers.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		if !exists {
			helpers.WriteJSONError(w, "no open reception found for this PVZ", http.StatusNotFound)
			return
		}

		var receptionId string
		var dateTime string
		var status string
		err = DBConn.QueryRow(`
            UPDATE receptions SET status = $1
			WHERE pvz_id = $2 AND status = $3
            RETURNING id, date_time, status
        `, close, pvzId, in_progress).Scan(&receptionId, &dateTime, &status)
		if err != nil {
			log.Error("error message", sl.Err(err))
			helpers.WriteJSONError(w, "failed to close reception", http.StatusInternalServerError)
			return
		}

		resp := CloseLastReceptionResponse{Id: receptionId, DateTime: dateTime, PVZId: pvzId, Status: status}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(resp)
	}
}
