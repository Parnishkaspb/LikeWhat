package service

import (
	"context"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	recipestore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/recipe"
	tobaccostore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/tobacco"
	userstore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/user"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newSvc(t *testing.T) *RecipeService {
	t.Helper()
	client := db.NewClient(testpostgres.Pool())
	return NewRecipeService(recipestore.NewStore(client), userstore.NewStore(client), tobaccostore.NewStore(client))
}

// createUser inserts a user row directly, because a recipe requires a user
// foreign key.
func createUser(t *testing.T) int64 {
	t.Helper()
	var id int64
	err := testpostgres.Pool().QueryRow(context.Background(),
		`INSERT INTO users (telegram_id, nick_name, name) VALUES ($1, $2, $3) RETURNING id`,
		int64(100500), "@tester", "Tester").Scan(&id)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

// createTobacco inserts a tobacco row directly, because a recipe mix requires
// tobacco foreign keys.
func createTobacco(t *testing.T, taste string) string {
	t.Helper()
	var manufactureID, id string
	err := testpostgres.Pool().QueryRow(context.Background(),
		`INSERT INTO manufactures (name) VALUES ($1) RETURNING id`, "Ozon").Scan(&manufactureID)
	if err != nil {
		t.Fatalf("insert manufacture: %v", err)
	}
	err = testpostgres.Pool().QueryRow(context.Background(),
		`INSERT INTO tobaccos (taste, manufacture_id) VALUES ($1, $2) RETURNING id`, taste, manufactureID).Scan(&id)
	if err != nil {
		t.Fatalf("insert tobacco: %v", err)
	}
	return id
}

func TestCreateStoresFullRecipe(t *testing.T) {
	svc := newSvc(t)
	userID := createUser(t)
	cherry := createTobacco(t, "BlackBurn Cherry")
	choco := createTobacco(t, "Dark Side Chocolate")

	created, err := svc.Create(context.Background(), CreateInput{
		UserID: userID,
		Title:  "  Шоколадная вишня  ",
		Tobaccos: []RecipeTobaccoInput{
			{TobaccoID: cherry, Percent: 60},
			{TobaccoID: choco, Percent: 40},
		},
		Steps: []RecipeStepInput{
			{WhatDo: "Растопите шоколадку"},
			{TobaccoID: cherry, WhatDo: "Зарядите табак в чашу"},
			{WhatDo: "Полейте шоколадом"},
		},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" || created.Title != "Шоколадная вишня" || created.UserID != userID {
		t.Fatalf("Create() = %+v, want trimmed title and user", created)
	}
	if len(created.Tobaccos) != 2 {
		t.Fatalf("Create() tobaccos = %d, want 2", len(created.Tobaccos))
	}
	if len(created.Steps) != 3 || created.Steps[0].StepNumber != 1 || created.Steps[2].StepNumber != 3 {
		t.Fatalf("Create() steps = %+v, want numbered 1..3", created.Steps)
	}

	got, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Title != created.Title || len(got.Tobaccos) != 2 || len(got.Steps) != 3 {
		t.Fatalf("Get() = %+v, want the full recipe back", got)
	}
	if got.Steps[1].TobaccoID != cherry {
		t.Fatalf("Get() step 2 tobacco = %q, want %q", got.Steps[1].TobaccoID, cherry)
	}
}

func TestCreateRejectsPercentsNotSummingTo100(t *testing.T) {
	svc := newSvc(t)
	userID := createUser(t)
	cherry := createTobacco(t, "Cherry")
	choco := createTobacco(t, "Chocolate")

	_, err := svc.Create(context.Background(), CreateInput{
		UserID: userID,
		Title:  "Плохая смесь",
		Tobaccos: []RecipeTobaccoInput{
			{TobaccoID: cherry, Percent: 60},
			{TobaccoID: choco, Percent: 30},
		},
		Steps: []RecipeStepInput{{WhatDo: "Смешать"}},
	})
	if err == nil {
		t.Fatalf("Create() with percents summing to 90 must fail")
	}
}

func TestCreateRejectsUnknownStepTobacco(t *testing.T) {
	svc := newSvc(t)
	userID := createUser(t)
	cherry := createTobacco(t, "Cherry")
	alien := createTobacco(t, "Alien")

	_, err := svc.Create(context.Background(), CreateInput{
		UserID: userID,
		Title:  "Чужой табак",
		Tobaccos: []RecipeTobaccoInput{
			{TobaccoID: cherry, Percent: 100},
		},
		Steps: []RecipeStepInput{
			{TobaccoID: alien, WhatDo: "Зарядите чужой табак"},
		},
	})
	if err == nil {
		t.Fatalf("Create() with a step tobacco outside the mix must fail")
	}
}

func TestCreateRejectsMissingUserOrTobacco(t *testing.T) {
	svc := newSvc(t)
	userID := createUser(t)
	cherry := createTobacco(t, "Cherry")

	if _, err := svc.Create(context.Background(), CreateInput{
		UserID:   424242,
		Title:    "Нет пользователя",
		Tobaccos: []RecipeTobaccoInput{{TobaccoID: cherry, Percent: 100}},
		Steps:    []RecipeStepInput{{WhatDo: "Что-то"}},
	}); err == nil {
		t.Fatalf("Create() with an unknown user must fail")
	}

	if _, err := svc.Create(context.Background(), CreateInput{
		UserID:   userID,
		Title:    "Нет табака",
		Tobaccos: []RecipeTobaccoInput{{TobaccoID: "00000000-0000-0000-0000-000000000001", Percent: 100}},
		Steps:    []RecipeStepInput{{WhatDo: "Что-то"}},
	}); err == nil {
		t.Fatalf("Create() with an unknown tobacco must fail")
	}
}

func TestListFiltersByUserAndTobacco(t *testing.T) {
	svc := newSvc(t)
	userID := createUser(t)
	cherry := createTobacco(t, "Cherry")
	choco := createTobacco(t, "Chocolate")

	_, err := svc.Create(context.Background(), CreateInput{
		UserID:   userID,
		Title:    "Вишнёвая",
		Tobaccos: []RecipeTobaccoInput{{TobaccoID: cherry, Percent: 100}},
		Steps:    []RecipeStepInput{{WhatDo: "Зарядить"}},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, err = svc.Create(context.Background(), CreateInput{
		UserID:   userID,
		Title:    "Шоколадная",
		Tobaccos: []RecipeTobaccoInput{{TobaccoID: choco, Percent: 100}},
		Steps:    []RecipeStepInput{{WhatDo: "Зарядить"}},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	byUser, err := svc.List(context.Background(), ListInput{UserID: userID})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(byUser) != 2 {
		t.Fatalf("List(byUser) = %d recipes, want 2", len(byUser))
	}

	byTobacco, err := svc.List(context.Background(), ListInput{TobaccoIDs: []string{choco}})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(byTobacco) != 1 || byTobacco[0].Title != "Шоколадная" {
		t.Fatalf("List(byTobacco) = %+v, want only Шоколадная", byTobacco)
	}
}

func TestGetUnknownReturnsNotFound(t *testing.T) {
	svc := newSvc(t)

	_, err := svc.Get(context.Background(), "00000000-0000-0000-0000-000000000001")
	if err != ErrNotFound {
		t.Fatalf("Get(unknown) error = %v, want ErrNotFound", err)
	}
}
