package validation

import (
	"GoPVZ/internal/dto"
	"GoPVZ/pkg/pkgValidator"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type PVZValidator struct {
	Payload dto.PostPvzJSONRequestBody
}

func NewPVZValidator(payload dto.PostPvzJSONRequestBody) *PVZValidator {
	return &PVZValidator{Payload: payload}
}

func (v *PVZValidator) Validate() error {
	if v.Payload.City != dto.PVZRequestCityMoscow && v.Payload.City != dto.PVZRequestCitySaintPetersburg && v.Payload.City != dto.PVZRequestCityKazan {
		return pkgValidator.ErrInvalidCity
	}

	return nil
}

type ReceptionsValidator struct {
	Payload dto.PostReceptionsJSONBody
}

func NewReceptionsValidator(payload dto.PostReceptionsJSONBody) *ReceptionsValidator {
	return &ReceptionsValidator{Payload: payload}
}

func (v *ReceptionsValidator) Validate() error {

	if uuid.UUID(v.Payload.PvzId).String() == "" {
		return pkgValidator.ErrInvalidPVZID
	}

	// Проверка что ID является валидным UUID
	if _, err := uuid.Parse(uuid.UUID(v.Payload.PvzId).String()); err != nil {
		return pkgValidator.ErrInvalidPVZID
	}

	return nil
}

type ProductsValidator struct {
	Payload dto.PostProductsJSONBody
}

func NewProductsValidator(payload dto.PostProductsJSONBody) *ProductsValidator {
	return &ProductsValidator{Payload: payload}
}

func (v *ProductsValidator) Validate() error {

	if uuid.UUID(v.Payload.PvzId).String() == "" {
		return pkgValidator.ErrInvalidPVZID
	}

	// Проверка что ID является валидным UUID
	if _, err := uuid.Parse(uuid.UUID(v.Payload.PvzId).String()); err != nil {
		return pkgValidator.ErrInvalidPVZID
	}

	if v.Payload.Type != dto.PostProductsJSONBodyTypeClothes && v.Payload.Type != dto.PostProductsJSONBodyTypeElectronics && v.Payload.Type != dto.PostProductsJSONBodyTypeShoes {
		return pkgValidator.ErrInvalidProductType
	}

	return nil
}

type DeleteLastProductValidator struct {
	PVZID string
}

func NewDeleteLastProductValidator(pvzId string) *DeleteLastProductValidator {
	return &DeleteLastProductValidator{PVZID: pvzId}
}

func (v *DeleteLastProductValidator) Validate() error {
	if v.PVZID == "" {
		return pkgValidator.ErrInvalidPVZID
	}

	// Проверка что ID является валидным UUID
	if _, err := uuid.Parse(v.PVZID); err != nil {
		return pkgValidator.ErrInvalidPVZID
	}

	return nil
}

type CloseReceptionValidator struct {
	PVZID string
}

func NewCloseReceptionValidator(pvzId string) *CloseReceptionValidator {
	return &CloseReceptionValidator{PVZID: pvzId}
}

func (v *CloseReceptionValidator) Validate() error {
	if v.PVZID == "" {
		return pkgValidator.ErrInvalidPVZID
	}

	// Проверка что ID является валидным UUID
	if _, err := uuid.Parse(v.PVZID); err != nil {
		return pkgValidator.ErrInvalidPVZID
	}

	return nil
}

type PVZsFilterValidator struct {
	StartDate string
	EndDate   string
	PageStr   string
	LimitStr  string
}

func NewPVZsFilterValidator(startDate, endDate, pageStr, limitStr string) *PVZsFilterValidator {
	return &PVZsFilterValidator{
		StartDate: startDate,
		EndDate:   endDate,
		PageStr:   pageStr,
		LimitStr:  limitStr,
	}
}

func (v *PVZsFilterValidator) Validate() error {
	// Валидация page
	if _, err := strconv.Atoi(v.PageStr); err != nil {
		return pkgValidator.ErrInvalidPage
	}
	page, _ := strconv.Atoi(v.PageStr) 
	if page < 1 {
		return pkgValidator.ErrInvalidPage
	}

	// Валидация limit
	if _, err := strconv.Atoi(v.LimitStr); err != nil {
		return pkgValidator.ErrInvalidLimit
	}
	limit, _ := strconv.Atoi(v.LimitStr) 
	if limit < 1 {
		return pkgValidator.ErrInvalidLimit
	}
	if limit > 30 { 
		return pkgValidator.ErrLimitTooHigh
	}

	// Валидация дат
	if v.StartDate != "" {
		if _, err := time.Parse(time.RFC3339, v.StartDate); err != nil {
			return pkgValidator.ErrInvalidDateFormat
		}
	}

	if v.EndDate != "" {
		if _, err := time.Parse(time.RFC3339, v.EndDate); err != nil {
			return pkgValidator.ErrInvalidDateFormat
		}
	}

	// Дополнительная проверка, что endDate не раньше startDate
	if v.StartDate != "" && v.EndDate != "" {
		startTime, _ := time.Parse(time.RFC3339, v.StartDate)
		endTime, _ := time.Parse(time.RFC3339, v.EndDate)
		if endTime.Before(startTime) {
			return pkgValidator.ErrInvalidDateRange
		}
	}

	return nil
}