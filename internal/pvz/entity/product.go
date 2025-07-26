package entity

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeElectronics Type = "electronics"
	TypeClothes     Type = "clothes"
	TypeShoes       Type = "shoes"
)

type Product struct {
	ID          uuid.UUID `json:"id"          db:"id"           example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	ReceptionID uuid.UUID `json:"receptionId" db:"reception_id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	DateTime    time.Time `json:"dateTime"    db:"date_time"    example:"2025-07-17T12:15:49.386Z"`
	Type        Type      `json:"type"        db:"type"         example:"electronics"`
}
