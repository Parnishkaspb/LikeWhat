package recipe

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	pg "github.com/Parnishkaspb/LikeWhat/internal/platform/postgres"
)

// Store holds recipe queries and shares the connection pool with the other
// entity stores.
type Store struct {
	*db.Client
}

// NewStore creates a recipe store on top of the shared client.
func NewStore(client *db.Client) *Store {
	return &Store{Client: client}
}

func (s *Store) CreateRecipe(ctx context.Context, item models.Recipe) (models.Recipe, error) {
	// Recipe, mix composition and steps must be created atomically: a recipe
	// without steps or with a half-written mix is never visible.
	txCtx, cancel := context.WithTimeout(ctx, db.QueryTimeout)
	defer cancel()
	tx, err := s.Client.Begin(txCtx)
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
		return models.Recipe{}, db.NotFound(err)
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
