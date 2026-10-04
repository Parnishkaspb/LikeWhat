package db_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	recipestore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/recipe"
	tobaccostore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/tobacco"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDebugE2E is a temporary diagnostic: it connects to the e2e postgres and
// prints the real error behind the gRPC "internal server error" responses.
func TestDebugE2E(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	client := db.NewClient(pool)
	store := tobaccostore.NewStore(client)
	recipes := recipestore.NewStore(client)
	ctx := context.Background()

	tob, listErr := store.ListTobaccos(ctx, models.TobaccoFilter{})
	fmt.Printf("ListTobaccos: err=%v items=%d\n", listErr, len(tob))

	if listErr == nil && len(tob) > 0 {
		got, getErr := store.GetTobacco(ctx, tob[0].ID)
		fmt.Printf("GetTobacco: err=%v item=%+v\n", getErr, got)
	}

	recipe, recipeErr := recipes.CreateRecipe(ctx, models.Recipe{
		UserID: 1,
		Title:  "debug",
		Tobaccos: []models.RecipeTobacco{{
			TobaccoID: "f13f6c72-3d72-4f02-8872-9873a8a8e801",
			Percent:   100,
		}},
		Steps: []models.RecipeStep{{StepNumber: 1, WhatDo: "pack"}},
	})
	fmt.Printf("CreateRecipe: err=%v recipe=%+v\n", recipeErr, recipe)
}
