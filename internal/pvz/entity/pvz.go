package entity

import (
	"time"

	"github.com/google/uuid"
)

type City string

const (
	CityMoscow          City = "Moscow"
	CitySaintPetersburg City = "Saint Petersburg"
	CityKazan           City = "Kazan"
)

type PVZ struct {
	ID               uuid.UUID `json:"id"               db:"id"                example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	RegistrationDate time.Time `json:"registrationDate" db:"registration_date" example:"2025-07-17T12:15:49.386Z"`
	City             City      `json:"city"             db:"city"              example:"Moscow"`
}

type PVZWithReceptions struct {
	PVZ        *PVZ                     `json:"pvz"`
	Receptions []*ReceptionWithProducts `json:"receptions"`
}

type ReceptionWithProducts struct {
	Reception *Reception `json:"reception"`
	Products  []*Product `json:"products"`
}
