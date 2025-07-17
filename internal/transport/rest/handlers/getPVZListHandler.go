package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/helpers"
	"GoPVZ/internal/transport/rest/models"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// GetPVZListHandler godoc
// @Summary Получение списка ПВЗ с фильтрацией по дате приемки и пагинацией (только для сотрудников или модераторов)
// @Description Возвращает список ПВЗ с вложенной информацией о приемках и товарах за указанный период
// @Tags Protected
// @Produce json
// @Param startDate query string false "Начальная дата диапазона (формат 2025-07-17T12:45:55.122Z)"
// @Param endDate query string false "Конечная дата диапазона (формат 2025-07-17T12:45:55.122Z)"
// @Param page query int false "Номер страницы (по умолчанию 1)"
// @Param limit query int false "Количество элементов на странице (по умолчанию 10)"
// @Success 200 {array} models.PVZWithReceptionsResponse
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse "Доступ запрещен"
// @Failure 500 {object} helpers.ErrorResponse
// @Security BearerAuth
// @Router /pvz [get]
func GetPVZListHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			helpers.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Парсинг параметров запроса
		query := r.URL.Query()

		// Параметры пагинации
		page, err := strconv.Atoi(query.Get("page"))
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(query.Get("limit"))
		if err != nil || limit < 1 {
			limit = 10
		}
		offset := (page - 1) * limit

		// Параметры фильтрации по дате
		startDate := query.Get("startDate")
		endDate := query.Get("endDate")

		// Проверка валидности дат
		if startDate != "" {
			if _, err := time.Parse(time.RFC3339, startDate); err != nil {
				helpers.WriteJSONError(w, "Invalid startDate format", http.StatusBadRequest)
				return
			}
		}
		if endDate != "" {
			if _, err := time.Parse(time.RFC3339, endDate); err != nil {
				helpers.WriteJSONError(w, "Invalid endDate format", http.StatusBadRequest)
				return
			}
		}

		// Получаем список ПВЗ с пагинацией
		pvzQuery := `
            SELECT id, registration_date, city 
            FROM pvz
            ORDER BY registration_date DESC
            LIMIT $1 OFFSET $2
        `
		pvzRows, err := DBConn.Query(pvzQuery, limit, offset)
		if err != nil {
			log.Error("Database error (pvz)", sl.Err(err))
			helpers.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer pvzRows.Close()

		var response []models.PVZWithReceptionsResponse

		for pvzRows.Next() {
			var pvz models.PVZ
			err := pvzRows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
			if err != nil {
				log.Error("Error scanning pvz row", sl.Err(err))
				continue
			}

			// Для каждого ПВЗ получаем приемки в указанном диапазоне дат
			receptionQuery := `
                SELECT id, date_time, status
    			FROM receptions
    			WHERE pvz_id = $1
    			AND ($2 = '' OR date_time >= to_timestamp($2, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
    			AND ($3 = '' OR date_time <= to_timestamp($3, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
    			ORDER BY date_time DESC
            `
			receptionRows, err := DBConn.Query(receptionQuery, pvz.ID, startDate, endDate)
			if err != nil {
				log.Error("Database error (receptions)", sl.Err(err))
				helpers.WriteJSONError(w, "Database error", http.StatusInternalServerError)
				return
			}

			var receptions []models.ReceptionWithProducts
			for receptionRows.Next() {
				var reception models.Reception
				err := receptionRows.Scan(&reception.ID, &reception.DateTime, &reception.Status)
				if err != nil {
					log.Error("Error scanning reception row", sl.Err(err))
					continue
				}

				// Для каждой приемки получаем товары
				productsQuery := `
                    SELECT id, date_time, type
                    FROM products
                    WHERE reception_id = $1
                    ORDER BY date_time DESC
                `
				productRows, err := DBConn.Query(productsQuery, reception.ID)
				if err != nil {
					log.Error("Database error (products)", sl.Err(err))
					helpers.WriteJSONError(w, "Database error", http.StatusInternalServerError)
					return
				}

				var products []models.Product
				for productRows.Next() {
					var product models.Product
					err := productRows.Scan(&product.ID, &product.DateTime, &product.Type)
					if err != nil {
						log.Error("Error scanning product row", sl.Err(err))
						continue
					}
					product.ReceptionID = reception.ID
					products = append(products, product)
				}
				productRows.Close()

				receptions = append(receptions, models.ReceptionWithProducts{
					Reception: reception,
					Products:  products,
				})
			}
			receptionRows.Close()

			response = append(response, models.PVZWithReceptionsResponse{
				PVZ:        pvz,
				Receptions: receptions,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	}
}
