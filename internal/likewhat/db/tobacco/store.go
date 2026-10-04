package tobacco

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	manufacturestore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/manufacture"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	pg "github.com/Parnishkaspb/LikeWhat/internal/platform/postgres"
)

// Store holds tobacco queries and shares the connection pool with the other
// entity stores.
type Store struct {
	*db.Client
	manufactures *manufacturestore.Store
}

// NewStore creates a tobacco store on top of the shared client.
func NewStore(client *db.Client) *Store {
	return &Store{Client: client, manufactures: manufacturestore.NewStore(client)}
}

func (s *Store) CreateTobacco(ctx context.Context, item models.Tobacco) (models.Tobacco, error) {
	// The FK alone is not enough: it would also accept a soft-deleted
	// manufacture, and surface as an opaque 500 instead of a 4xx.
	if _, err := s.manufactures.GetManufacture(ctx, item.Manufacture.ID); err != nil {
		return models.Tobacco{}, err
	}

	query, args, err := pg.Builder.
		Insert(models.TableTobaccos).
		Columns(models.ColTobaccoTaste, models.ColTobaccoPhoto, models.ColTobaccoManufactureID).
		Values(item.Taste, item.Photo, item.Manufacture.ID).
		Suffix("RETURNING " + models.ColTobaccoID + ", " + models.ColTobaccoCreatedAt + ", " + models.ColTobaccoUpdatedAt).
		ToSql()
	if err != nil {
		return models.Tobacco{}, err
	}

	var id string
	var createdAt, updatedAt time.Time
	if err := s.QueryRow(ctx, query, args...).
		Scan(&id, &createdAt, &updatedAt); err != nil {
		return models.Tobacco{}, err
	}

	item.ID = id
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return item, nil
}

func (s *Store) GetTobacco(ctx context.Context, id string) (models.Tobacco, error) {
	// Columns are qualified with the table name: the query joins manufactures,
	// whose id/created_at/updated_at would otherwise be ambiguous.
	query, args, err := pg.Builder.
		Select(
			models.TableTobaccos+"."+models.ColTobaccoID, models.TableTobaccos+"."+models.ColTobaccoTaste, models.TableTobaccos+"."+models.ColTobaccoPhoto,
			"m."+models.ColManufactureID, "m."+models.ColManufactureName,
			models.TableTobaccos+"."+models.ColTobaccoCreatedAt, models.TableTobaccos+"."+models.ColTobaccoUpdatedAt, models.TableTobaccos+"."+models.ColTobaccoDeletedAt,
		).
		From(models.TableTobaccos).
		Join(models.TableManufactures + " m ON m." + models.ColManufactureID + " = " + models.TableTobaccos + "." + models.ColTobaccoManufactureID).
		Where(squirrel.Eq{models.TableTobaccos + "." + models.ColTobaccoID: id}).
		Where(squirrel.Eq{models.TableTobaccos + "." + models.ColTobaccoDeletedAt: nil}).
		ToSql()
	if err != nil {
		return models.Tobacco{}, err
	}

	var item models.Tobacco
	var photo *string
	var deletedAt *time.Time
	if err := s.QueryRow(ctx, query, args...).
		Scan(&item.ID, &item.Taste, &photo, &item.Manufacture.ID, &item.Manufacture.Name, &item.CreatedAt, &item.UpdatedAt, &deletedAt); err != nil {
		return models.Tobacco{}, db.NotFound(err)
	}

	if photo != nil {
		item.Photo = *photo
	}
	item.DeletedAt = deletedAt
	return item, nil
}

func (s *Store) ListTobaccos(ctx context.Context, filter models.TobaccoFilter) ([]models.Tobacco, error) {
	// Columns are qualified with the table name: the query joins manufactures,
	// whose id/created_at/updated_at would otherwise be ambiguous.
	builder := pg.Builder.
		Select(
			models.TableTobaccos+"."+models.ColTobaccoID, models.TableTobaccos+"."+models.ColTobaccoTaste, models.TableTobaccos+"."+models.ColTobaccoPhoto,
			"m."+models.ColManufactureID, "m."+models.ColManufactureName,
			models.TableTobaccos+"."+models.ColTobaccoCreatedAt, models.TableTobaccos+"."+models.ColTobaccoUpdatedAt, models.TableTobaccos+"."+models.ColTobaccoDeletedAt,
		).
		From(models.TableTobaccos).
		Join(models.TableManufactures + " m ON m." + models.ColManufactureID + " = " + models.TableTobaccos + "." + models.ColTobaccoManufactureID).
		Where(squirrel.Eq{models.TableTobaccos + "." + models.ColTobaccoDeletedAt: nil}).
		OrderBy(models.TableTobaccos + "." + models.ColTobaccoCreatedAt)

	if filter.Taste != "" {
		builder = builder.Where(db.ILikeExpr(models.ColTobaccoTaste, filter.Taste))
	}
	if len(filter.ManufactureIDs) > 0 {
		builder = builder.Where(squirrel.Eq{models.TableTobaccos + "." + models.ColTobaccoManufactureID: filter.ManufactureIDs})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Tobacco
	for rows.Next() {
		var item models.Tobacco
		var photo *string
		var deletedAt *time.Time
		if err := rows.Scan(&item.ID, &item.Taste, &photo, &item.Manufacture.ID, &item.Manufacture.Name, &item.CreatedAt, &item.UpdatedAt, &deletedAt); err != nil {
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
