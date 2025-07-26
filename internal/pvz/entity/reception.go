package entity

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusClose      Status = "close"
)

type Reception struct {
	ID       uuid.UUID `json:"id"       db:"id"        example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	PvzID    uuid.UUID `json:"pvzId"    db:"pvz_id"    example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	DateTime time.Time `json:"dateTime" db:"date_time" example:"2025-07-17T12:15:49.386Z"`
	Status   Status    `json:"status"   db:"status"    example:"in_progress"`
}
