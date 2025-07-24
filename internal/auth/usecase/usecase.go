package usecase

import (
	"GoPVZ/internal/auth/entity"
	"GoPVZ/internal/auth/repo"
	"GoPVZ/pkg/pkgValidator"
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	repo       repo.UserRepository
	jwtManager *JwtManager
}

func (uc *AuthUseCase) GetJwtManager() *JwtManager {
	return uc.jwtManager
}

func NewAuthUseCase(r repo.UserRepository, jm *JwtManager) *AuthUseCase {
	return &AuthUseCase{repo: r, jwtManager: jm}
}

func (uc *AuthUseCase) Register(ctx context.Context, email, password string, role entity.Role) (*entity.User, error) {
	if existing, _ := uc.repo.GetByEmail(ctx, email); existing != nil {
		return nil, pkgValidator.ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", pkgValidator.ErrInvalidCredentials
	}
	return uc.jwtManager.GenerateToken(user)
}

func (uc *AuthUseCase) DummyLogin(ctx context.Context, role entity.Role) (string, error) {
	if role != entity.RoleEmployee && role != entity.RoleModerator {
		return "", pkgValidator.ErrInvalidRole
	}

	user := &entity.User{
		ID:    uuid.New(),
		Email: "dummy@pvz",
		Role:  role,
	}
	return uc.jwtManager.GenerateToken(user)
}
