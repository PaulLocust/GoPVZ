package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"GoPVZ/internal/auth/entity"
	"GoPVZ/internal/auth/repo"
	"GoPVZ/internal/auth/usecase"
	"GoPVZ/internal/dto"
	"GoPVZ/pkg/pkgPostgres"
	"GoPVZ/pkg/pkgValidator"

	"github.com/gin-gonic/gin"
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
				WithStartupTimeout(30 * time.Second),
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

func setupTestHandler(t *testing.T) (*AuthHandler, func()) {
	pg, err := pkgPostgres.New(testConnStr)
	require.NoError(t, err)

	_, err = pg.Pool.Exec(context.Background(), "TRUNCATE TABLE users CASCADE")
	require.NoError(t, err)

	userRepo := repo.NewUserRepo(pg.Pool)
	jwtManager := usecase.NewJwtManager("test-secret", 3600)
	uc := usecase.NewAuthUseCase(userRepo, jwtManager)
	handler := NewAuthHandler(uc)

	return handler, func() {
		pg.Close()
	}
}

func TestRegisterHandler(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	router := gin.Default()
	router.POST("/register", handler.Register)

	tests := []struct {
		name         string
		payload      interface{}
		wantStatus   int
		wantErrorMsg string
	}{
		{
			name: "successful registration",
			payload: dto.PostRegisterJSONBody{
				Email:    "test@example.com",
				Password: "password123",
				Role:     dto.Employee,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid email format",
			payload: dto.PostRegisterJSONBody{
				Email:    "invalid-email",
				Password: "password123",
				Role:     dto.Employee,
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrInvalidEmail.Error(),
		},
		{
			name: "weak password",
			payload: dto.PostRegisterJSONBody{
				Email:    "test@example.com",
				Password: "123",
				Role:     dto.Employee,
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrPasswordTooWeak.Error(),
		},
		{
			name: "duplicate email",
			payload: dto.PostRegisterJSONBody{
				Email:    "duplicate@example.com",
				Password: "password123",
				Role:     dto.Employee,
			},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "user already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "duplicate email" {
				_, err := handler.uc.Register(context.Background(), "duplicate@example.com", "password123", string(entity.RoleEmployee))
				require.NoError(t, err)
			}

			bodyBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErrorMsg != "" {
				var errResp dto.Error
				err = json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
				require.Contains(t, errResp.Message, tt.wantErrorMsg)
			} else {
				var resp dto.User
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.NotNil(t, resp.Id)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	router := gin.Default()
	router.POST("/login", handler.Login)

	email := "test@example.com"
	password := "password123"
	_, err := handler.uc.Register(context.Background(), email, password, string(entity.RoleEmployee))
	require.NoError(t, err)

	tests := []struct {
		name         string
		payload      interface{}
		wantStatus   int
		wantErrorMsg string
	}{
		{
			name: "successful login",
			payload: dto.PostLoginJSONBody{
				Email:    "test@example.com",
				Password: password,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid password",
			payload: dto.PostLoginJSONBody{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "invalid credentials",
		},
		{
			name: "invalid email format",
			payload: dto.PostLoginJSONBody{
				Email:    "invalid-email",
				Password: "password123",
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrInvalidEmail.Error(),
		},
		{
			name: "user not found",
			payload: dto.PostLoginJSONBody{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "no rows in result set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErrorMsg != "" {
				var errResp dto.Error
				err = json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
				require.Contains(t, errResp.Message, tt.wantErrorMsg)
			} else {
				var resp dto.TokenResponse
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.NotEmpty(t, resp.Token)
			}
		})
	}
}

func TestDummyLoginHandler(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	router := gin.Default()
	router.POST("/dummyLogin", handler.DummyLogin)

	tests := []struct {
		name         string
		payload      interface{}
		wantStatus   int
		wantErrorMsg string
	}{
		{
			name: "successful dummy login - employee",
			payload: dto.PostDummyLoginJSONBody{
				Role: dto.PostDummyLoginJSONBodyRoleEmployee,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "successful dummy login - moderator",
			payload: dto.PostDummyLoginJSONBody{
				Role: dto.PostDummyLoginJSONBodyRoleModerator,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid role",
			payload: dto.PostDummyLoginJSONBody{
				Role: "invalid",
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrInvalidRole.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, tt.wantStatus, w.Code)

			if tt.wantErrorMsg != "" {
				var errResp dto.Error
				err = json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
				require.Contains(t, errResp.Message, tt.wantErrorMsg)
			} else {
				var resp dto.TokenResponse
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.NotEmpty(t, resp.Token)
			}
		})
	}
}
