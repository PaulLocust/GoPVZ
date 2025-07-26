package http

import (
	"GoPVZ/internal/dto"
	"GoPVZ/internal/pvz/usecase"
	"GoPVZ/internal/pvz/validation"
	"GoPVZ/pkg/pkgValidator"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PVZHandler struct {
	uc *usecase.PVZUseCase
}

func NewPVZHandler(uc *usecase.PVZUseCase) *PVZHandler {
	return &PVZHandler{uc: uc}
}

// CreatePVZ godoc
// @Summary Создание ПВЗ (только для модераторов)
// @Description Добавляет новый пункт выдачи заказов в систему
// @Tags Domain pvz
// @Accept json
// @Produce json
// @Param input body dto.PostPvzJSONRequestBody true "Данные для создания ПВЗ"
// @Success 201 {object} dto.PVZ "ПВЗ успешно создан"
// @Failure 400 {object} dto.Error "Неверный формат запроса или ошибка валидации"
// @Failure 401 {object} dto.Error "Ошибка авторизации"
// @Failure 403 {object} dto.Error "Доступ запрещен"
// @Failure 500 {object} dto.Error "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /pvz [post]
func (h *PVZHandler) CreatePVZ(c *gin.Context) {
	var req dto.PostPvzJSONRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: pkgValidator.ErrInvalidInput.Error()})
		return
	}

	validator := validation.NewPVZValidator(req)
	if err := validator.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		return
	}

	pvz, err := h.uc.CreatePVZ(c, string(req.City))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.PVZ{Id: pvz.ID, City: dto.PVZCity(pvz.City), RegistrationDate: pvz.RegistrationDate})
}

// CreateReception godoc
// @Summary Создание новой приемки товаров (только для сотрудников ПВЗ)
// @Description Создает новую запись о приеме в ПВЗ (пункте выдачи заказов) с указанным PVZ ID
// @Tags Domain pvz
// @Accept json
// @Produce json
// @Param input body dto.PostReceptionsJSONBody true "Данные для создания записи приема"
// @Success 201 {object} dto.Reception "Успешно созданная запись приема"
// @Failure 400 {object} dto.Error "Невалидные входные данные"
// @Failure 401 {object} dto.Error "Ошибка авторизации"
// @Failure 403 {object} dto.Error "Доступ запрещен"
// @Failure 500 {object} dto.Error "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /receptions [post]
func (h *PVZHandler) CreateReception(c *gin.Context) {
	var req dto.PostReceptionsJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: pkgValidator.ErrInvalidPVZID.Error()})
		return
	}

	validator := validation.NewReceptionsValidator(req)
	if err := validator.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		return
	}

	reception, err := h.uc.CreateReception(c, uuid.UUID(req.PvzId).String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.Reception{Id: reception.ID, PvzId: reception.PvzID, DateTime: reception.DateTime, Status: dto.ReceptionStatus(reception.Status)})
}

// CreateProduct godoc
// @Summary Добавление товара в текущую приемку (только для сотрудников ПВЗ)
// @Description Создает запись о новом продукте в системе, привязывая его к указанному ПВЗ и приему
// @Tags Domain pvz
// @Accept json
// @Produce json
// @Param input body dto.PostProductsJSONBody true "Данные для создания продукта"
// @Success 201 {object} dto.Product "Успешно созданный продукт"
// @Failure 400 {object} dto.Error "Невалидные входные данные"
// @Failure 401 {object} dto.Error "Ошибка авторизации"
// @Failure 403 {object} dto.Error "Доступ запрещен"
// @Failure 500 {object} dto.Error "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /products [post]
func (h *PVZHandler) CreateProduct(c *gin.Context) {
	var req dto.PostProductsJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: pkgValidator.ErrInvalidPVZID.Error()})
		return
	}

	validator := validation.NewProductsValidator(req)
	if err := validator.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		return
	}

	product, err := h.uc.CreateProduct(c, string(req.Type), uuid.UUID(req.PvzId).String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.Product{Id: product.ID, ReceptionId: product.ReceptionID, DateTime: product.DateTime, Type: dto.ProductType(product.Type)})
}

// DeleteLastProduct godoc
// @Summary Удаление последнего добавленного товара из текущей приемки (LIFO, только для сотрудников ПВЗ)
// @Description Удаляет последний добавленный товар по принципу LIFO из активной приемки
// @Tags Domain pvz
// @Accept json
// @Produce json
// @Param pvzId path string true "pvzId"
// @Success 200 "Товар успешно удален"
// @Failure 400 {object} dto.Error "Нет активной приемки или другие ошибки валидации"
// @Failure 401 {object} dto.Error "Ошибка авторизации"
// @Failure 403 {object} dto.Error "Доступ запрещен"
// @Failure 500 {object} dto.Error "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /pvz/{pvzId}/delete_last_product [post]
func (h *PVZHandler) DeleteLastProduct(c *gin.Context) {
    pvzId := c.Param("pvzId")
    
    // Валидация входных данных
    validator := validation.NewDeleteLastProductValidator(pvzId) // Используем тот же валидатор
    if err := validator.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
        return
    }

    if err := h.uc.DeleteLastProduct(c, pvzId); err != nil {
        if errors.Is(err, pkgValidator.ErrNoActiveReception) {
            c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
        } else {
            c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
        }
        return
    }

    c.Status(http.StatusOK)
}


// CloseReception godoc
// @Summary Закрытие последней открытой приемки товаров в рамках ПВЗ (только для сотрудников ПВЗ)
// @Description Закрывает активную приёмку товаров для указанного ПВЗ
// @Tags Domain pvz
// @Accept json
// @Produce json
// @Param pvzId path string true "pvzId"
// @Success 200 {object} dto.Reception "Приёмка успешно закрыта"
// @Failure 400 {object} dto.Error "Нет активной приемки или другие ошибки валидации"
// @Failure 401 {object} dto.Error "Ошибка авторизации"
// @Failure 403 {object} dto.Error "Доступ запрещен"
// @Failure 500 {object} dto.Error "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /pvz/{pvzId}/close_last_reception [post]
func (h *PVZHandler) CloseReception(c *gin.Context) {
    pvzId := c.Param("pvzId")
    
    // Валидация входных данных
    validator := validation.NewCloseReceptionValidator(pvzId)
    if err := validator.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
        return
    }

    reception, err := h.uc.CloseReception(c, pvzId)
    if err != nil {
        if errors.Is(err, pkgValidator.ErrNoActiveReception) {
            c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
        } else {
            c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, dto.Reception{
        Id:      reception.ID,
        PvzId:   reception.PvzID,
        DateTime: reception.DateTime,
        Status:  dto.ReceptionStatus(reception.Status),
    })
}


// GetPVZsWithReceptions godoc
// @Summary Получение списка ПВЗ с фильтрацией по дате приемки и пагинацией (только для сотрудников ПВЗ или модераторов)
// @Description Возвращает список ПВЗ с информацией о приёмках и товарах с возможностью фильтрации по дате
// @Tags Domain pvz
// @Accept json
// @Produce json
// @Param startDate query string false "Начальная дата диапазона (RFC3339)"
// @Param endDate query string false "Конечная дата диапазона (RFC3339)"
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Success 200 {array} dto.PVZWithReceptions "Список ПВЗ с приёмками и товарами"
// @Failure 400 {object} dto.Error "Неверные параметры запроса"
// @Failure 401 {object} dto.Error "Ошибка авторизации"
// @Failure 403 {object} dto.Error "Доступ запрещен"
// @Failure 500 {object} dto.Error "Внутренняя ошибка сервера"
// @Security BearerAuth
// @Router /pvz [get]
func (h *PVZHandler) GetPVZsWithReceptions(c *gin.Context) {
    // Получаем параметры запроса
    startDate := c.Query("startDate")
    endDate := c.Query("endDate")
    pageStr := c.DefaultQuery("page", "1")
    limitStr := c.DefaultQuery("limit", "10")

    // Валидация параметров
    validator := validation.NewPVZsFilterValidator(startDate, endDate, pageStr, limitStr)
    if err := validator.Validate(); err != nil {
        c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
        return
    }

    // Парсинг параметров
    page, _ := strconv.Atoi(pageStr)
    limit, _ := strconv.Atoi(limitStr)

    // Парсинг дат
    var startTime, endTime *time.Time
    if startDate != "" {
        st, _ := time.Parse(time.RFC3339, startDate)
        startTime = &st
    }
    if endDate != "" {
        et, _ := time.Parse(time.RFC3339, endDate)
        endTime = &et
    }

    // Получаем данные из usecase
    entities, err := h.uc.GetPVZsWithReceptions(c, startTime, endTime, page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
        return
    }

    // Преобразование entity в DTO
    response := make([]dto.PVZWithReceptions, len(entities))
    for i, pvzEntity := range entities {
        // Преобразование ReceptionWithProducts
        receptionsDTO := make([]dto.ReceptionWithProducts, len(pvzEntity.Receptions))
        for j, receptionEntity := range pvzEntity.Receptions {
            // Преобразование Products
            productsDTO := make([]dto.Product, len(receptionEntity.Products))
            for k, productEntity := range receptionEntity.Products {
                productsDTO[k] = dto.Product{
                    Id:          productEntity.ID,
                    ReceptionId: productEntity.ReceptionID,
                    DateTime:    productEntity.DateTime,
                    Type:        dto.ProductType(productEntity.Type),
                }
            }

            receptionsDTO[j] = dto.ReceptionWithProducts{
                Reception: dto.Reception{
                    Id:       receptionEntity.Reception.ID,
                    PvzId:    receptionEntity.Reception.PvzID,
                    DateTime: receptionEntity.Reception.DateTime,
                    Status:   dto.ReceptionStatus(receptionEntity.Reception.Status),
                },
                Products: productsDTO,
            }
        }

        response[i] = dto.PVZWithReceptions{
            Pvz: dto.PVZ{
                Id:               pvzEntity.PVZ.ID,
                RegistrationDate: pvzEntity.PVZ.RegistrationDate,
                City:             dto.PVZCity(pvzEntity.PVZ.City),
            },
            Receptions: receptionsDTO,
        }
    }

    c.JSON(http.StatusOK, response)
}