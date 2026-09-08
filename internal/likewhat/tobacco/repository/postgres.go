package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco"
	pg "github.com/Parnishkaspb/LikeWhat/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableTobaccos = "tobaccos"

// Postgres is a PostgreSQL-backed TobaccoRepository implementation.
type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

var _ TobaccoRepository = (*Postgres)(nil)

func (r *Postgres) Create(ctx context.Context, item tobacco.Tobacco) (tobacco.Tobacco, error) {
	const query = `INSERT INTO tobaccos (taste, photo, manufacture_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	var id string
	var createdAt, updatedAt time.Time
	err := r.pool.QueryRow(ctx, query, item.Taste, item.Photo, item.Manufacture.ID).
		Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return tobacco.Tobacco{}, err
	}

	item.ID = id
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return item, nil
}

func (r *Postgres) Get(ctx context.Context, id string) (tobacco.Tobacco, error) {
	const query = `SELECT id, taste, photo, manufacture_id, created_at, updated_at, deleted_at
		FROM tobaccos
		WHERE id = $1 AND deleted_at IS NULL`

	var item tobacco.Tobacco
	var photo *string
	var deletedAt *time.Time
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&item.ID, &item.Taste, &photo, &item.Manufacture.ID, &item.CreatedAt, &item.UpdatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tobacco.Tobacco{}, ErrNotFound
		}
		return tobacco.Tobacco{}, err
	}

	if photo != nil {
		item.Photo = *photo
	}
	item.DeletedAt = deletedAt
	return item, nil
}

func (r *Postgres) List(ctx context.Context, filter tobacco.ListFilter) ([]tobacco.Tobacco, error) {
	builder := pg.Builder.
		Select("id", "taste", "photo", "manufacture_id", "created_at", "updated_at", "deleted_at").
		From(tableTobaccos).
		Where("deleted_at IS NULL").
		OrderBy("created_at")

	if filter.Taste != "" {
		builder = builder.Where("taste ILIKE ?", "%"+filter.Taste+"%")
	}
	if len(filter.ManufactureIDs) > 0 {
		builder = builder.Where("manufacture_id = ANY(?::uuid[])", filter.ManufactureIDs)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []tobacco.Tobacco
	for rows.Next() {
		var item tobacco.Tobacco
		var photo *string
		var deletedAt *time.Time
		if err := rows.Scan(&item.ID, &item.Taste, &photo, &item.Manufacture.ID, &item.CreatedAt, &item.UpdatedAt, &deletedAt); err != nil {
			return nil, err
		}
		if photo != nil {
			item.Photo = *photo
		}
		item.DeletedAt = deletedAt
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
