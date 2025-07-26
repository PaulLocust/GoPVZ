package repo

import (
    "context"
    "GoPVZ/internal/auth/entity"
)

type UserRepository interface {
    Create(ctx context.Context, user *entity.User) error
    GetByEmail(ctx context.Context, email string) (*entity.User, error)
}