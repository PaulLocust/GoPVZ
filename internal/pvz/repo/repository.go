package repo

import (
	"GoPVZ/internal/pvz/entity"
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pvzRepo struct {
	db *pgxpool.Pool
}

func NewPVZRepo(db *pgxpool.Pool) PVZRepository {
	return &pvzRepo{db: db}
}

func (r *pvzRepo) CreatePVZ(ctx context.Context, pvz *entity.PVZ) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO pvz (id, registration_date, city) VALUES ($1,$2,$3)`,
		pvz.ID, pvz.RegistrationDate, pvz.City,
	)
	return err
}

func (r *pvzRepo) GetById(ctx context.Context, id string) (*entity.PVZ, error) {
	var u entity.PVZ
	err := r.db.QueryRow(ctx,
		`SELECT id, registration_date, city FROM pvz WHERE id=$1`, id,
	).Scan(&u.ID, &u.RegistrationDate, &u.City)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *pvzRepo) CreateReception(ctx context.Context, reception *entity.Reception) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO receptions (id, pvz_id, date_time, status) VALUES ($1,$2,$3,$4)`,
		reception.ID, reception.PvzID, reception.DateTime, reception.Status,
	)
	return err
}

func (r *pvzRepo) CheckPvzsLastReceptionStatusInProgress(ctx context.Context, pvzId string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM receptions WHERE pvz_id=$1 AND status=$2)`, pvzId, entity.StatusInProgress).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *pvzRepo) CreateProduct(ctx context.Context, product *entity.Product) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO products (id, reception_id, date_time, type) VALUES ($1,$2,$3,$4)`,
		product.ID, product.ReceptionID, product.DateTime, product.Type,
	)
	return err
}

func (r *pvzRepo) GetInProgressReceptionIdByPVZId(ctx context.Context, pvzId string) (string, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM receptions WHERE pvz_id=$1 AND status=$2)`, pvzId, entity.StatusInProgress).Scan(&exists)
	if err != nil {
		return "", err
	}
	
	var receptionId string
	err = r.db.QueryRow(ctx, `SELECT id FROM receptions WHERE pvz_id=$1 AND status=$2`, pvzId, entity.StatusInProgress).Scan(&receptionId)
	if err != nil {
		return "", err
	}
	return receptionId, nil
}

func (r *pvzRepo) DeleteLastProductFromReception(ctx context.Context, pvzId string) error {
    // Получаем ID активной приёмки
    receptionId, err := r.GetInProgressReceptionIdByPVZId(ctx, pvzId)
    if err != nil {
        return err
    }

    // Удаляем последний добавленный товар для этой приёмки
    _, err = r.db.Exec(ctx, `
        DELETE FROM products 
        WHERE id = (
            SELECT id FROM products 
            WHERE reception_id = $1 
            ORDER BY date_time DESC 
            LIMIT 1
        )`, receptionId)
    
    return err
}

func (r *pvzRepo) CloseReception(ctx context.Context, pvzId string) (*entity.Reception, error) {
    // Получаем ID активной приёмки
    receptionId, err := r.GetInProgressReceptionIdByPVZId(ctx, pvzId)
    if err != nil {
        return nil, err
    }

    // Обновляем статус приёмки на "close"
    _, err = r.db.Exec(ctx, `
        UPDATE receptions 
        SET status = $1 
        WHERE id = $2 AND status = $3`,
        entity.StatusClose, receptionId, entity.StatusInProgress)
    if err != nil {
        return nil, err
    }

    // Получаем обновлённую запись приёмки
    var reception entity.Reception
    err = r.db.QueryRow(ctx, `
        SELECT id, pvz_id, date_time, status 
        FROM receptions 
        WHERE id = $1`, receptionId).Scan(
        &reception.ID, &reception.PvzID, &reception.DateTime, &reception.Status)
    if err != nil {
        return nil, err
    }

    return &reception, nil
}


func (r *pvzRepo) GetPVZsWithReceptions(ctx context.Context, startDate, endDate *time.Time, limit, offset int) ([]*entity.PVZWithReceptions, error) {
	query := `
		WITH filtered_receptions AS (
			SELECT * FROM receptions
			WHERE ($1::timestamp IS NULL OR date_time >= $1)
			AND ($2::timestamp IS NULL OR date_time <= $2)
		)
		SELECT 
			p.id, p.registration_date, p.city,
			r.id, r.pvz_id, r.date_time, r.status,
			pr.id, pr.reception_id, pr.date_time, pr.type
		FROM pvz p
		LEFT JOIN filtered_receptions r ON p.id = r.pvz_id
		LEFT JOIN products pr ON r.id = pr.reception_id
		ORDER BY p.registration_date DESC, r.date_time DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pvzMap := make(map[uuid.UUID]*entity.PVZWithReceptions)
	receptionMap := make(map[uuid.UUID]*entity.ReceptionWithProducts)

	for rows.Next() {
		var (
			pvz       entity.PVZ
			reception entity.Reception
			product   entity.Product
		)

		err := rows.Scan(
			&pvz.ID, &pvz.RegistrationDate, &pvz.City,
			&reception.ID, &reception.PvzID, &reception.DateTime, &reception.Status,
			&product.ID, &product.ReceptionID, &product.DateTime, &product.Type,
		)
		if err != nil {
			return nil, err
		}

		// Если PVZ еще не в мапе, добавляем
		if _, exists := pvzMap[pvz.ID]; !exists {
			pvzMap[pvz.ID] = &entity.PVZWithReceptions{
				PVZ:        &pvz,
				Receptions: []*entity.ReceptionWithProducts{},
			}
		}

		// Если есть reception и ее еще нет в мапе
		if reception.ID != uuid.Nil {
			if _, exists := receptionMap[reception.ID]; !exists {
				receptionWithProducts := &entity.ReceptionWithProducts{
					Reception: &reception,
					Products:  []*entity.Product{},
				}
				receptionMap[reception.ID] = receptionWithProducts
				pvzMap[pvz.ID].Receptions = append(pvzMap[pvz.ID].Receptions, receptionWithProducts)
			}

			// Если есть product, добавляем к соответствующей reception
			if product.ID != uuid.Nil {
				receptionMap[reception.ID].Products = append(
					receptionMap[reception.ID].Products,
					&product,
				)
			}
		}
	}

	// Преобразуем мапу в слайс
	result := make([]*entity.PVZWithReceptions, 0, len(pvzMap))
	for _, pvzWithReceptions := range pvzMap {
		result = append(result, pvzWithReceptions)
	}

	return result, nil
}