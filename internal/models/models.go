package models

import "time"

type User struct {
	ID       string `json:"id" db:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Email    string `json:"email" db:"email" example:"user@example.com`
	Password string `json:"-" db:"password_hash" example:"strongpassword123"`
	Role     string `json:"role" db:"role" example:"employee" validate:"required,oneof=employee moderator"`
}

type PVZ struct {
	ID               string    `json:"id" db:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	RegistrationDate time.Time `json:"registrationDate" db:"registration_date" example:"2025-07-17T12:15:49.386Z"`
	City             string    `json:"city" db:"city" example:"Москва" validate:"required,oneof=Москва Санкт-Петербург Казань"`
}

type Reception struct {
	ID       string    `json:"id" db:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	PvzID    string    `json:"pvzId" db:"pvz_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	DateTime time.Time `json:"dateTime" db:"date_time" example:"2025-07-17T12:15:49.386Z"` // Дата и время проведения приёмки
	Status   string    `json:"status" db:"status" example:"in_progress"`
}

type Product struct {
	ID          string    `json:"id" db:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	ReceptionID string    `json:"receptionId" db:"reception_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	DateTime    time.Time `json:"dateTime" db:"date_time" example:"2025-07-17T12:15:49.386Z"` // Дата и время приёма товара
	Type        string    `json:"type" db:"type" example:"in_progress" validate:"required,oneof=электроника одежда обувь"`
}

type Token struct {
	TokenValue string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type ReceptionWithProducts struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}

type PVZWithReceptionsResponse struct {
	PVZ        PVZ                     `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}
