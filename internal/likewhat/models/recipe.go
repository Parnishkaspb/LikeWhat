package models

import "time"

// Table and column names of the recipes, recipe_tobaccos and recipe_steps
// tables. They live next to the domain model so repositories share a single
// source of truth for SQL fragments.
const (
	TableRecipes       = "recipes"
	TableRecipeTobacco = "recipe_tobaccos"
	TableRecipeSteps   = "recipe_steps"

	ColRecipeID        = "id"
	ColRecipeUserID    = "user_id"
	ColRecipeTitle     = "title"
	ColRecipeCreatedAt = "created_at"
	ColRecipeUpdatedAt = "updated_at"
	ColRecipeDeletedAt = "deleted_at"

	ColRecipeTobaccoRecipeID  = "recipe_id"
	ColRecipeTobaccoTobaccoID = "tobacco_id"
	ColRecipeTobaccoPercent   = "percent"

	ColRecipeStepID        = "id"
	ColRecipeStepRecipeID  = "recipe_id"
	ColRecipeStepNumber    = "step_number"
	ColRecipeStepTobaccoID = "tobacco_id"
	ColRecipeStepWhatDo    = "what_do"
)

// Recipe is the domain model of a hookah mix recipe. It deliberately does not
// depend on protobuf.
type Recipe struct {
	ID        string
	UserID    int64
	Title     string
	Tobaccos  []RecipeTobacco
	Steps     []RecipeStep
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// RecipeTobacco is one tobacco of the mix with its percentage.
type RecipeTobacco struct {
	TobaccoID string
	Percent   float64
}

// RecipeStep is one preparation step. TobaccoID is empty when the step
// concerns the whole mix rather than a single tobacco.
type RecipeStep struct {
	StepNumber int
	TobaccoID  string
	WhatDo     string
}

// RecipeFilter limits recipe records returned by a repository.
type RecipeFilter struct {
	UserID     int64
	TobaccoIDs []string
}
