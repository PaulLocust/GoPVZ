package usecase

import (
	"GoPVZ/internal/pvz/entity"
	"GoPVZ/internal/pvz/repo"
	"GoPVZ/pkg/pkgValidator"
	"GoPVZ/pkg/pkgMetrics"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PVZUseCase struct {
	repo repo.PVZRepository
}

func NewPVZUseCase(r repo.PVZRepository) *PVZUseCase {
	return &PVZUseCase{repo: r}
}

func (uc *PVZUseCase) CreatePVZ(ctx context.Context, city string) (*entity.PVZ, error) {
	pvz := &entity.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now().UTC(),
		City:             entity.City(city),
	}

	if err := uc.repo.CreatePVZ(ctx, pvz); err != nil {
		return nil, err
	}

	// Метрика: количество созданных ПВЗ
	pkgMetrics.PVZCreatedTotal.Inc()
	return pvz, nil
}

func (uc *PVZUseCase) CreateReception(ctx context.Context, pvzId string) (*entity.Reception, error) {
	pvzUUID, err := uuid.Parse(pvzId)
	if err != nil {
		return nil, err
	}

	isInProgress, err := uc.repo.CheckPvzsLastReceptionStatusInProgress(ctx, pvzId)
	if err != nil {
		return nil, err
	}

	if isInProgress {
		return nil, pkgValidator.ErrInvalidReceptionCreation
	}

	reception := &entity.Reception{
		ID:       uuid.New(),
		PvzID:    pvzUUID,
		DateTime: time.Now().UTC(),
		Status:   entity.StatusInProgress,
	}

	if err := uc.repo.CreateReception(ctx, reception); err != nil {
		return nil, err
	}

	// Метрика: количество созданных приемок
	pkgMetrics.ReceptionsCreatedTotal.Inc()
	return reception, nil
}

func (uc *PVZUseCase) CreateProduct(ctx context.Context, productType, pvzId string) (*entity.Product, error) {
	receptionId, err := uc.repo.GetInProgressReceptionIdByPVZId(ctx, pvzId)
	if err != nil {
		return nil, err
	}

	receptionUUID, err := uuid.Parse(receptionId)
	if err != nil {
		return nil, err
	}

	product := &entity.Product{
		ID:          uuid.New(),
		ReceptionID: receptionUUID,
		DateTime:    time.Now().UTC(),
		Type:        entity.Type(productType),
	}

	if err := uc.repo.CreateProduct(ctx, product); err != nil {
		return nil, err
	}

	// Метрика: количество добавленных товаров
	pkgMetrics.ProductsAddedTotal.Inc()
	return product, nil
}

func (uc *PVZUseCase) DeleteLastProduct(ctx context.Context, pvzId string) error {
	isInProgress, err := uc.repo.CheckPvzsLastReceptionStatusInProgress(ctx, pvzId)
	if err != nil {
		return err
	}
	if !isInProgress {
		return pkgValidator.ErrNoActiveReception
	}

	return uc.repo.DeleteLastProductFromReception(ctx, pvzId)
}

func (uc *PVZUseCase) CloseReception(ctx context.Context, pvzId string) (*entity.Reception, error) {
	isInProgress, err := uc.repo.CheckPvzsLastReceptionStatusInProgress(ctx, pvzId)
	if err != nil {
		return nil, err
	}
	if !isInProgress {
		return nil, pkgValidator.ErrNoActiveReception
	}

	return uc.repo.CloseReception(ctx, pvzId)
}

func (uc *PVZUseCase) GetPVZsWithReceptions(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]*entity.PVZWithReceptions, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 30 {
		limit = 10
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return nil, errors.New("startDate cannot be after endDate")
	}

	return uc.repo.GetPVZsWithReceptions(ctx, startDate, endDate, limit, (page-1)*limit)
}