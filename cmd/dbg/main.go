// Temporary diagnostic tool: prints the real errors behind the gRPC
// "internal server error" responses. Run with DATABASE_URL pointing at a
// reachable postgres, then delete this directory.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	recipestore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/recipe"
	tobaccostore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/tobacco"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println("connect:", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := tobaccostore.NewStore(db.NewClient(pool))
	recipes := recipestore.NewStore(db.NewClient(pool))
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
