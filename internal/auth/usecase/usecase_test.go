package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"GoPVZ/internal/auth/entity"
	"GoPVZ/internal/auth/validation"
	"GoPVZ/internal/dto"
	"GoPVZ/pkg/pkgValidator"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
}


func (m *MockUserRepo) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*entity.User), args.Error(1)
}

// Unit-тесты для usecase
func TestAuthUseCase_Register(t *testing.T) {
	tests := []struct {
		name      string
		payload   dto.PostRegisterJSONBody
		repoError error
		wantError bool
	}{
		{
			name: "success",
			payload: dto.PostRegisterJSONBody{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "employee",
			},
			repoError: nil,
			wantError: false,
		},
		{
			name: "user exists",
			payload: dto.PostRegisterJSONBody{
				Email:    "exists@example.com",
				Password: "password123",
				Role:     "moderator",
			},
			repoError: pkgValidator.ErrUserExists,
			wantError: true,
		},
		{
			name: "invalid email",
			payload: dto.PostRegisterJSONBody{
				Email:    "invalid-email",
				Password: "password123",
				Role:     "employee",
			},
			wantError: true,
		},
		{
			name: "weak password",
			payload: dto.PostRegisterJSONBody{
				Email:    "test@example.com",
				Password: "123",
				Role:     "employee",
			},
			wantError: true,
		},
		{
			name: "invalid role",
			payload: dto.PostRegisterJSONBody{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "invalid",
			},
			wantError: true,
		},
	}

	for _, tCase := range tests {
		t.Run(tCase.name, func(t *testing.T) {
			// Валидатор
			validator := validation.NewRegisterValidator(tCase.payload)
			if err := validator.Validate(); err != nil {
				if !tCase.wantError {
					t.Errorf("expected success, got validation error: %v", err)
				}
				return
			}

			// Моки
			mockRepo := new(MockUserRepo)
			jm := NewJwtManager("secret", 24*time.Hour)
			uc := NewAuthUseCase(mockRepo, jm)

			if tCase.repoError != nil {
				mockRepo.On("GetByEmail", mock.Anything, string(tCase.payload.Email)).Return(&entity.User{}, nil)
			} else {
				mockRepo.On("GetByEmail", mock.Anything, string(tCase.payload.Email)).Return((*entity.User)(nil), errors.New("not found"))
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(tCase.repoError)
			}

			role := entity.Role(tCase.payload.Role)
			user, err := uc.Register(context.Background(), string(tCase.payload.Email), tCase.payload.Password, role)

			if tCase.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, string(tCase.payload.Email), user.Email)
				assert.Equal(t, role, user.Role)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}


func TestAuthUseCase_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &entity.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Role:         entity.RoleEmployee,
	}

	tests := []struct {
		name      string
		payload   dto.PostLoginJSONBody
		setupUser *entity.User
		wantError bool
	}{
		{
			name: "success",
			payload: dto.PostLoginJSONBody{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupUser: testUser,
			wantError: false,
		},
		{
			name: "invalid password",
			payload: dto.PostLoginJSONBody{
				Email:    "test@example.com",
				Password: "wrongpass",
			},
			setupUser: testUser,
			wantError: true,
		},
		{
			name: "user not found",
			payload: dto.PostLoginJSONBody{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupUser: nil,
			wantError: true,
		},
		{
			name: "invalid email format",
			payload: dto.PostLoginJSONBody{
				Email:    "bad-email",
				Password: "password123",
			},
			wantError: true,
		},
		{
			name: "weak password",
			payload: dto.PostLoginJSONBody{
				Email:    "test@example.com",
				Password: "short",
			},
			wantError: true,
		},
	}

	for _, tCase := range tests {
		t.Run(tCase.name, func(t *testing.T) {
			validator := validation.NewLoginValidator(tCase.payload)
			if err := validator.Validate(); err != nil {
				if !tCase.wantError {
					t.Errorf("expected success, got validation error: %v", err)
				}
				return
			}

			mockRepo := new(MockUserRepo)
			jm := NewJwtManager("secret", 24*time.Hour)
			uc := NewAuthUseCase(mockRepo, jm)

			mockRepo.On("GetByEmail", mock.Anything, string(tCase.payload.Email)).Return(tCase.setupUser, func() error {
				if tCase.setupUser == nil {
					return errors.New("not found")
				}
				return nil
			}())

			token, err := uc.Login(context.Background(), string(tCase.payload.Email), tCase.payload.Password)

			if tCase.wantError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestAuthUseCase_DummyLogin(t *testing.T) {
    tests := []struct {
        name      string
        role      entity.Role
        wantError bool
		errorString string
    }{
        {
            name:      "success employee",
            role:      entity.RoleEmployee,
            wantError: false,
        },
        {
            name:      "success moderator",
            role:      entity.RoleModerator,
            wantError: false,
        },
        {
            name:        "invalid role",
            role:        "invalid_role",
            wantError:   true,
            errorString: pkgValidator.ErrInvalidRole.Error(),
        },
    }

    for _, tCase := range tests {
        t.Run(tCase.name, func(t *testing.T) {
            mockRepo := new(MockUserRepo)
            jm := NewJwtManager("secret", 24*time.Hour)
            uc := NewAuthUseCase(mockRepo, jm)

            token, err := uc.DummyLogin(context.Background(), tCase.role)

            if tCase.wantError {
                assert.Error(t, err)
                assert.Empty(t, token)
            } else {
                assert.NoError(t, err)
                assert.NotEmpty(t, token)
                
                // Дополнительная проверка токена
                claims, err := jm.VerifyToken(token)
                assert.NoError(t, err)
                assert.Equal(t, string(tCase.role), (*claims)["role"])
            }
        })
    }
}