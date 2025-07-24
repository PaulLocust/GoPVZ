package repo

import (
	"context"
	"os"
	"testing"
	"time"

	"GoPVZ/internal/auth/entity"
	"GoPVZ/pkg/pkgPostgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testConnStr string
	pgContainer *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	pgContainer, err = postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}

	testConnStr, err = pgContainer.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	pg, err := pkgPostgres.New(testConnStr)
	if err != nil {
		panic(err)
	}
	
	_, err = pg.Pool.Exec(context.Background(), `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL CHECK (role IN ('employee', 'moderator'))
		);
	`)
	if err != nil {
		panic(err)
	}
	pg.Close()

	code := m.Run()

	if err := pgContainer.Terminate(ctx); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func setupTestRepo(t *testing.T) (UserRepository, func()) {
	pg, err := pkgPostgres.New(testConnStr)
	require.NoError(t, err)

	_, err = pg.Pool.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	require.NoError(t, err)

	repo := NewUserRepo(pg.Pool)

	return repo, func() {
		pg.Close()
	}
}

func TestUserRepository_Create(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	tests := []struct {
		name        string
		user        *entity.User
		wantError   bool
		errorString string
	}{
		{
			name: "successful creation",
			user: &entity.User{
				ID:           uuid.New(),
				Email:        "test1@example.com",
				PasswordHash: "hashed_password1",
				Role:         entity.RoleEmployee,
			},
			wantError: false,
		},
		{
			name: "duplicate email",
			user: &entity.User{
				ID:           uuid.New(),
				Email:        "test1@example.com", // Используем тот же email
				PasswordHash: "hashed_password2",
				Role:         entity.RoleModerator,
			},
			wantError:   true,
			errorString: "duplicate key value violates unique constraint",
		},
	}

	// Создаем первого пользователя
	err := repo.Create(context.Background(), tests[0].user)
	require.NoError(t, err)

	// Пытаемся создать пользователя с тем же email
	err = repo.Create(context.Background(), tests[1].user)
	if tests[1].wantError {
		require.Error(t, err)
		require.Contains(t, err.Error(), tests[1].errorString)
	} else {
		require.NoError(t, err)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	// Подготовка тестовых данных
	testUser := &entity.User{
		ID:           uuid.New(),
		Email:        "getbyemail@example.com",
		PasswordHash: "hashed_password",
		Role:         entity.RoleModerator,
	}

	// Создаем пользователя для теста
	err := repo.Create(context.Background(), testUser)
	require.NoError(t, err)

	tests := []struct {
		name        string
		email       string
		wantUser    *entity.User
		wantError   bool
		errorString string
	}{
		{
			name:      "existing user",
			email:     "getbyemail@example.com",
			wantUser:  testUser,
			wantError: false,
		},
		{
			name:        "non-existing user",
			email:       "nonexistent@example.com",
			wantUser:    nil,
			wantError:   true,
			errorString: "no rows in result set",
		},
		{
			name:        "empty email",
			email:       "",
			wantUser:    nil,
			wantError:   true,
			errorString: "no rows in result set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByEmail(context.Background(), tt.email)

			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, user)
				if tt.errorString != "" {
					require.Contains(t, err.Error(), tt.errorString)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantUser.ID, user.ID)
				require.Equal(t, tt.wantUser.Email, user.Email)
				require.Equal(t, tt.wantUser.PasswordHash, user.PasswordHash)
				require.Equal(t, tt.wantUser.Role, user.Role)
			}
		})
	}
}

