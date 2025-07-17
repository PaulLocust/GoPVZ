package handlers

import (
	"GoPVZ/internal/lib/sl"
	"GoPVZ/internal/transport/rest/helpers"
	"database/sql"
	"log/slog"
	"net/http"
	"strings"
)

// DeleteLastProductHandler godoc
// @Summary Удаление последнего добавленного товара из текущей приемки (LIFO, только для сотрудников ПВЗ)
// @Description Удаляет самый последний добавленный товар (LIFO) (по дате) из открытой приемки указанного ПВЗ
// @Tags Protected
// @Param pvzId path string true "ID пункта выдачи заказов (ПВЗ)"
// @Success 200 {string} string "Товар успешно удален"
// @Failure 400 {object} helpers.ErrorResponse "Некорректный путь запроса"
// @Failure 403 {object} helpers.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} helpers.ErrorResponse "Не найдено: либо нет открытой приемки, либо нет товаров для удаления"
// @Failure 405 {object} helpers.ErrorResponse "Метод не разрешен"
// @Failure 500 {object} helpers.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /pvz/{pvzId}/delete_last_product [post]
func DeleteLastProductHandler(log *slog.Logger, DBConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			helpers.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Извлекаем pvzId из пути
		urlPath := r.URL.Path
		
		// Путь должен быть вида /pvz/{pvzId}/delete_last_product
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

		// Удаляем последний товар (LIFO)
        _, err = DBConn.Exec(`
            WITH last_product AS (
                SELECT p.id 
                FROM products p
                JOIN receptions r ON p.reception_id = r.id
                WHERE r.pvz_id = $1 AND r.status = $2
                ORDER BY p.date_time DESC
                LIMIT 1
            )
            DELETE FROM products
            WHERE id IN (SELECT id FROM last_product)
        `, pvzId, in_progress)

		// Обработка результата
		switch {
		case err == sql.ErrNoRows:
			helpers.WriteJSONError(w, "No products to delete", http.StatusNotFound)
			return
		case err != nil:
			log.Error("Failed to delete product", sl.Err(err))
			helpers.WriteJSONError(w, "Failed to delete product", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
