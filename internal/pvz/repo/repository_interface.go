package repo

import (
	"GoPVZ/internal/pvz/entity"
	"context"
	"time"
)

type PVZRepository interface {
	CreatePVZ(ctx context.Context, user *entity.PVZ) error
	GetById(ctx context.Context, id string) (*entity.PVZ, error)

	CreateReception(ctx context.Context, reception *entity.Reception) error
	CheckPvzsLastReceptionStatusInProgress(ctx context.Context, pvzId string) (bool, error)

	CreateProduct(ctx context.Context, product *entity.Product) error
	GetInProgressReceptionIdByPVZId(ctx context.Context, pvzId string) (string, error)
	DeleteLastProductFromReception(ctx context.Context, pvzId string) error
	CloseReception(ctx context.Context, pvzId string) (*entity.Reception, error)
	GetPVZsWithReceptions(ctx context.Context, startDate, endDate *time.Time, limit, offset int) ([]*entity.PVZWithReceptions, error)

}
