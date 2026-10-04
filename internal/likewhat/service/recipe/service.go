package service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
)

// ErrNotFound is returned when a recipe does not exist.
var ErrNotFound = db.ErrNotFound

// RecipeTobaccoInput is one tobacco of the mix accepted on creation.
type RecipeTobaccoInput struct {
	TobaccoID string
	Percent   float64
}

// RecipeStepInput is one preparation step accepted on creation.
type RecipeStepInput struct {
	StepNumber int
	TobaccoID  string
	WhatDo     string
}

// CreateInput contains the business fields accepted when a recipe is created.
type CreateInput struct {
	UserID   int64
	Title    string
	Tobaccos []RecipeTobaccoInput
	Steps    []RecipeStepInput
}

// ListInput contains optional search criteria.
type ListInput struct {
	UserID     int64
	TobaccoIDs []string
}

// RecipeService contains recipe use cases and is independent of transports.
// Input validation happens at the transport layer; this service enforces the
// cross-field invariants a transport cannot know about.
type RecipeService struct {
	store *db.Store
}

func NewRecipeService(store *db.Store) *RecipeService {
	return &RecipeService{store: store}
}

func (s *RecipeService) Create(ctx context.Context, input CreateInput) (models.Recipe, error) {
	title := strings.TrimSpace(input.Title)

	tobaccos := make([]models.RecipeTobacco, 0, len(input.Tobaccos))
	tobaccoIDs := make([]string, 0, len(input.Tobaccos))
	total := 0.0
	for _, t := range input.Tobaccos {
		id := strings.TrimSpace(t.TobaccoID)
		if id == "" {
			return models.Recipe{}, fmt.Errorf("recipe tobacco: tobacco_id is required")
		}
		if t.Percent <= 0 || t.Percent > 100 {
			return models.Recipe{}, fmt.Errorf("recipe tobacco %q: percent must be in (0, 100]", id)
		}
		if slices.Contains(tobaccoIDs, id) {
			return models.Recipe{}, fmt.Errorf("recipe tobacco %q: duplicate", id)
		}
		total += t.Percent
		tobaccoIDs = append(tobaccoIDs, id)
		tobaccos = append(tobaccos, models.RecipeTobacco{TobaccoID: id, Percent: t.Percent})
	}
	if len(tobaccos) > 0 && (total < 99.99 || total > 100.01) {
		return models.Recipe{}, fmt.Errorf("recipe tobaccos: percents must sum to 100, got %.2f", total)
	}
	if len(tobaccos) == 0 {
		return models.Recipe{}, fmt.Errorf("recipe: at least one tobacco is required")
	}

	steps := make([]models.RecipeStep, 0, len(input.Steps))
	if len(input.Steps) == 0 {
		return models.Recipe{}, fmt.Errorf("recipe: at least one step is required")
	}
	for i, st := range input.Steps {
		number := st.StepNumber
		if number == 0 {
			// Numbering may be omitted: steps are applied in request order.
			number = i + 1
		}
		if number != i+1 {
			return models.Recipe{}, fmt.Errorf("recipe step %d: steps must be numbered from 1 without gaps", number)
		}
		whatDo := strings.TrimSpace(st.WhatDo)
		if whatDo == "" {
			return models.Recipe{}, fmt.Errorf("recipe step %d: what_do is required", number)
		}
		tobaccoID := strings.TrimSpace(st.TobaccoID)
		if tobaccoID != "" && !slices.Contains(tobaccoIDs, tobaccoID) {
			return models.Recipe{}, fmt.Errorf("recipe step %d: tobacco %q is not in the mix", number, tobaccoID)
		}
		steps = append(steps, models.RecipeStep{
			StepNumber: number,
			TobaccoID:  tobaccoID,
			WhatDo:     whatDo,
		})
	}

	// Every tobacco of the mix and every referenced user must exist.
	if _, err := s.store.GetUser(ctx, input.UserID); err != nil {
		return models.Recipe{}, fmt.Errorf("recipe user %d: %w", input.UserID, err)
	}
	for _, id := range tobaccoIDs {
		if _, err := s.store.GetTobacco(ctx, id); err != nil {
			return models.Recipe{}, fmt.Errorf("recipe tobacco %q: %w", id, err)
		}
	}

	return s.store.CreateRecipe(ctx, models.Recipe{
		UserID:   input.UserID,
		Title:    title,
		Tobaccos: tobaccos,
		Steps:    steps,
	})
}

func (s *RecipeService) Get(ctx context.Context, id string) (models.Recipe, error) {
	return s.store.GetRecipe(ctx, strings.TrimSpace(id))
}

func (s *RecipeService) List(ctx context.Context, input ListInput) ([]models.Recipe, error) {
	return s.store.ListRecipes(ctx, models.RecipeFilter{
		UserID:     input.UserID,
		TobaccoIDs: normalizeIDs(input.TobaccoIDs),
	})
}

// normalizeIDs trims and de-duplicates an identifier list.
func normalizeIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
