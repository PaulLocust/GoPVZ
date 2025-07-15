package models

import "time"

type User struct {
	ID        string     `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password_hash"`
	Role      string     `json:"role" db:"role"` // employee или moderator
}

type PVZ struct {
	ID               string     `json:"id" db:"id"`
	RegistrationDate time.Time  `json:"registrationDate" db:"registration_date"`
	City             string     `json:"city" db:"city"` // Москва, Санкт-Петербург, Казань
}

type Reception struct {
	ID          string     `json:"id" db:"id"`
	PvzID       string     `json:"pvzId" db:"pvz_id"`
	DateTime    time.Time  `json:"dateTime" db:"date_time"` // Дата и время проведения приёмки
	Status      string     `json:"status" db:"status"`
}

type Product struct {
	ID          string     `json:"id" db:"id"`
	ReceptionID string     `json:"receptionId" db:"reception_id"`
	DateTime    time.Time  `json:"dateTime" db:"date_time"` // Дата и время приёма товара
	Type        string     `json:"type" db:"type"`          // электроника, одежда, обувь
}
