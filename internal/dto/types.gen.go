// Package dto provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.3 DO NOT EDIT.
package dto

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// Defines values for PVZCity.
const (
	PVZCityKazan           PVZCity = "Kazan"
	PVZCityMoscow          PVZCity = "Moscow"
	PVZCitySaintPetersburg PVZCity = "Saint Petersburg"
)

// Defines values for PVZRequestCity.
const (
	PVZRequestCityKazan           PVZRequestCity = "Kazan"
	PVZRequestCityMoscow          PVZRequestCity = "Moscow"
	PVZRequestCitySaintPetersburg PVZRequestCity = "Saint Petersburg"
)

// Defines values for ProductType.
const (
	ProductTypeClothes     ProductType = "clothes"
	ProductTypeElectronics ProductType = "electronics"
	ProductTypeShoes       ProductType = "shoes"
)

// Defines values for ReceptionStatus.
const (
	Close      ReceptionStatus = "close"
	InProgress ReceptionStatus = "in_progress"
)

// Defines values for UserRole.
const (
	UserRoleEmployee  UserRole = "employee"
	UserRoleModerator UserRole = "moderator"
)

// Defines values for PostDummyLoginJSONBodyRole.
const (
	PostDummyLoginJSONBodyRoleEmployee  PostDummyLoginJSONBodyRole = "employee"
	PostDummyLoginJSONBodyRoleModerator PostDummyLoginJSONBodyRole = "moderator"
)

// Defines values for PostProductsJSONBodyType.
const (
	PostProductsJSONBodyTypeClothes     PostProductsJSONBodyType = "clothes"
	PostProductsJSONBodyTypeElectronics PostProductsJSONBodyType = "electronics"
	PostProductsJSONBodyTypeShoes       PostProductsJSONBodyType = "shoes"
)

// Defines values for PostRegisterJSONBodyRole.
const (
	Employee  PostRegisterJSONBodyRole = "employee"
	Moderator PostRegisterJSONBodyRole = "moderator"
)

// Error defines model for Error.
type Error struct {
	Message string `json:"message"`
}

// PVZ defines model for PVZ.
type PVZ struct {
	City             PVZCity            `json:"city"`
	Id               openapi_types.UUID `json:"id"`
	RegistrationDate time.Time          `json:"registrationDate"`
}

// PVZCity defines model for PVZ.City.
type PVZCity string

// PVZListResponse defines model for PVZListResponse.
type PVZListResponse = []PVZWithReceptions

// PVZWithReceptions defines model for PVZWithReceptions.
type PVZWithReceptions struct {
	Pvz        PVZ                     `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

// PVZRequest defines model for PVZ_Request.
type PVZRequest struct {
	City PVZRequestCity `json:"city"`
}

// PVZRequestCity defines model for PVZRequest.City.
type PVZRequestCity string

// Product defines model for Product.
type Product struct {
	DateTime    time.Time          `json:"dateTime"`
	Id          openapi_types.UUID `json:"id"`
	ReceptionId openapi_types.UUID `json:"receptionId"`
	Type        ProductType        `json:"type"`
}

// ProductType defines model for Product.Type.
type ProductType string

// Reception defines model for Reception.
type Reception struct {
	DateTime time.Time          `json:"dateTime"`
	Id       openapi_types.UUID `json:"id"`
	PvzId    openapi_types.UUID `json:"pvzId"`
	Status   ReceptionStatus    `json:"status"`
}

// ReceptionStatus defines model for Reception.Status.
type ReceptionStatus string

// ReceptionWithProducts defines model for ReceptionWithProducts.
type ReceptionWithProducts struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}

// TokenResponse defines model for TokenResponse.
type TokenResponse struct {
	Token string `json:"token"`
}

// User defines model for User.
type User struct {
	Email string              `json:"email"`
	Id    *openapi_types.UUID `json:"id,omitempty"`
	Role  UserRole            `json:"role"`
}

// UserRole defines model for User.Role.
type UserRole string

// PostDummyLoginJSONBody defines parameters for PostDummyLogin.
type PostDummyLoginJSONBody struct {
	Role PostDummyLoginJSONBodyRole `json:"role"`
}

// PostDummyLoginJSONBodyRole defines parameters for PostDummyLogin.
type PostDummyLoginJSONBodyRole string

// PostLoginJSONBody defines parameters for PostLogin.
type PostLoginJSONBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PostProductsJSONBody defines parameters for PostProducts.
type PostProductsJSONBody struct {
	PvzId openapi_types.UUID       `json:"pvzId"`
	Type  PostProductsJSONBodyType `json:"type"`
}

// PostProductsJSONBodyType defines parameters for PostProducts.
type PostProductsJSONBodyType string

// GetPvzParams defines parameters for GetPvz.
type GetPvzParams struct {
	// StartDate Начальная дата диапазона
	StartDate *time.Time `form:"startDate,omitempty" json:"startDate,omitempty"`

	// EndDate Конечная дата диапазона
	EndDate *time.Time `form:"endDate,omitempty" json:"endDate,omitempty"`

	// Page Номер страницы
	Page *int `form:"page,omitempty" json:"page,omitempty"`

	// Limit Количество элементов на странице
	Limit *int `form:"limit,omitempty" json:"limit,omitempty"`
}

// PostReceptionsJSONBody defines parameters for PostReceptions.
type PostReceptionsJSONBody struct {
	PvzId openapi_types.UUID `json:"pvzId"`
}

// PostRegisterJSONBody defines parameters for PostRegister.
type PostRegisterJSONBody struct {
	Email    string                   `json:"email"`
	Password string                   `json:"password"`
	Role     PostRegisterJSONBodyRole `json:"role"`
}

// PostRegisterJSONBodyRole defines parameters for PostRegister.
type PostRegisterJSONBodyRole string

// PostDummyLoginJSONRequestBody defines body for PostDummyLogin for application/json ContentType.
type PostDummyLoginJSONRequestBody PostDummyLoginJSONBody

// PostLoginJSONRequestBody defines body for PostLogin for application/json ContentType.
type PostLoginJSONRequestBody PostLoginJSONBody

// PostProductsJSONRequestBody defines body for PostProducts for application/json ContentType.
type PostProductsJSONRequestBody PostProductsJSONBody

// PostPvzJSONRequestBody defines body for PostPvz for application/json ContentType.
type PostPvzJSONRequestBody = PVZRequest

// PostReceptionsJSONRequestBody defines body for PostReceptions for application/json ContentType.
type PostReceptionsJSONRequestBody PostReceptionsJSONBody

// PostRegisterJSONRequestBody defines body for PostRegister for application/json ContentType.
type PostRegisterJSONRequestBody PostRegisterJSONBody
