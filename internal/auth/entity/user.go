package entity

import "github.com/google/uuid"

type Role string

const (
	RoleEmployee  Role = "employee"
	RoleModerator Role = "moderator"
)

type User struct {
	ID           uuid.UUID `json:"id"    db:"id"            example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	Email        string    `json:"email" db:"email"         example:"user@example.com"`
	PasswordHash string    `json:"-"     db:"password_hash" example:"strongpassword123"`
	Role         Role      `json:"role"  db:"role"          example:"employee"`
}