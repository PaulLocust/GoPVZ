package repo

import (
    "context"
    "GoPVZ/internal/auth/entity"
    "github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
    db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) UserRepository {
    return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *entity.User) error {
    _, err := r.db.Exec(ctx,
        `INSERT INTO users (id, email, password_hash, role) VALUES ($1,$2,$3,$4)`,
        user.ID, user.Email, user.PasswordHash, user.Role,
    )
    return err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
    var u entity.User
    err := r.db.QueryRow(ctx,
        `SELECT id, email, password_hash, role FROM users WHERE email=$1`, email,
    ).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
    if err != nil {
        return nil, err
    }
    return &u, nil
}