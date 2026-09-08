package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture"
	pg "github.com/Parnishkaspb/LikeWhat/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableManufactures = "manufactures"

// Postgres is a PostgreSQL-backed ManufactureRepository implementation.
type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

var _ ManufactureRepository = (*Postgres)(nil)

func (r *Postgres) Create(ctx context.Context, item manufacture.Manufacture) (manufacture.Manufacture, error) {
	const query = `INSERT INTO manufactures (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at`

	var id string
	var createdAt, updatedAt time.Time
	err := r.pool.QueryRow(ctx, query, item.Name).
		Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return manufacture.Manufacture{}, err
	}

	item.ID = id
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return item, nil
}

func (r *Postgres) Get(ctx context.Context, id string) (manufacture.Manufacture, error) {
	const query = `SELECT id, name, created_at, updated_at, deleted_at
		FROM manufactures
		WHERE id = $1 AND deleted_at IS NULL`

	var item manufacture.Manufacture
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return manufacture.Manufacture{}, ErrNotFound
		}
		return manufacture.Manufacture{}, err
	}
	return item, nil
}

func (r *Postgres) List(ctx context.Context, filter manufacture.ListFilter) ([]manufacture.Manufacture, int, error) {
	where := []pgExpr{
		{"deleted_at IS NULL", nil},
	}
	if filter.NameLike != "" {
		where = append(where, pgExpr{"name ILIKE ?", []any{"%" + filter.NameLike + "%"}})
	}
	if len(filter.IDs) > 0 {
		where = append(where, pgExpr{"id = ANY(?::uuid[])", []any{filter.IDs}})
	}

	countQuery, countArgs := buildCount(where)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery, selectArgs := buildSelect(where, filter.Page, filter.PerPage)
	rows, err := r.pool.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []manufacture.Manufacture
	for rows.Next() {
		var item manufacture.Manufacture
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Postgres) Update(ctx context.Context, item manufacture.Manufacture) (manufacture.Manufacture, error) {
	const query = `UPDATE manufactures
		SET name = $2
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, created_at, updated_at, deleted_at`

	var updated manufacture.Manufacture
	err := r.pool.QueryRow(ctx, query, item.ID, item.Name).
		Scan(&updated.ID, &updated.Name, &updated.CreatedAt, &updated.UpdatedAt, &updated.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return manufacture.Manufacture{}, ErrNotFound
		}
		return manufacture.Manufacture{}, err
	}
	return updated, nil
}

func (r *Postgres) Delete(ctx context.Context, id string) error {
	const query = `UPDATE manufactures
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// pgExpr is a raw SQL fragment with positional placeholders and its arguments.
type pgExpr struct {
	sql  string
	args []any
}

func buildSelect(where []pgExpr, page, perPage uint64) (string, []any) {
	builder := pg.Builder.
		Select("id", "name", "created_at", "updated_at", "deleted_at").
		From(tableManufactures).
		OrderBy("created_at")

	var args []any
	for _, w := range where {
		builder = builder.Where(w.sql, w.args...)
		args = append(args, w.args...)
	}

	if perPage > 0 {
		offset := uint64(0)
		if page > 0 {
			offset = (page - 1) * perPage
		}
		builder = builder.Limit(perPage).Offset(offset)
	}

	query, _, err := builder.ToSql()
	if err != nil {
		return "", nil
	}
	return query, args
}

func buildCount(where []pgExpr) (string, []any) {
	builder := pg.Builder.
		Select("COUNT(*)").
		From(tableManufactures)

	var args []any
	for _, w := range where {
		builder = builder.Where(w.sql, w.args...)
		args = append(args, w.args...)
	}

	query, _, err := builder.ToSql()
	if err != nil {
		return "", nil
	}
	return query, args
}
