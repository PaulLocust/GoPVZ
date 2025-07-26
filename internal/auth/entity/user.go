// user.go
package entity

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

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

var (
	ErrInvalidRole  = errors.New("invalid role")
	ErrInvalidEmail = errors.New("invalid email")
)


func (r Role) IsValid() bool {
	switch r {
	case RoleEmployee, RoleModerator:
		return true
	default:
		return false
	}
}


func (r Role) Validate() error {
	if !r.IsValid() {
		return ErrInvalidRole
	}
	return nil
}

// Validate проверяет все поля пользователя
func (u *User) Validate() error {
	if u.ID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}
	if u.Email == "" || !strings.Contains(string(u.Email), "@"){
		return ErrInvalidEmail
	}
	if u.PasswordHash == "" {
		return errors.New("password hash cannot be empty")
	}
	return u.Role.Validate()
}