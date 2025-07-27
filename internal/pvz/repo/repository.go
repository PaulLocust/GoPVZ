package repo

import (
	"GoPVZ/internal/pvz/entity"
	"context"
	"sort"
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
	var receptionId string
	err := r.db.QueryRow(ctx, `SELECT id FROM receptions WHERE pvz_id=$1 AND status=$2`, pvzId, entity.StatusInProgress).Scan(&receptionId)
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
        SELECT 
            p.id AS pvz_id, 
            p.registration_date AS pvz_registration_date, 
            p.city AS pvz_city,
            r.id AS reception_id, 
            r.date_time AS reception_date, 
            r.status AS reception_status,
            pr.id AS product_id, 
            pr.date_time AS product_date, 
            pr.type AS product_type
        FROM pvz p
        LEFT JOIN receptions r ON p.id = r.pvz_id
            AND ($1::timestamp IS NULL OR r.date_time >= $1)
            AND ($2::timestamp IS NULL OR r.date_time <= $2)
        LEFT JOIN products pr ON r.id = pr.reception_id
        ORDER BY p.registration_date DESC, r.date_time DESC, pr.date_time DESC
        LIMIT $3 OFFSET $4
    `

    rows, err := r.db.Query(ctx, query, startDate, endDate, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    pvzMap := make(map[uuid.UUID]*entity.PVZWithReceptions)
    receptionMap := make(map[uuid.UUID]*entity.ReceptionWithProducts)

    for rows.Next() {
        var (
            pvzID           uuid.UUID
            pvzRegDate      time.Time
            pvzCity         string
            receptionID     *uuid.UUID
            receptionDate   *time.Time
            receptionStatus *string
            productID       *uuid.UUID
            productDate     *time.Time
            productType     *string
        )

        err := rows.Scan(
            &pvzID, &pvzRegDate, &pvzCity,
            &receptionID, &receptionDate, &receptionStatus,
            &productID, &productDate, &productType,
        )
        if err != nil {
            return nil, err
        }

        // Обработка PVZ
        if _, exists := pvzMap[pvzID]; !exists {
            pvzMap[pvzID] = &entity.PVZWithReceptions{
                PVZ: &entity.PVZ{
                    ID:               pvzID,
                    RegistrationDate: pvzRegDate,
                    City:             entity.City(pvzCity),
                },
                Receptions: []*entity.ReceptionWithProducts{},
            }
        }

        // Обработка Reception
        if receptionID != nil {
            if _, exists := receptionMap[*receptionID]; !exists {
                reception := &entity.Reception{
                    ID:       *receptionID,
                    PvzID:    pvzID,
                    DateTime: *receptionDate,
                    Status:   entity.Status(*receptionStatus),
                }
                receptionWithProducts := &entity.ReceptionWithProducts{
                    Reception: reception,
                    Products:  []*entity.Product{},
                }
                receptionMap[*receptionID] = receptionWithProducts
                pvzMap[pvzID].Receptions = append(pvzMap[pvzID].Receptions, receptionWithProducts)
            }

            // Обработка Product
            if productID != nil {
                product := &entity.Product{
                    ID:          *productID,
                    ReceptionID: *receptionID,
                    DateTime:    *productDate,
                    Type:        entity.Type(*productType),
                }
                receptionMap[*receptionID].Products = append(receptionMap[*receptionID].Products, product)
            }
        }
    }

    // Преобразование map в slice с сохранением порядка
    result := make([]*entity.PVZWithReceptions, 0, len(pvzMap))
    for _, pvz := range pvzMap {
        // Сортируем приемки по дате (DESC)
        sort.Slice(pvz.Receptions, func(i, j int) bool {
            return pvz.Receptions[i].Reception.DateTime.After(pvz.Receptions[j].Reception.DateTime)
        })
        
        // Сортируем товары внутри каждой приемки по дате (DESC)
        for _, rec := range pvz.Receptions {
            sort.Slice(rec.Products, func(i, j int) bool {
                return rec.Products[i].DateTime.After(rec.Products[j].DateTime)
            })
        }
        
        result = append(result, pvz)
    }

    // Сортируем ПВЗ по дате регистрации (DESC)
    sort.Slice(result, func(i, j int) bool {
        return result[i].PVZ.RegistrationDate.After(result[j].PVZ.RegistrationDate)
    })

    return result, nil
}