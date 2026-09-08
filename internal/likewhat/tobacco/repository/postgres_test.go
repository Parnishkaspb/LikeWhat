package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newRepo() *Postgres { return NewPostgres(testpostgres.Pool()) }

// createManufacture inserts a manufacture row directly, because a tobacco
// record requires a foreign key to manufactures.
func createManufacture(t *testing.T, name string) string {
	t.Helper()
	var id string
	err := testpostgres.Pool().QueryRow(context.Background(),
		`INSERT INTO manufactures (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		t.Fatalf("insert manufacture: %v", err)
	}
	return id
}

func TestPostgresCreateGetAndList(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()
	mID := createManufacture(t, "Ozon")

	created, err := repo.Create(ctx, tobacco.Tobacco{
		Taste:       "Vanilla",
		Photo:       "https://example.test/photo.jpg",
		Manufacture: tobacco.Manufacture{ID: mID},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" || created.Taste != "Vanilla" || created.Photo != "https://example.test/photo.jpg" ||
		created.Manufacture.ID != mID || created.CreatedAt.IsZero() {
		t.Fatalf("Create() = %+v, want assigned ID, fields and timestamps", created)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil || got.Taste != "Vanilla" || got.Manufacture.ID != mID {
		t.Fatalf("Get() = %+v, %v", got, err)
	}

	if _, err := repo.Create(ctx, tobacco.Tobacco{
		Taste:       "Cherry",
		Manufacture: tobacco.Manufacture{ID: mID},
	}); err != nil {
		t.Fatalf("Create() second item error = %v", err)
	}

	items, err := repo.List(ctx, tobacco.ListFilter{Taste: "vanilla"})
	if err != nil || len(items) != 1 || items[0].Taste != "Vanilla" {
		t.Fatalf("List(taste) = %+v, %v; want one Vanilla", items, err)
	}

	items, err = repo.List(ctx, tobacco.ListFilter{ManufactureIDs: []string{mID}})
	if err != nil || len(items) != 2 {
		t.Fatalf("List(manufacture_ids) = %d items, %v; want 2", len(items), err)
	}

	if _, err := repo.Get(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
}
