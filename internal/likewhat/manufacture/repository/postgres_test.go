package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newRepo() *Postgres { return NewPostgres(testpostgres.Pool()) }

func TestPostgresCreateGetAndList(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	created, err := repo.Create(ctx, manufacture.Manufacture{Name: "Ozon"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" || created.Name != "Ozon" || created.CreatedAt.IsZero() {
		t.Fatalf("Create() = %+v, want assigned ID, name and timestamps", created)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil || got.Name != "Ozon" || got.ID != created.ID {
		t.Fatalf("Get() = %+v, %v", got, err)
	}

	if _, err := repo.Create(ctx, manufacture.Manufacture{Name: "Starline"}); err != nil {
		t.Fatalf("Create() second item error = %v", err)
	}

	items, total, err := repo.List(ctx, manufacture.ListFilter{NameLike: "tarli"})
	if err != nil || total != 1 || len(items) != 1 || items[0].Name != "Starline" {
		t.Fatalf("List(name_like) = %+v total=%d, %v; want one Starline", items, total, err)
	}

	ids, total, err := repo.List(ctx, manufacture.ListFilter{IDs: []string{created.ID}})
	if err != nil || total != 1 || len(ids) != 1 || ids[0].ID != created.ID {
		t.Fatalf("List(ids) = %+v total=%d, %v; want one matching id", ids, total, err)
	}

	if _, err := repo.Get(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
}

func TestPostgresListPagination(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if _, err := repo.Create(ctx, manufacture.Manufacture{Name: "M"}); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	items, total, err := repo.List(ctx, manufacture.ListFilter{Page: 1, PerPage: 2})
	if err != nil || total != 5 || len(items) != 2 {
		t.Fatalf("List(page=1,per=2) = %d items total=%d, %v; want 2/5", len(items), total, err)
	}

	items, total, err = repo.List(ctx, manufacture.ListFilter{Page: 3, PerPage: 2})
	if err != nil || total != 5 || len(items) != 1 {
		t.Fatalf("List(page=3,per=2) = %d items total=%d, %v; want 1/5", len(items), total, err)
	}
}

func TestPostgresUpdateAndDelete(t *testing.T) {
	repo := newRepo()
	ctx := context.Background()

	created, err := repo.Create(ctx, manufacture.Manufacture{Name: "Old"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := repo.Update(ctx, manufacture.Manufacture{ID: created.ID, Name: "New"})
	if err != nil || updated.Name != "New" {
		t.Fatalf("Update() = %+v, %v", updated, err)
	}
	if updated.CreatedAt.IsZero() || updated.UpdatedAt.IsZero() {
		t.Fatalf("Update() = %+v, want timestamps", updated)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repo.Get(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(deleted) error = %v, want ErrNotFound", err)
	}
}
