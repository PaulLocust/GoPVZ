package repo

import (
	"context"
	"os"
	"testing"
	"time"

	"GoPVZ/internal/pvz/entity"
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

	_, err = pg.Pool.Exec(ctx, `
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

func setupPVZRepo(t *testing.T) (PVZRepository, func()) {
	pg, err := pkgPostgres.New(testConnStr)
	require.NoError(t, err)

	_, err = pg.Pool.Exec(context.Background(), `
		TRUNCATE TABLE products, receptions, pvz CASCADE
	`)
	require.NoError(t, err)

	repo := NewPVZRepo(pg.Pool)

	return repo, func() {
		pg.Close()
	}
}

func TestPVZRepository_CreatePVZ(t *testing.T) {
	repo, cleanup := setupPVZRepo(t)
	defer cleanup()

	tests := []struct {
		name        string
		pvz         *entity.PVZ
		wantError   bool
		errorString string
	}{
		{
			name: "valid PVZ",
			pvz: &entity.PVZ{
				ID:               uuid.New(),
				RegistrationDate: time.Now(),
				City:             "Moscow",
			},
			wantError: false,
		},
		{
			name: "invalid city",
			pvz: &entity.PVZ{
				ID:               uuid.New(),
				RegistrationDate: time.Now(),
				City:             "InvalidCity",
			},
			wantError:   true,
			errorString: "violates check constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreatePVZ(context.Background(), tt.pvz)
			if tt.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPVZRepository_CreateReception(t *testing.T) {
	repo, cleanup := setupPVZRepo(t)
	defer cleanup()

	ctx := context.Background()
	validPVZ := &entity.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		City:             "Kazan",
	}
	require.NoError(t, repo.CreatePVZ(ctx, validPVZ))

	tests := []struct {
		name        string
		reception   *entity.Reception
		wantError   bool
		errorString string
	}{
		{
			name: "valid reception",
			reception: &entity.Reception{
				ID:       uuid.New(),
				PvzID:    validPVZ.ID,
				DateTime: time.Now(),
				Status:   entity.StatusInProgress,
			},
			wantError: false,
		},
		{
			name: "invalid status",
			reception: &entity.Reception{
				ID:       uuid.New(),
				PvzID:    validPVZ.ID,
				DateTime: time.Now(),
				Status:   "invalid_status",
			},
			wantError:   true,
			errorString: "violates check constraint",
		},
		{
			name: "invalid pvz_id",
			reception: &entity.Reception{
				ID:       uuid.New(),
				PvzID:    uuid.New(),
				DateTime: time.Now(),
				Status:   entity.StatusInProgress,
			},
			wantError:   true,
			errorString: "violates foreign key constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateReception(ctx, tt.reception)
			if tt.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPVZRepository_CreateProduct(t *testing.T) {
	repo, cleanup := setupPVZRepo(t)
	defer cleanup()

	ctx := context.Background()
	pvz := &entity.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), City: "Saint Petersburg"}
	require.NoError(t, repo.CreatePVZ(ctx, pvz))

	reception := &entity.Reception{ID: uuid.New(), PvzID: pvz.ID, DateTime: time.Now(), Status: entity.StatusInProgress}
	require.NoError(t, repo.CreateReception(ctx, reception))

	tests := []struct {
		name        string
		product     *entity.Product
		wantError   bool
		errorString string
	}{
		{
			name: "valid product",
			product: &entity.Product{
				ID:          uuid.New(),
				ReceptionID: reception.ID,
				DateTime:    time.Now(),
				Type:        "electronics",
			},
			wantError: false,
		},
		{
			name: "invalid type",
			product: &entity.Product{
				ID:          uuid.New(),
				ReceptionID: reception.ID,
				DateTime:    time.Now(),
				Type:        "food",
			},
			wantError:   true,
			errorString: "violates check constraint",
		},
		{
			name: "invalid reception_id",
			product: &entity.Product{
				ID:          uuid.New(),
				ReceptionID: uuid.New(),
				DateTime:    time.Now(),
				Type:        "clothes",
			},
			wantError:   true,
			errorString: "violates foreign key constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateProduct(ctx, tt.product)
			if tt.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPVZRepository_CloseReception(t *testing.T) {
	repo, cleanup := setupPVZRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Подготовка PVZ и активной приёмки
	pvz := &entity.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), City: "Moscow"}
	require.NoError(t, repo.CreatePVZ(ctx, pvz))

	reception := &entity.Reception{
		ID:       uuid.New(),
		PvzID:    pvz.ID,
		DateTime: time.Now(),
		Status:   entity.StatusInProgress,
	}
	require.NoError(t, repo.CreateReception(ctx, reception))

	tests := []struct {
		name        string
		pvzID       string
		setup       func()
		wantError   bool
		errorString string
	}{
		{
			name:  "close active reception successfully",
			pvzID: pvz.ID.String(),
			setup: func() {},
			wantError: false,
		},
		{
			name:  "no active reception",
			pvzID: uuid.New().String(),
			setup: func() {},
			wantError:   true,
			errorString: "no rows in result set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			reception, err := repo.CloseReception(ctx, tt.pvzID)
			if tt.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorString)
			} else {
				require.NoError(t, err)
				require.Equal(t, entity.StatusClose, reception.Status)
			}
		})
	}
}


func TestPVZRepository_DeleteLastProductFromReception(t *testing.T) {
	repo, cleanup := setupPVZRepo(t)
	defer cleanup()

	ctx := context.Background()

	pvz := &entity.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), City: "Kazan"}
	require.NoError(t, repo.CreatePVZ(ctx, pvz))

	reception := &entity.Reception{
		ID:       uuid.New(),
		PvzID:    pvz.ID,
		DateTime: time.Now(),
		Status:   entity.StatusInProgress,
	}
	require.NoError(t, repo.CreateReception(ctx, reception))

	// Добавим 2 товара
	product1 := &entity.Product{ID: uuid.New(), ReceptionID: reception.ID, DateTime: time.Now().Add(-2 * time.Minute), Type: "clothes"}
	product2 := &entity.Product{ID: uuid.New(), ReceptionID: reception.ID, DateTime: time.Now(), Type: "shoes"}

	require.NoError(t, repo.CreateProduct(ctx, product1))
	require.NoError(t, repo.CreateProduct(ctx, product2))

	tests := []struct {
		name        string
		pvzID       string
		wantDeleted uuid.UUID // ожидаем, что будет удалён product2
		wantError   bool
		errorString string
	}{
		{
			name:        "delete last product successfully",
			pvzID:       pvz.ID.String(),
			wantDeleted: product2.ID,
			wantError:   false,
		},
		{
			name:        "no active reception for random PVZ",
			pvzID:       uuid.New().String(),
			wantError:   true,
			errorString: "no rows in result set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.DeleteLastProductFromReception(ctx, tt.pvzID)
			if tt.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorString)
			} else {
				require.NoError(t, err)

				// Проверим, остался только product1
				products := getProductsForReception(t, repo, reception.ID)
				require.Len(t, products, 1)
				require.Equal(t, product1.ID, products[0].ID)
			}
		})
	}
}


func getProductsForReception(t *testing.T, repo PVZRepository, receptionID uuid.UUID) []*entity.Product {
	ctx := context.Background()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(24 * time.Hour)

	pvzs, err := repo.GetPVZsWithReceptions(ctx, &start, &end, 10, 0)
	require.NoError(t, err)

	for _, pvz := range pvzs {
		for _, r := range pvz.Receptions {
			if r.Reception.ID == receptionID {
				return r.Products
			}
		}
	}
	return nil
}

func TestPVZRepository_GetPVZsWithReceptions(t *testing.T) {
    repo, cleanup := setupPVZRepo(t)
    defer cleanup()

    ctx := context.Background()
    now := time.Now().UTC()
    yesterday := now.AddDate(0, 0, -1)

    // Создаем тестовые данные
    pvz1 := &entity.PVZ{
        ID:               uuid.New(),
        RegistrationDate: yesterday,
        City:             "Moscow",
    }
    pvz2 := &entity.PVZ{
        ID:               uuid.New(),
        RegistrationDate: now,
        City:             "Saint Petersburg",
    }
    require.NoError(t, repo.CreatePVZ(ctx, pvz1))
    require.NoError(t, repo.CreatePVZ(ctx, pvz2))

    // Приемки для pvz1
    reception1 := &entity.Reception{
        ID:       uuid.New(),
        PvzID:    pvz1.ID,
        DateTime: yesterday.Add(2 * time.Hour),
        Status:   entity.StatusInProgress,
    }
    reception2 := &entity.Reception{
        ID:       uuid.New(),
        PvzID:    pvz1.ID,
        DateTime: now.Add(-1 * time.Hour),
        Status:   entity.StatusClose,
    }
    require.NoError(t, repo.CreateReception(ctx, reception1))
    require.NoError(t, repo.CreateReception(ctx, reception2))

    // Товары для reception1
    product1 := &entity.Product{
        ID:          uuid.New(),
        ReceptionID: reception1.ID,
        DateTime:    yesterday.Add(2*time.Hour + 5*time.Minute),
        Type:        "electronics",
    }
    product2 := &entity.Product{
        ID:          uuid.New(),
        ReceptionID: reception1.ID,
        DateTime:    yesterday.Add(2*time.Hour + 10*time.Minute),
        Type:        "clothes",
    }
    require.NoError(t, repo.CreateProduct(ctx, product1))
    require.NoError(t, repo.CreateProduct(ctx, product2))

    // Приемка для pvz2 без товаров
    reception3 := &entity.Reception{
        ID:       uuid.New(),
        PvzID:    pvz2.ID,
        DateTime: now,
        Status:   entity.StatusInProgress,
    }
    require.NoError(t, repo.CreateReception(ctx, reception3))

    tests := []struct {
        name           string
        startDate      *time.Time
        endDate        *time.Time
        limit          int
        offset         int
        expectedPVZs   int
        expectedRecs   map[uuid.UUID]int // map[pvzID]количество_приемок
        expectedProds  map[uuid.UUID]int // map[receptionID]количество_товаров
    }{
        {
            name:           "get all without filters",
            startDate:      nil,
            endDate:        nil,
            limit:          10,
            offset:         0,
            expectedPVZs:   2,
            expectedRecs:   map[uuid.UUID]int{pvz1.ID: 2, pvz2.ID: 1},
            expectedProds:  map[uuid.UUID]int{reception1.ID: 2},
        },
        
        {
            name:           "pagination - first page",
            startDate:      nil,
            endDate:        nil,
            limit:          1,
            offset:         0,
            expectedPVZs:   1,
            expectedRecs:   map[uuid.UUID]int{pvz2.ID: 1}, // pvz2 должен быть первым из-за сортировки
            expectedProds:  map[uuid.UUID]int{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := repo.GetPVZsWithReceptions(ctx, tt.startDate, tt.endDate, tt.limit, tt.offset)
            require.NoError(t, err)
            require.Len(t, result, tt.expectedPVZs)

            // Проверяем структуру ответа
            for _, pvzWithRecs := range result {
                // Проверяем наличие PVZ
                require.NotNil(t, pvzWithRecs.PVZ)
                
                // Проверяем количество приемок
                expectedRecCount, ok := tt.expectedRecs[pvzWithRecs.PVZ.ID]
                require.True(t, ok, "unexpected PVZ in result")
                require.Len(t, pvzWithRecs.Receptions, expectedRecCount)

                // Проверяем товары в приемках
                for _, recWithProds := range pvzWithRecs.Receptions {
                    expectedProdCount, ok := tt.expectedProds[recWithProds.Reception.ID]
                    if ok {
                        require.Len(t, recWithProds.Products, expectedProdCount)
                    } else {
                        require.Empty(t, recWithProds.Products)
                    }
                }
            }

            // Проверяем сортировку PVZ по registration_date DESC
            if len(result) > 1 {
                require.True(t, 
                    result[0].PVZ.RegistrationDate.After(result[1].PVZ.RegistrationDate),
                    "PVZs should be sorted by registration_date DESC")
            }

            // Проверяем сортировку приемок внутри PVZ по date_time DESC
            for _, pvzWithRecs := range result {
                if len(pvzWithRecs.Receptions) > 1 {
                    require.True(t,
                        pvzWithRecs.Receptions[0].Reception.DateTime.After(
                            pvzWithRecs.Receptions[1].Reception.DateTime),
                        "Receptions should be sorted by date_time DESC")
                }
            }

            // Проверяем сортировку товаров внутри приемок по date_time DESC
            for _, pvzWithRecs := range result {
                for _, recWithProds := range pvzWithRecs.Receptions {
                    if len(recWithProds.Products) > 1 {
                        require.True(t,
                            recWithProds.Products[0].DateTime.After(
                                recWithProds.Products[1].DateTime),
                            "Products should be sorted by date_time DESC")
                    }
                }
            }
        })
    }
}