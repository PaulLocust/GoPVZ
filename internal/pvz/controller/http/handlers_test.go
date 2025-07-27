package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"GoPVZ/internal/dto"
	"GoPVZ/internal/pvz/repo"
	"GoPVZ/internal/pvz/usecase"
	"GoPVZ/pkg/pkgPostgres"
	"GoPVZ/pkg/pkgValidator"

	"github.com/gin-gonic/gin"
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
		CREATE TABLE IF NOT EXISTS pvz (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			registration_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			city VARCHAR(50) NOT NULL CHECK (city IN ('Moscow', 'Saint Petersburg', 'Kazan'))
		);
		
		CREATE TABLE IF NOT EXISTS receptions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			pvz_id UUID NOT NULL REFERENCES pvz(id),
			date_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			status VARCHAR(20) NOT NULL CHECK (status IN ('in_progress', 'close'))
		);
		
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			reception_id UUID NOT NULL REFERENCES receptions(id),
			date_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			type VARCHAR(20) NOT NULL CHECK (type IN ('electronics', 'clothes', 'shoes'))
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

func setupTestPVZHandler(t *testing.T) (*PVZHandler, func()) {
	pg, err := pkgPostgres.New(testConnStr)
	require.NoError(t, err)

	_, err = pg.Pool.Exec(context.Background(), "TRUNCATE TABLE products, receptions, pvz CASCADE")
	require.NoError(t, err)

	pvzRepo := repo.NewPVZRepo(pg.Pool)
	uc := usecase.NewPVZUseCase(pvzRepo)
	handler := NewPVZHandler(uc)

	return handler, func() {
		pg.Close()
	}
}

func TestCreatePVZHandler(t *testing.T) {
	handler, cleanup := setupTestPVZHandler(t)
	defer cleanup()

	router := gin.Default()
	router.POST("/pvz", handler.CreatePVZ)

	tests := []struct {
		name         string
		payload      interface{}
		wantStatus   int
		wantErrorMsg string
	}{
		{
			name: "successful creation",
			payload: dto.PostPvzJSONRequestBody{
				City: dto.PVZRequestCity("Moscow"),
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid city",
			payload: dto.PostPvzJSONRequestBody{
				City: dto.PVZRequestCity("InvalidCity"),
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "city must be Moscow, Saint Petersburg or Kazan",
		},
		{
			name: "empty city",
			payload: dto.PostPvzJSONRequestBody{
				City: dto.PVZRequestCity(""),
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "city must be Moscow, Saint Petersburg or Kazan",
		},
		{
			name: "invalid payload format",
			payload: map[string]interface{}{
				"city": 123, // should be string
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrInvalidInput.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(bodyBytes))
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
				var resp dto.PVZ
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.NotNil(t, resp.Id)
				require.Equal(t, string(tt.payload.(dto.PostPvzJSONRequestBody).City), string(resp.City))
				require.False(t, resp.RegistrationDate.IsZero())
			}
		})
	}
}

func TestCreateReceptionHandler(t *testing.T) {
	handler, cleanup := setupTestPVZHandler(t)
	defer cleanup()

	router := gin.Default()
	router.POST("/receptions", handler.CreateReception)

	// Создаем тестовый PVZ без приемки
	pvzID := func() string {
		pg, err := pkgPostgres.New(testConnStr)
		require.NoError(t, err)
		defer pg.Close()

		id := uuid.New()
		_, err = pg.Pool.Exec(context.Background(), 
			`INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)`,
			id, time.Now().UTC(), "Moscow",
		)
		require.NoError(t, err)
		return id.String()
	}()

	tests := []struct {
		name         string
		payload      interface{}
		wantStatus   int
		wantErrorMsg string
		prepare      func() // Дополнительная подготовка перед тестом
	}{
		{
			name: "successful creation",
			payload: dto.PostReceptionsJSONRequestBody{
				PvzId: uuid.MustParse(pvzID),
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid pvz id format",
			payload: map[string]interface{}{
				"pvzId": "invalid-uuid",
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrInvalidPVZID.Error(),
		},
		{
			name: "non-existent pvz",
			payload: dto.PostReceptionsJSONRequestBody{
				PvzId: uuid.New(),
			},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "violates foreign key constraint",
		},
		
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Дополнительная подготовка если нужно
			if tt.prepare != nil {
				tt.prepare()
			}

			bodyBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Logf("Response body: %s", w.Body.String())
			}
			require.Equal(t, tt.wantStatus, w.Code, "unexpected status code")

			if tt.wantErrorMsg != "" {
				var errResp dto.Error
				err = json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
				require.Contains(t, errResp.Message, tt.wantErrorMsg, "error message mismatch")
			} else if tt.wantStatus == http.StatusCreated {
				var resp dto.Reception
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.NotEqual(t, uuid.Nil, resp.Id, "reception ID should not be empty")
				require.Equal(t, uuid.MustParse(pvzID), resp.PvzId, "pvz ID mismatch")
				require.Equal(t, dto.ReceptionStatus("in_progress"), resp.Status, "status should be in_progress")
				require.False(t, resp.DateTime.IsZero(), "date time should not be zero")
			}
		})
	}
}

func TestCreateProductHandler(t *testing.T) {
	handler, cleanup := setupTestPVZHandler(t)
	defer cleanup()

	router := gin.Default()
	router.POST("/products", handler.CreateProduct)

	// Создаем тестовые данные: PVZ -> Reception -> Product
	pvzID, receptionID := func() (string, string) {
		// Создаем PVZ
		pg, err := pkgPostgres.New(testConnStr)
		require.NoError(t, err)
		defer pg.Close()

		pvzID := uuid.New()
		_, err = pg.Pool.Exec(context.Background(),
			`INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)`,
			pvzID, time.Now().UTC(), "Moscow",
		)
		require.NoError(t, err)

		// Создаем Reception
		receptionID := uuid.New()
		_, err = pg.Pool.Exec(context.Background(),
			`INSERT INTO receptions (id, pvz_id, date_time, status) VALUES ($1, $2, $3, $4)`,
			receptionID, pvzID, time.Now().UTC(), "in_progress",
		)
		require.NoError(t, err)

		return pvzID.String(), receptionID.String()
	}()

	tests := []struct {
		name         string
		payload      interface{}
		wantStatus   int
		wantErrorMsg string
		prepare      func()
	}{
		{
			name: "successful creation - electronics",
			payload: dto.PostProductsJSONRequestBody{
				PvzId: uuid.MustParse(pvzID),
				Type:  dto.PostProductsJSONBodyTypeElectronics,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "successful creation - clothes",
			payload: dto.PostProductsJSONRequestBody{
				PvzId: uuid.MustParse(pvzID),
				Type:  dto.PostProductsJSONBodyTypeClothes,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "successful creation - shoes",
			payload: dto.PostProductsJSONRequestBody{
				PvzId: uuid.MustParse(pvzID),
				Type:  dto.PostProductsJSONBodyTypeShoes,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid pvz id format",
			payload: map[string]interface{}{
				"pvzId": "invalid-uuid",
				"type":  "electronics",
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrInvalidPVZID.Error(),
		},
		{
			name: "missing pvz id",
			payload: map[string]interface{}{
				"type": "electronics",
			},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "no rows in result set",
		},
		{
			name: "missing type",
			payload: map[string]interface{}{
				"pvzId": pvzID,
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "type must be electronics, clothes or shoes",
		},
		{
			name: "invalid type",
			payload: map[string]interface{}{
				"pvzId": pvzID,
				"type":  "invalid_type",
			},
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "type must be electronics, clothes or shoes",
		},
		{
			name: "non-existent pvz",
			payload: dto.PostProductsJSONRequestBody{
				PvzId: uuid.New(),
				Type:  dto.PostProductsJSONBodyTypeElectronics,
			},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "no rows in result set",
		},
		{
			name: "closed reception",
			payload: dto.PostProductsJSONRequestBody{
				PvzId: uuid.MustParse(pvzID),
				Type:  dto.PostProductsJSONBodyTypeElectronics,
			},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: "no rows in result set",
			prepare: func() {
				// Закрываем приемку
				pg, err := pkgPostgres.New(testConnStr)
				require.NoError(t, err)
				defer pg.Close()

				_, err = pg.Pool.Exec(context.Background(),
					`UPDATE receptions SET status = 'close' WHERE id = $1`,
					uuid.MustParse(receptionID),
				)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Дополнительная подготовка если нужно
			if tt.prepare != nil {
				tt.prepare()
			}

			bodyBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Logf("Response body: %s", w.Body.String())
			}
			require.Equal(t, tt.wantStatus, w.Code, "unexpected status code")

			if tt.wantErrorMsg != "" {
				var errResp dto.Error
				err = json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
				require.Contains(t, errResp.Message, tt.wantErrorMsg, "error message mismatch")
			} else if tt.wantStatus == http.StatusCreated {
				var resp dto.Product
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.NotEqual(t, uuid.Nil, resp.Id, "product ID should not be empty")
				require.Equal(t, uuid.MustParse(receptionID), resp.ReceptionId, "reception ID mismatch")
				require.NotEmpty(t, resp.Type, "product type should not be empty")
				require.False(t, resp.DateTime.IsZero(), "date time should not be zero")
			}
		})
	}
}

func TestDeleteLastProductHandler(t *testing.T) {
	handler, cleanup := setupTestPVZHandler(t)
	defer cleanup()

	router := gin.Default()
	router.POST("/pvz/:pvzId/delete_last_product", handler.DeleteLastProduct)

	// Создаем тестовые данные: PVZ -> Reception -> Products
	pvzID, productIDs := func() (string, []string) {
		pg, err := pkgPostgres.New(testConnStr)
		require.NoError(t, err)
		defer pg.Close()

		// Создаем PVZ
		pvzID := uuid.New()
		_, err = pg.Pool.Exec(context.Background(),
			`INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)`,
			pvzID, time.Now().UTC(), "Moscow",
		)
		require.NoError(t, err)

		// Создаем Reception
		receptionID := uuid.New()
		_, err = pg.Pool.Exec(context.Background(),
			`INSERT INTO receptions (id, pvz_id, date_time, status) VALUES ($1, $2, $3, $4)`,
			receptionID, pvzID, time.Now().UTC(), "in_progress",
		)
		require.NoError(t, err)

		// Создаем несколько товаров
		productIDs := make([]string, 3)
		for i := 0; i < 3; i++ {
			productID := uuid.New()
			_, err = pg.Pool.Exec(context.Background(),
				`INSERT INTO products (id, reception_id, date_time, type) VALUES ($1, $2, $3, $4)`,
				productID, receptionID, time.Now().UTC().Add(time.Duration(i)*time.Second), "electronics",
			)
			require.NoError(t, err)
			productIDs[i] = productID.String()
		}

		return pvzID.String(), productIDs
	}()

	tests := []struct {
		name           string
		pvzId          string
		wantStatus     int
		wantErrorMsg   string
		prepare        func()
		checkDBAfter   func(t *testing.T, pvzId string)
	}{
		{
			name:       "successful deletion",
			pvzId:      pvzID,
			wantStatus: http.StatusOK,
			checkDBAfter: func(t *testing.T, pvzId string) {
				pg, err := pkgPostgres.New(testConnStr)
				require.NoError(t, err)
				defer pg.Close()

				var count int
				err = pg.Pool.QueryRow(context.Background(),
					`SELECT COUNT(*) FROM products WHERE reception_id IN (
						SELECT id FROM receptions WHERE pvz_id = $1
					)`, uuid.MustParse(pvzId)).Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 2, count, "should have 2 products left after deletion")
			},
		},
		{
			name:       "delete all products one by one",
			pvzId:      pvzID,
			wantStatus: http.StatusOK,
			prepare: func() {
				// Удаляем первые два продукта, чтобы остался один
				pg, err := pkgPostgres.New(testConnStr)
				require.NoError(t, err)
				defer pg.Close()

				_, err = pg.Pool.Exec(context.Background(),
					`DELETE FROM products WHERE id IN ($1, $2)`,
					uuid.MustParse(productIDs[0]), uuid.MustParse(productIDs[1]),
				)
				require.NoError(t, err)
			},
			checkDBAfter: func(t *testing.T, pvzId string) {
				pg, err := pkgPostgres.New(testConnStr)
				require.NoError(t, err)
				defer pg.Close()

				var count int
				err = pg.Pool.QueryRow(context.Background(),
					`SELECT COUNT(*) FROM products WHERE reception_id IN (
						SELECT id FROM receptions WHERE pvz_id = $1
					)`, uuid.MustParse(pvzId)).Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 0, count, "should have no products left after last deletion")
			},
		},
		{
			name:         "invalid pvz id format",
			pvzId:        "invalid-uuid",
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrInvalidPVZID.Error(),
		},
		{
			name:         "non-existent pvz",
			pvzId:        uuid.New().String(),
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrNoActiveReception.Error(),
		},
		{
			name:       "no active reception",
			pvzId:      pvzID,
			wantStatus: http.StatusBadRequest,
			wantErrorMsg: pkgValidator.ErrNoActiveReception.Error(),
			prepare: func() {
				// Закрываем приемку
				pg, err := pkgPostgres.New(testConnStr)
				require.NoError(t, err)
				defer pg.Close()

				_, err = pg.Pool.Exec(context.Background(),
					`UPDATE receptions SET status = 'close' WHERE pvz_id = $1`,
					uuid.MustParse(pvzID),
				)
				require.NoError(t, err)
			},
		},
		{
			name:       "no products to delete",
			pvzId:      pvzID,
			wantStatus: http.StatusBadRequest,
			prepare: func() {
				// Удаляем все продукты
				pg, err := pkgPostgres.New(testConnStr)
				require.NoError(t, err)
				defer pg.Close()

				_, err = pg.Pool.Exec(context.Background(),
					`DELETE FROM products WHERE reception_id IN (
						SELECT id FROM receptions WHERE pvz_id = $1
					)`, uuid.MustParse(pvzID),
				)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Дополнительная подготовка если нужно
			if tt.prepare != nil {
				tt.prepare()
			}

			req := httptest.NewRequest(http.MethodPost, "/pvz/"+tt.pvzId+"/delete_last_product", nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Logf("Response body: %s", w.Body.String())
			}
			require.Equal(t, tt.wantStatus, w.Code, "unexpected status code")

			if tt.wantErrorMsg != "" {
				var errResp dto.Error
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				require.NoError(t, err)
				require.Contains(t, errResp.Message, tt.wantErrorMsg, "error message mismatch")
			}

			// Проверка состояния БД после выполнения
			if tt.checkDBAfter != nil {
				tt.checkDBAfter(t, tt.pvzId)
			}
		})
	}
}

func TestCloseReceptionHandler(t *testing.T) {
    handler, cleanup := setupTestPVZHandler(t)
    defer cleanup()

    router := gin.Default()
    router.POST("/pvz/:pvzId/close_last_reception", handler.CloseReception)

    // Создаем тестовые данные
    pg, err := pkgPostgres.New(testConnStr)
    require.NoError(t, err)
    defer pg.Close()

    // Вспомогательная функция для создания тестовых данных
    createTestData := func(withReception bool, receptionStatus string) (string, string) {
        // Создаем PVZ
        pvzID := uuid.New()
        _, err := pg.Pool.Exec(context.Background(),
            `INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)`,
            pvzID, time.Now().UTC(), "Moscow",
        )
        require.NoError(t, err)

        var receptionID string
        if withReception {
            // Создаем Reception
            receptionID = uuid.New().String()
            _, err = pg.Pool.Exec(context.Background(),
                `INSERT INTO receptions (id, pvz_id, date_time, status) VALUES ($1, $2, $3, $4)`,
                uuid.MustParse(receptionID), pvzID, time.Now().UTC(), receptionStatus,
            )
            require.NoError(t, err)
        }

        return pvzID.String(), receptionID
    }

    tests := []struct {
        name         string
        pvzId        string
        prepare      func() string // возвращает pvzId
        wantStatus   int
        wantErrorMsg string
        verify       func(t *testing.T, pvzId string, responseBody []byte)
    }{
        {
            name: "successful closure",
            prepare: func() string {
                pvzId, _ := createTestData(true, "in_progress")
                return pvzId
            },
            wantStatus: http.StatusOK,
            verify: func(t *testing.T, pvzId string, responseBody []byte) {
                // Проверяем ответ
                var resp dto.Reception
                require.NoError(t, json.Unmarshal(responseBody, &resp))
                require.Equal(t, "close", string(resp.Status))

                // Проверяем статус в БД
                var status string
                err := pg.Pool.QueryRow(context.Background(),
                    `SELECT status FROM receptions WHERE pvz_id = $1`,
                    uuid.MustParse(pvzId)).Scan(&status)
                require.NoError(t, err)
                require.Equal(t, "close", status)
            },
        },
        {
            name: "no active reception",
            prepare: func() string {
                pvzId, _ := createTestData(false, "")
                return pvzId
            },
            wantStatus:   http.StatusBadRequest,
            wantErrorMsg: pkgValidator.ErrNoActiveReception.Error(),
        },
        {
            name: "reception already closed",
            prepare: func() string {
                pvzId, _ := createTestData(true, "close")
                return pvzId
            },
            wantStatus:   http.StatusBadRequest,
            wantErrorMsg: pkgValidator.ErrNoActiveReception.Error(),
        },
        {
            name: "invalid pvz id format",
            pvzId:      "invalid-uuid",
            wantStatus: http.StatusBadRequest,
            wantErrorMsg: pkgValidator.ErrInvalidPVZID.Error(),
        },
        {
            name: "non-existent pvz",
            pvzId:      uuid.New().String(),
            wantStatus: http.StatusBadRequest,
            wantErrorMsg: pkgValidator.ErrNoActiveReception.Error(),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var pvzId string
            if tt.prepare != nil {
                pvzId = tt.prepare()
            } else {
                pvzId = tt.pvzId
            }

            req := httptest.NewRequest(http.MethodPost, "/pvz/"+pvzId+"/close_last_reception", nil)
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)

            require.Equal(t, tt.wantStatus, w.Code, "unexpected status code")

            if tt.wantErrorMsg != "" {
                var errResp dto.Error
                require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
                require.Contains(t, errResp.Message, tt.wantErrorMsg)
            }

            if tt.verify != nil {
                tt.verify(t, pvzId, w.Body.Bytes())
            }
        })
    }
}

func TestGetPVZsWithReceptionsHandler(t *testing.T) {
    handler, cleanup := setupTestPVZHandler(t)
    defer cleanup()

    router := gin.Default()
    router.GET("/pvz", handler.GetPVZsWithReceptions)

    // Подготовка тестовых данных
    pg, err := pkgPostgres.New(testConnStr)
    require.NoError(t, err)
    defer pg.Close()

    now := time.Now().UTC()
    testData := []struct {
        pvzID           uuid.UUID
        registrationDate time.Time
        city             string
        receptions       []struct {
            id        uuid.UUID
            dateTime  time.Time
            status    string
            products  []struct {
                id       uuid.UUID
                dateTime time.Time
                ptype    string
            }
        }
    }{
        {
            pvzID:           uuid.New(),
            registrationDate: now.Add(-24 * time.Hour),
            city:             "Moscow",
            receptions: []struct {
                id        uuid.UUID
                dateTime  time.Time
                status    string
                products  []struct {
                    id       uuid.UUID
                    dateTime time.Time
                    ptype    string
                }
            }{
                {
                    id:       uuid.New(),
                    dateTime: now.Add(-12 * time.Hour),
                    status:   "close",
                    products: []struct {
                        id       uuid.UUID
                        dateTime time.Time
                        ptype    string
                    }{
                        {id: uuid.New(), dateTime: now.Add(-11 * time.Hour), ptype: "electronics"},
                        {id: uuid.New(), dateTime: now.Add(-10 * time.Hour), ptype: "clothes"},
                    },
                },
            },
        },
        {
            pvzID:           uuid.New(),
            registrationDate: now.Add(-2 * time.Hour),
            city:             "Saint Petersburg",
            receptions: []struct {
                id        uuid.UUID
                dateTime  time.Time
                status    string
                products  []struct {
                    id       uuid.UUID
                    dateTime time.Time
                    ptype    string
                }
            }{
                {
                    id:       uuid.New(),
                    dateTime: now.Add(-1 * time.Hour),
                    status:   "in_progress",
                    products: []struct {
                        id       uuid.UUID
                        dateTime time.Time
                        ptype    string
                    }{
                        {id: uuid.New(), dateTime: now.Add(-30 * time.Minute), ptype: "shoes"},
                    },
                },
            },
        },
    }

    // Заполняем БД тестовыми данными
    for _, data := range testData {
        // Создаем PVZ
        _, err := pg.Pool.Exec(context.Background(),
            `INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)`,
            data.pvzID, data.registrationDate, data.city,
        )
        require.NoError(t, err)

        // Создаем Reception и Products
        for _, rec := range data.receptions {
            _, err := pg.Pool.Exec(context.Background(),
                `INSERT INTO receptions (id, pvz_id, date_time, status) VALUES ($1, $2, $3, $4)`,
                rec.id, data.pvzID, rec.dateTime, rec.status,
            )
            require.NoError(t, err)

            for _, prod := range rec.products {
                _, err := pg.Pool.Exec(context.Background(),
                    `INSERT INTO products (id, reception_id, date_time, type) VALUES ($1, $2, $3, $4)`,
                    prod.id, rec.id, prod.dateTime, prod.ptype,
                )
                require.NoError(t, err)
            }
        }
    }

    tests := []struct {
        name           string
        queryParams    string
        wantStatus     int
        wantPVZCount   int
        wantErrorMsg   string
        verifyResponse func(t *testing.T, body []byte)
    }{
        {
            name:         "successful request without filters",
            queryParams:  "",
            wantStatus:   http.StatusOK,
            wantPVZCount: 2,
            verifyResponse: func(t *testing.T, body []byte) {
                var response []dto.PVZWithReceptions
                require.NoError(t, json.Unmarshal(body, &response))
                require.Len(t, response, 2)
                
                // Проверяем сортировку (по убыванию даты регистрации)
                require.True(t, response[0].Pvz.RegistrationDate.After(response[1].Pvz.RegistrationDate))
                
                // Проверяем структуру данных
                for _, pvz := range response {
                    require.NotEmpty(t, pvz.Pvz.Id)
                    require.NotEmpty(t, pvz.Pvz.City)
                    require.False(t, pvz.Pvz.RegistrationDate.IsZero())
                    
                    for _, rec := range pvz.Receptions {
                        require.NotEmpty(t, rec.Reception.Id)
                        require.Equal(t, pvz.Pvz.Id, rec.Reception.PvzId)
                        require.False(t, rec.Reception.DateTime.IsZero())
                        require.Contains(t, []string{"in_progress", "close"}, string(rec.Reception.Status))
                        
                        // Проверяем сортировку товаров (по убыванию даты)
                        for i := 1; i < len(rec.Products); i++ {
                            require.True(t, rec.Products[i-1].DateTime.After(rec.Products[i].DateTime))
                        }
                    }
                }
            },
        },
        {
            name:         "pagination - first page",
            queryParams:  "page=1&limit=1",
            wantStatus:   http.StatusOK,
            wantPVZCount: 1,
            verifyResponse: func(t *testing.T, body []byte) {
                var response []dto.PVZWithReceptions
                require.NoError(t, json.Unmarshal(body, &response))
                require.Len(t, response, 1)
                // Должен вернуться самый новый PVZ (Saint Petersburg)
                require.Equal(t, "Saint Petersburg", string(response[0].Pvz.City))
            },
        },
        {
            name:         "pagination - second page",
            queryParams:  "page=2&limit=1",
            wantStatus:   http.StatusOK,
            wantPVZCount: 1,
            verifyResponse: func(t *testing.T, body []byte) {
                var response []dto.PVZWithReceptions
                require.NoError(t, json.Unmarshal(body, &response))
                require.Len(t, response, 1)
                // Должен вернуться второй PVZ (Moscow)
                require.Equal(t, "Moscow", string(response[0].Pvz.City))
            },
        },
        {
            name:         "invalid date format",
            queryParams: "startDate=invalid-date",
            wantStatus:   http.StatusBadRequest,
            wantErrorMsg: "date must be in RFC3339 format",
        },
        {
            name:         "invalid page number",
            queryParams: "page=0",
            wantStatus:   http.StatusBadRequest,
            wantErrorMsg: "page must be greater than 0",
        },
        {
            name:         "invalid limit value",
            queryParams: "limit=0",
            wantStatus:   http.StatusBadRequest,
            wantErrorMsg: "limit must be between 1 and 30",
        },
        {
            name:         "start date after end date",
            queryParams: fmt.Sprintf("startDate=%s&endDate=%s", 
                url.QueryEscape(now.Format(time.RFC3339)),
                url.QueryEscape(now.Add(-1*time.Hour).Format(time.RFC3339))),
            wantStatus:   http.StatusBadRequest,
            wantErrorMsg: "end date must be after start date",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodGet, "/pvz?"+tt.queryParams, nil)
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)

            require.Equal(t, tt.wantStatus, w.Code)

            if tt.wantErrorMsg != "" {
                var errResp dto.Error
                require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
                require.Contains(t, errResp.Message, tt.wantErrorMsg)
            } else {
                if tt.verifyResponse != nil {
                    tt.verifyResponse(t, w.Body.Bytes())
                }
                
                if tt.wantPVZCount > 0 {
                    var response []dto.PVZWithReceptions
                    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
                    require.Len(t, response, tt.wantPVZCount)
                }
            }
        })
    }
}