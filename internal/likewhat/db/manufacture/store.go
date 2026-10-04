package manufacture

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	pg "github.com/Parnishkaspb/LikeWhat/internal/platform/postgres"
)

// Store holds manufacture queries and shares the connection pool with the
// other entity stores.
type Store struct {
	*db.Client
}

// NewStore creates a manufacture store on top of the shared client.
func NewStore(client *db.Client) *Store {
	return &Store{Client: client}
}

func (s *Store) CreateManufacture(ctx context.Context, item models.Manufacture) (models.Manufacture, error) {
	query, args, err := pg.Builder.
		Insert(models.TableManufactures).
		Columns(models.ColManufactureName).
		Values(item.Name).
		Suffix("RETURNING " + models.ColManufactureID + ", " + models.ColManufactureCreatedAt + ", " + models.ColManufactureUpdatedAt).
		ToSql()
	if err != nil {
		return models.Manufacture{}, err
	}

	var id string
	var createdAt, updatedAt time.Time
	if err := s.QueryRow(ctx, query, args...).
		Scan(&id, &createdAt, &updatedAt); err != nil {
		return models.Manufacture{}, err
	}

	item.ID = id
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return item, nil
}

func (s *Store) GetManufacture(ctx context.Context, id string) (models.Manufacture, error) {
	query, args, err := pg.Builder.
		Select(models.ColManufactureID, models.ColManufactureName, models.ColManufactureCreatedAt, models.ColManufactureUpdatedAt, models.ColManufactureDeletedAt).
		From(models.TableManufactures).
		Where(squirrel.Eq{models.ColManufactureID: id}).
		Where(squirrel.Eq{models.ColManufactureDeletedAt: nil}).
		ToSql()
	if err != nil {
		return models.Manufacture{}, err
	}

	var item models.Manufacture
	if err := s.QueryRow(ctx, query, args...).
		Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
		return models.Manufacture{}, db.NotFound(err)
	}
	return item, nil
}

func (s *Store) ListManufactures(ctx context.Context, filter models.ManufactureFilter) ([]models.Manufacture, int, error) {
	preds := manufacturePredicates(filter)

	countQuery, countArgs, err := buildManufactureCount(preds)
	if err != nil {
		return nil, 0, err
	}
	selectQuery, selectArgs, err := buildManufactureSelect(preds, filter.Page, filter.PerPage)
	if err != nil {
		return nil, 0, err
	}

	// Count and page run in one read-only transaction so total_count and the
	// returned page come from a single snapshot.
	txCtx, cancel := context.WithTimeout(ctx, db.QueryTimeout)
	defer cancel()
	tx, err := s.Client.Begin(txCtx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var total int
	if err := tx.QueryRow(txCtx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := tx.Query(txCtx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}

	var items []models.Manufacture
	for rows.Next() {
		var item models.Manufacture
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	rows.Close()

	if err := tx.Commit(txCtx); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Store) UpdateManufacture(ctx context.Context, item models.Manufacture) (models.Manufacture, error) {
	query, args, err := pg.Builder.
		Update(models.TableManufactures).
		Set(models.ColManufactureName, item.Name).
		Where(squirrel.Eq{models.ColManufactureID: item.ID}).
		Where(squirrel.Eq{models.ColManufactureDeletedAt: nil}).
		Suffix("RETURNING " + models.ColManufactureID + ", " + models.ColManufactureName + ", " + models.ColManufactureCreatedAt + ", " + models.ColManufactureUpdatedAt + ", " + models.ColManufactureDeletedAt).
		ToSql()
	if err != nil {
		return models.Manufacture{}, err
	}

	var updated models.Manufacture
	if err := s.QueryRow(ctx, query, args...).
		Scan(&updated.ID, &updated.Name, &updated.CreatedAt, &updated.UpdatedAt, &updated.DeletedAt); err != nil {
		return models.Manufacture{}, db.NotFound(err)
	}
	return updated, nil
}

func (s *Store) DeleteManufacture(ctx context.Context, id string) error {
	query, args, err := pg.Builder.
		Update(models.TableManufactures).
		Set(models.ColManufactureDeletedAt, squirrel.Expr("now()")).
		Where(squirrel.Eq{models.ColManufactureID: id}).
		Where(squirrel.Eq{models.ColManufactureDeletedAt: nil}).
		ToSql()
	if err != nil {
		return err
	}

	tag, err := s.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return db.ErrNotFound
	}
	return nil
}

// manufacturePredicates translates a list filter into WHERE conditions applied
// uniformly to both the counting and the fetching queries.
func manufacturePredicates(filter models.ManufactureFilter) []squirrel.Sqlizer {
	preds := []squirrel.Sqlizer{squirrel.Eq{models.ColManufactureDeletedAt: nil}}
	if filter.NameLike != "" {
		preds = append(preds, db.ILikeExpr(models.ColManufactureName, filter.NameLike))
	}
	if len(filter.IDs) > 0 {
		preds = append(preds, squirrel.Eq{models.ColManufactureID: filter.IDs})
	}
	return preds
}

func applyManufacturePredicates(builder squirrel.SelectBuilder, preds []squirrel.Sqlizer) squirrel.SelectBuilder {
	for _, p := range preds {
		builder = builder.Where(p)
	}
	return builder
}

func buildManufactureSelect(preds []squirrel.Sqlizer, page, perPage uint64) (string, []any, error) {
	builder := applyManufacturePredicates(
		pg.Builder.
			Select(models.ColManufactureID, models.ColManufactureName, models.ColManufactureCreatedAt, models.ColManufactureUpdatedAt, models.ColManufactureDeletedAt).
			From(models.TableManufactures).
			OrderBy(models.ColManufactureCreatedAt),
		preds,
	)

	if perPage > 0 {
		offset := uint64(0)
		if page > 0 {
			offset = (page - 1) * perPage
		}
		builder = builder.Limit(perPage).Offset(offset)
	}

	return builder.ToSql()
}

func buildManufactureCount(preds []squirrel.Sqlizer) (string, []any, error) {
	builder := applyManufacturePredicates(
		pg.Builder.Select("COUNT(*)").From(models.TableManufactures),
		preds,
	)
	return builder.ToSql()
}
