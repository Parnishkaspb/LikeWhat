package db

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	pg "github.com/Parnishkaspb/LikeWhat/internal/platform/postgres"
)

// Store is the single place where queries for every entity live. Repositories
// no longer exist per entity: services call Store methods directly, and all
// SQL is defined here, next to the shared Client.
type Store struct {
	*Client
}

func NewStore(client *Client) *Store {
	return &Store{Client: client}
}

// --- User ---

func (s *Store) CreateUser(ctx context.Context, item models.User) (models.User, error) {
	query, args, err := pg.Builder.
		Insert(models.TableUsers).
		Columns(models.ColUserTelegramID, models.ColUserNickName, models.ColUserName).
		Values(item.TelegramID, item.NickName, item.Name).
		Suffix("RETURNING " + models.ColUserID).
		ToSql()
	if err != nil {
		return models.User{}, err
	}

	if err := s.QueryRow(ctx, query, args...).Scan(&item.ID); err != nil {
		return models.User{}, err
	}
	return item, nil
}

func (s *Store) GetUser(ctx context.Context, id int64) (models.User, error) {
	query, args, err := pg.Builder.
		Select(models.ColUserID, models.ColUserTelegramID, models.ColUserNickName, models.ColUserName).
		From(models.TableUsers).
		Where(squirrel.Eq{models.ColUserID: id}).
		ToSql()
	if err != nil {
		return models.User{}, err
	}

	var item models.User
	if err := s.QueryRow(ctx, query, args...).
		Scan(&item.ID, &item.TelegramID, &item.NickName, &item.Name); err != nil {
		return models.User{}, NotFound(err)
	}
	return item, nil
}

func (s *Store) ListUsers(ctx context.Context, filter models.UserFilter) ([]models.User, error) {
	builder := pg.Builder.
		Select(models.ColUserID, models.ColUserTelegramID, models.ColUserNickName, models.ColUserName).
		From(models.TableUsers).
		OrderBy(models.ColUserID)

	if filter.NameLike != "" {
		builder = builder.Where(squirrel.ILike{models.ColUserName: "%" + filter.NameLike + "%"})
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

	var items []models.User
	for rows.Next() {
		var item models.User
		if err := rows.Scan(&item.ID, &item.TelegramID, &item.NickName, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// --- Tobacco ---

func (s *Store) CreateTobacco(ctx context.Context, item models.Tobacco) (models.Tobacco, error) {
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
	query, args, err := pg.Builder.
		Select(models.ColTobaccoID, models.ColTobaccoTaste, models.ColTobaccoPhoto, models.ColTobaccoManufactureID, models.ColTobaccoCreatedAt, models.ColTobaccoUpdatedAt, models.ColTobaccoDeletedAt).
		From(models.TableTobaccos).
		Where(squirrel.Eq{models.ColTobaccoID: id}).
		Where(squirrel.Eq{models.ColTobaccoDeletedAt: nil}).
		ToSql()
	if err != nil {
		return models.Tobacco{}, err
	}

	var item models.Tobacco
	var photo *string
	var deletedAt *time.Time
	if err := s.QueryRow(ctx, query, args...).
		Scan(&item.ID, &item.Taste, &photo, &item.Manufacture.ID, &item.CreatedAt, &item.UpdatedAt, &deletedAt); err != nil {
		return models.Tobacco{}, NotFound(err)
	}

	if photo != nil {
		item.Photo = *photo
	}
	item.DeletedAt = deletedAt
	return item, nil
}

func (s *Store) ListTobaccos(ctx context.Context, filter models.TobaccoFilter) ([]models.Tobacco, error) {
	builder := pg.Builder.
		Select(models.ColTobaccoID, models.ColTobaccoTaste, models.ColTobaccoPhoto, models.ColTobaccoManufactureID, models.ColTobaccoCreatedAt, models.ColTobaccoUpdatedAt, models.ColTobaccoDeletedAt).
		From(models.TableTobaccos).
		Where(squirrel.Eq{models.ColTobaccoDeletedAt: nil}).
		OrderBy(models.ColTobaccoCreatedAt)

	if filter.Taste != "" {
		builder = builder.Where(squirrel.ILike{models.ColTobaccoTaste: "%" + filter.Taste + "%"})
	}
	if len(filter.ManufactureIDs) > 0 {
		builder = builder.Where(squirrel.Eq{models.ColTobaccoManufactureID: filter.ManufactureIDs})
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

// --- Manufacture ---

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
		return models.Manufacture{}, NotFound(err)
	}
	return item, nil
}

func (s *Store) ListManufactures(ctx context.Context, filter models.ManufactureFilter) ([]models.Manufacture, int, error) {
	preds := manufacturePredicates(filter)

	var total int
	countQuery, countArgs, err := buildManufactureCount(preds)
	if err != nil {
		return nil, 0, err
	}
	if err := s.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery, selectArgs, err := buildManufactureSelect(preds, filter.Page, filter.PerPage)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.Query(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

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
		return models.Manufacture{}, NotFound(err)
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
		return ErrNotFound
	}
	return nil
}

// manufacturePredicates translates a list filter into WHERE conditions applied
// uniformly to both the counting and the fetching queries.
func manufacturePredicates(filter models.ManufactureFilter) []squirrel.Sqlizer {
	preds := []squirrel.Sqlizer{squirrel.Eq{models.ColManufactureDeletedAt: nil}}
	if filter.NameLike != "" {
		preds = append(preds, squirrel.ILike{models.ColManufactureName: "%" + filter.NameLike + "%"})
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
