package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture/repository"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newSvc() *ManufactureService {
	return NewManufactureService(repository.NewPostgres(testpostgres.Pool()))
}

func TestCreateNormalizesInput(t *testing.T) {
	svc := newSvc()

	created, err := svc.Create(context.Background(), CreateInput{Name: "  Ozon  "})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" || created.Name != "Ozon" {
		t.Fatalf("Create() = %+v, want trimmed name", created)
	}
}

func TestGetReturnsCreated(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateInput{Name: "Ozon"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := svc.Get(ctx, created.ID)
	if err != nil || got.ID != created.ID || got.Name != "Ozon" {
		t.Fatalf("Get() = %+v, %v", got, err)
	}
}

func TestUpdateAndDelete(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateInput{Name: "Old"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := svc.Update(ctx, UpdateInput{ID: created.ID, Name: "New"})
	if err != nil || updated.Name != "New" {
		t.Fatalf("Update() = %+v, %v", updated, err)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := svc.Get(ctx, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(deleted) error = %v, want ErrNotFound", err)
	}
}
