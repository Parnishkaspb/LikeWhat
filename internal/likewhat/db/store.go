package db

import (
	"context"
	"strings"
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
		builder = builder.Where(ilikeExpr(models.ColUserName, filter.NameLike))
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
	// The FK alone is not enough: it would also accept a soft-deleted
	// manufacture, and surface as an opaque 500 instead of a 4xx.
	if _, err := s.GetManufacture(ctx, item.Manufacture.ID); err != nil {
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
		return models.Tobacco{}, NotFound(err)
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
		builder = builder.Where(ilikeExpr(models.ColTobaccoTaste, filter.Taste))
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
	txCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	tx, err := s.pool.Begin(txCtx)
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

// --- Recipe ---

func (s *Store) CreateRecipe(ctx context.Context, item models.Recipe) (models.Recipe, error) {
	// Recipe, mix composition and steps must be created atomically: a recipe
	// without steps or with a half-written mix is never visible.
	txCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	tx, err := s.pool.Begin(txCtx)
	if err != nil {
		return models.Recipe{}, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	query, args, err := pg.Builder.
		Insert(models.TableRecipes).
		Columns(models.ColRecipeUserID, models.ColRecipeTitle).
		Values(item.UserID, item.Title).
		Suffix("RETURNING " + models.ColRecipeID + ", " + models.ColRecipeCreatedAt + ", " + models.ColRecipeUpdatedAt).
		ToSql()
	if err != nil {
		return models.Recipe{}, err
	}

	if err := tx.QueryRow(txCtx, query, args...).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return models.Recipe{}, err
	}

	for _, t := range item.Tobaccos {
		query, args, err := pg.Builder.
			Insert(models.TableRecipeTobacco).
			Columns(models.ColRecipeTobaccoRecipeID, models.ColRecipeTobaccoTobaccoID, models.ColRecipeTobaccoPercent).
			Values(item.ID, t.TobaccoID, t.Percent).
			ToSql()
		if err != nil {
			return models.Recipe{}, err
		}
		if _, err := tx.Exec(txCtx, query, args...); err != nil {
			return models.Recipe{}, err
		}
	}

	for _, st := range item.Steps {
		query, args, err := pg.Builder.
			Insert(models.TableRecipeSteps).
			Columns(models.ColRecipeStepRecipeID, models.ColRecipeStepNumber, models.ColRecipeStepTobaccoID, models.ColRecipeStepWhatDo).
			Values(item.ID, st.StepNumber, nullableString(st.TobaccoID), st.WhatDo).
			ToSql()
		if err != nil {
			return models.Recipe{}, err
		}
		if _, err := tx.Exec(txCtx, query, args...); err != nil {
			return models.Recipe{}, err
		}
	}

	if err := tx.Commit(txCtx); err != nil {
		return models.Recipe{}, err
	}
	return item, nil
}

func (s *Store) GetRecipe(ctx context.Context, id string) (models.Recipe, error) {
	query, args, err := pg.Builder.
		Select(models.ColRecipeID, models.ColRecipeUserID, models.ColRecipeTitle, models.ColRecipeCreatedAt, models.ColRecipeUpdatedAt, models.ColRecipeDeletedAt).
		From(models.TableRecipes).
		Where(squirrel.Eq{models.ColRecipeID: id}).
		Where(squirrel.Eq{models.ColRecipeDeletedAt: nil}).
		ToSql()
	if err != nil {
		return models.Recipe{}, err
	}

	var item models.Recipe
	if err := s.QueryRow(ctx, query, args...).
		Scan(&item.ID, &item.UserID, &item.Title, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
		return models.Recipe{}, NotFound(err)
	}

	if item.Tobaccos, item.Steps, err = s.recipeParts(ctx, item.ID); err != nil {
		return models.Recipe{}, err
	}
	return item, nil
}

func (s *Store) ListRecipes(ctx context.Context, filter models.RecipeFilter) ([]models.Recipe, error) {
	builder := pg.Builder.
		Select(models.ColRecipeID, models.ColRecipeUserID, models.ColRecipeTitle, models.ColRecipeCreatedAt, models.ColRecipeUpdatedAt, models.ColRecipeDeletedAt).
		From(models.TableRecipes).
		Where(squirrel.Eq{models.ColRecipeDeletedAt: nil}).
		OrderBy(models.ColRecipeCreatedAt)

	if filter.UserID != 0 {
		builder = builder.Where(squirrel.Eq{models.ColRecipeUserID: filter.UserID})
	}
	if len(filter.TobaccoIDs) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(filter.TobaccoIDs)), ",")
		tobaccoArgs := make([]any, len(filter.TobaccoIDs))
		for i, id := range filter.TobaccoIDs {
			tobaccoArgs[i] = id
		}
		builder = builder.Where(squirrel.Expr(
			"EXISTS (SELECT 1 FROM "+models.TableRecipeTobacco+
				" WHERE "+models.TableRecipeTobacco+"."+models.ColRecipeTobaccoRecipeID+" = "+models.TableRecipes+"."+models.ColRecipeID+
				" AND "+models.TableRecipeTobacco+"."+models.ColRecipeTobaccoTobaccoID+" IN ("+placeholders+"))",
			tobaccoArgs...,
		))
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

	var items []models.Recipe
	for rows.Next() {
		var item models.Recipe
		if err := rows.Scan(&item.ID, &item.UserID, &item.Title, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range items {
		if items[i].Tobaccos, items[i].Steps, err = s.recipeParts(ctx, items[i].ID); err != nil {
			return nil, err
		}
	}
	return items, nil
}

// recipeParts loads the mix composition and the ordered steps of one recipe.
func (s *Store) recipeParts(ctx context.Context, recipeID string) ([]models.RecipeTobacco, []models.RecipeStep, error) {
	tobaccoQuery, tobaccoArgs, err := pg.Builder.
		Select(models.ColRecipeTobaccoTobaccoID, models.ColRecipeTobaccoPercent).
		From(models.TableRecipeTobacco).
		Where(squirrel.Eq{models.ColRecipeTobaccoRecipeID: recipeID}).
		OrderBy(models.ColRecipeTobaccoTobaccoID).
		ToSql()
	if err != nil {
		return nil, nil, err
	}

	tobaccoRows, err := s.Query(ctx, tobaccoQuery, tobaccoArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer tobaccoRows.Close()

	var tobaccos []models.RecipeTobacco
	for tobaccoRows.Next() {
		var t models.RecipeTobacco
		if err := tobaccoRows.Scan(&t.TobaccoID, &t.Percent); err != nil {
			return nil, nil, err
		}
		tobaccos = append(tobaccos, t)
	}
	if err := tobaccoRows.Err(); err != nil {
		return nil, nil, err
	}

	stepQuery, stepArgs, err := pg.Builder.
		Select(models.ColRecipeStepNumber, models.ColRecipeStepTobaccoID, models.ColRecipeStepWhatDo).
		From(models.TableRecipeSteps).
		Where(squirrel.Eq{models.ColRecipeStepRecipeID: recipeID}).
		OrderBy(models.ColRecipeStepNumber).
		ToSql()
	if err != nil {
		return nil, nil, err
	}

	stepRows, err := s.Query(ctx, stepQuery, stepArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer stepRows.Close()

	var steps []models.RecipeStep
	for stepRows.Next() {
		var st models.RecipeStep
		var tobaccoID *string
		if err := stepRows.Scan(&st.StepNumber, &tobaccoID, &st.WhatDo); err != nil {
			return nil, nil, err
		}
		if tobaccoID != nil {
			st.TobaccoID = *tobaccoID
		}
		steps = append(steps, st)
	}
	if err := stepRows.Err(); err != nil {
		return nil, nil, err
	}
	return tobaccos, steps, nil
}

// nullableString maps an empty string to SQL NULL, so an optional tobacco
// reference stays NULL instead of an empty UUID.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// escapeLike escapes the LIKE wildcard characters of a user-supplied pattern,
// so "%", "_" and "\" in input are matched literally. The expression using the
// result must be built with escapeBackslash().
func escapeLike(pattern string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(pattern)
}

// ilikeExpr builds an ILIKE condition with an explicit escape character,
// because squirrel.Like cannot emit ESCAPE itself.
func ilikeExpr(column, pattern string) squirrel.Sqlizer {
	return squirrel.Expr(column+" ILIKE ? ESCAPE '\\'", "%"+escapeLike(pattern)+"%")
}

// manufacturePredicates translates a list filter into WHERE conditions applied
// uniformly to both the counting and the fetching queries.
func manufacturePredicates(filter models.ManufactureFilter) []squirrel.Sqlizer {
	preds := []squirrel.Sqlizer{squirrel.Eq{models.ColManufactureDeletedAt: nil}}
	if filter.NameLike != "" {
		preds = append(preds, ilikeExpr(models.ColManufactureName, filter.NameLike))
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
