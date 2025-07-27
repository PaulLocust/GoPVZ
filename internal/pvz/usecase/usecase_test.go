package usecase

import (
	"GoPVZ/internal/pvz/entity"
	"GoPVZ/pkg/pkgValidator"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPVZRepo struct {
    mock.Mock
}

func (m *MockPVZRepo) CreatePVZ(ctx context.Context, pvz *entity.PVZ) error {
    args := m.Called(ctx, pvz)
    return args.Error(0)
}

func (m *MockPVZRepo) GetById(ctx context.Context, id string) (*entity.PVZ, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*entity.PVZ), args.Error(1)
}

func (m *MockPVZRepo) CreateReception(ctx context.Context, reception *entity.Reception) error {
    args := m.Called(ctx, reception)
    return args.Error(0)
}

func (m *MockPVZRepo) CheckPvzsLastReceptionStatusInProgress(ctx context.Context, pvzId string) (bool, error) {
    args := m.Called(ctx, pvzId)
    return args.Bool(0), args.Error(1)
}

func (m *MockPVZRepo) CreateProduct(ctx context.Context, product *entity.Product) error {
    args := m.Called(ctx, product)
    return args.Error(0)
}

func (m *MockPVZRepo) GetInProgressReceptionIdByPVZId(ctx context.Context, pvzId string) (string, error) {
    args := m.Called(ctx, pvzId)
    return args.String(0), args.Error(1)
}

func (m *MockPVZRepo) DeleteLastProductFromReception(ctx context.Context, pvzId string) error {
    args := m.Called(ctx, pvzId)
    return args.Error(0)
}

func (m *MockPVZRepo) CloseReception(ctx context.Context, pvzId string) (*entity.Reception, error) {
    args := m.Called(ctx, pvzId)
    return args.Get(0).(*entity.Reception), args.Error(1)
}

func (m *MockPVZRepo) GetPVZsWithReceptions(ctx context.Context, startDate, endDate *time.Time, limit, offset int) ([]*entity.PVZWithReceptions, error) {
    args := m.Called(ctx, startDate, endDate, limit, offset)
    return args.Get(0).([]*entity.PVZWithReceptions), args.Error(1)
}


func TestPVZUseCase_CreatePVZ(t *testing.T) {
	tests := []struct {
		name      string
		city      string
		repoError error
		wantError bool
	}{
		{
			name:      "success",
			city:      "Kazan",
			repoError: nil,
			wantError: false,
		},
		{
			name:      "repository error",
			city:      "Moscow",
			repoError: errors.New("db error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepo)
			uc := NewPVZUseCase(mockRepo)

			mockRepo.On("CreatePVZ", mock.Anything, mock.MatchedBy(func(pvz *entity.PVZ) bool {
				return pvz.City == entity.City(tt.city)
			})).Return(tt.repoError)

			result, err := uc.CreatePVZ(context.Background(), tt.city)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, entity.City(tt.city), result.City)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}


func TestPVZUseCase_CreateReception(t *testing.T) {
	tests := []struct {
		name            string
		pvzId           string
		isInProgress    bool
		repoError       error
		receptionError  error
		wantError       bool
	}{
		{
			name:           "success",
			pvzId:          uuid.New().String(),
			isInProgress:   false,
			repoError:      nil,
			receptionError: nil,
			wantError:      false,
		},
		{
			name:           "reception in progress",
			pvzId:          uuid.New().String(),
			isInProgress:   true,
			repoError:      nil,
			receptionError: pkgValidator.ErrInvalidReceptionCreation,
			wantError:      true,
		},
		{
			name:           "repository error",
			pvzId:          uuid.New().String(),
			isInProgress:   false,
			repoError:      errors.New("db error"),
			receptionError: nil,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepo)
			uc := NewPVZUseCase(mockRepo)

			mockRepo.On("CheckPvzsLastReceptionStatusInProgress", mock.Anything, tt.pvzId).
				Return(tt.isInProgress, tt.repoError)

			if !tt.isInProgress && tt.repoError == nil {
				mockRepo.On("CreateReception", mock.Anything, mock.MatchedBy(func(r *entity.Reception) bool {
					return r.PvzID.String() == tt.pvzId && r.Status == entity.StatusInProgress
				})).Return(tt.receptionError)
			}

			result, err := uc.CreateReception(context.Background(), tt.pvzId)

			if tt.wantError {
				assert.Error(t, err)
				if tt.receptionError != nil {
					assert.ErrorIs(t, err, tt.receptionError)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, entity.StatusInProgress, result.Status)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZUseCase_CreateProduct(t *testing.T) {
	testReceptionId := uuid.New().String()
	tests := []struct {
		name           string
		productType    string
		pvzId          string
		receptionId    string
		repoError      error
		wantError      bool
	}{
		{
			name:          "success",
			productType:   "electronics",
			pvzId:         uuid.New().String(),
			receptionId:   testReceptionId,
			repoError:     nil,
			wantError:     false,
		},
		{
			name:          "no active reception",
			productType:   "clothes",
			pvzId:         uuid.New().String(),
			receptionId:   "",
			repoError:     errors.New("no active reception"),
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepo)
			uc := NewPVZUseCase(mockRepo)

			mockRepo.On("GetInProgressReceptionIdByPVZId", mock.Anything, tt.pvzId).
				Return(tt.receptionId, tt.repoError)

			if tt.receptionId != "" {
				mockRepo.On("CreateProduct", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
					return p.ReceptionID.String() == tt.receptionId && p.Type == entity.Type(tt.productType)
				})).Return(nil)
			}

			result, err := uc.CreateProduct(context.Background(), tt.productType, tt.pvzId)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, entity.Type(tt.productType), result.Type)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZUseCase_DeleteLastProduct(t *testing.T) {
	tests := []struct {
		name            string
		pvzId           string
		isInProgress    bool
		repoError       error
		deleteError     error
		wantError       bool
	}{
		{
			name:           "success",
			pvzId:          uuid.New().String(),
			isInProgress:   true,
			repoError:      nil,
			deleteError:    nil,
			wantError:      false,
		},
		{
			name:           "no active reception",
			pvzId:          uuid.New().String(),
			isInProgress:   false,
			repoError:      nil,
			deleteError:    pkgValidator.ErrNoActiveReception,
			wantError:      true,
		},
		{
			name:           "repository error",
			pvzId:          uuid.New().String(),
			isInProgress:   true,
			repoError:      errors.New("db error"),
			deleteError:    nil,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepo)
			uc := NewPVZUseCase(mockRepo)

			mockRepo.On("CheckPvzsLastReceptionStatusInProgress", mock.Anything, tt.pvzId).
				Return(tt.isInProgress, tt.repoError)

			if tt.isInProgress && tt.repoError == nil {
				mockRepo.On("DeleteLastProductFromReception", mock.Anything, tt.pvzId).
					Return(tt.deleteError)
			}

			err := uc.DeleteLastProduct(context.Background(), tt.pvzId)

			if tt.wantError {
				assert.Error(t, err)
				if tt.deleteError != nil {
					assert.ErrorIs(t, err, tt.deleteError)
				}
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZUseCase_CloseReception(t *testing.T) {
	testReception := &entity.Reception{
		ID:       uuid.New(),
		PvzID:    uuid.New(),
		DateTime: time.Now().UTC(),
		Status:   entity.StatusClose,
	}
	
	tests := []struct {
		name            string
		pvzId           string
		isInProgress    bool
		repoError       error
		closeError      error
		wantError       bool
	}{
		{
			name:           "success",
			pvzId:          uuid.New().String(),
			isInProgress:   true,
			repoError:      nil,
			closeError:     nil,
			wantError:      false,
		},
		{
			name:           "no active reception",
			pvzId:          uuid.New().String(),
			isInProgress:   false,
			repoError:      nil,
			closeError:     pkgValidator.ErrNoActiveReception,
			wantError:      true,
		},
		{
			name:           "repository error",
			pvzId:          uuid.New().String(),
			isInProgress:   true,
			repoError:      errors.New("db error"),
			closeError:    nil,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepo)
			uc := NewPVZUseCase(mockRepo)

			mockRepo.On("CheckPvzsLastReceptionStatusInProgress", mock.Anything, tt.pvzId).
				Return(tt.isInProgress, tt.repoError)

			if tt.isInProgress && tt.repoError == nil {
				mockRepo.On("CloseReception", mock.Anything, tt.pvzId).
					Return(testReception, tt.closeError)
			}

			result, err := uc.CloseReception(context.Background(), tt.pvzId)

			if tt.wantError {
				assert.Error(t, err)
				if tt.closeError != nil {
					assert.ErrorIs(t, err, tt.closeError)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, entity.StatusClose, result.Status)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZUseCase_GetPVZsWithReceptions(t *testing.T) {
    testPVZ := &entity.PVZ{
        ID:               uuid.New(),
        City:             "Kazan",
        RegistrationDate: time.Now().UTC(),
    }

    testReception := &entity.Reception{
        ID:       uuid.New(),
        PvzID:    uuid.New(),
        DateTime: time.Now().UTC(),
        Status:   "in_progress",
    }

    testProduct := &entity.Product{
        ID:          uuid.New(),
        ReceptionID: uuid.New(),
        DateTime:    time.Now().UTC(),
        Type:        "electronics",
    }

    testData := []*entity.PVZWithReceptions{
        {
            PVZ: testPVZ,
            Receptions: []*entity.ReceptionWithProducts{
                {
                    Reception: testReception,
                    Products:  []*entity.Product{testProduct},
                },
            },
        },
    }

    tests := []struct {
        name        string
        startDate   *time.Time
        endDate     *time.Time
        page        int
        limit       int
        repoResult  []*entity.PVZWithReceptions
        repoError   error
        wantError   bool
    }{
        {
            name:       "success with pagination",
            startDate:  nil,
            endDate:    nil,
            page:       1,
            limit:      10,
            repoResult: testData,
            repoError:  nil,
            wantError:  false,
        },
        {
            name:       "invalid page",
            startDate:  nil,
            endDate:    nil,
            page:       -1,
            limit:      10,
            repoResult: testData,
            repoError:  nil,
            wantError:  false, // page корректируется в usecase
        },
        {
            name:       "repository error",
            startDate:  nil,
            endDate:    nil,
            page:       1,
            limit:      10,
            repoResult: nil,
            repoError:  errors.New("db error"),
            wantError:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockPVZRepo)
            uc := NewPVZUseCase(mockRepo)

            expectedLimit := tt.limit
            if expectedLimit < 1 || expectedLimit > 30 {
                expectedLimit = 10
            }
            expectedOffset := (tt.page - 1) * expectedLimit
            if expectedOffset < 0 {
                expectedOffset = 0
            }

            mockRepo.On("GetPVZsWithReceptions", mock.Anything, tt.startDate, tt.endDate, expectedLimit, expectedOffset).
                Return(tt.repoResult, tt.repoError)

            result, err := uc.GetPVZsWithReceptions(context.Background(), tt.startDate, tt.endDate, tt.page, tt.limit)

            if tt.wantError {
                assert.Error(t, err)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                if tt.repoResult != nil {
                    assert.Equal(t, len(tt.repoResult), len(result))
                    // Дополнительные проверки структуры данных
                    if len(result) > 0 {
                        assert.Equal(t, tt.repoResult[0].PVZ.ID, result[0].PVZ.ID)
                        assert.Equal(t, len(tt.repoResult[0].Receptions), len(result[0].Receptions))
                    }
                }
            }
            mockRepo.AssertExpectations(t)
        })
    }
}