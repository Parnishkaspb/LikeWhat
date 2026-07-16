package repository

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco"
)

func TestMemoryCreateGetAndList(t *testing.T) {
	t.Parallel()
	repo := NewMemory()
	ctx := context.Background()

	first, err := repo.Create(ctx, tobacco.Tobacco{Taste: "Vanilla", Manufacture: tobacco.Manufacture{ID: "m-1"}})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if first.ID != "tobacco-1" || first.CreatedAt.IsZero() {
		t.Fatalf("Create() = %+v, want assigned ID and timestamps", first)
	}

	if _, err := repo.Create(ctx, tobacco.Tobacco{Taste: "Cherry", Manufacture: tobacco.Manufacture{ID: "m-2"}}); err != nil {
		t.Fatalf("Create() second item error = %v", err)
	}

	got, err := repo.Get(ctx, first.ID)
	if err != nil || got.Taste != "Vanilla" {
		t.Fatalf("Get() = %+v, %v", got, err)
	}

	items, err := repo.List(ctx, tobacco.ListFilter{ManufactureIDs: []string{"m-2"}})
	if err != nil || len(items) != 1 || items[0].Taste != "Cherry" {
		t.Fatalf("List() = %+v, %v", items, err)
	}

	if _, err := repo.Get(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
}

func TestMemoryConcurrentCreate(t *testing.T) {
	t.Parallel()
	repo := NewMemory()
	ctx := context.Background()

	const workers = 32
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			if _, err := repo.Create(ctx, tobacco.Tobacco{Taste: "Vanilla", Manufacture: tobacco.Manufacture{ID: "m-1"}}); err != nil {
				t.Errorf("Create() error = %v", err)
			}
		}()
	}
	wg.Wait()

	items, err := repo.List(ctx, tobacco.ListFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != workers {
		t.Fatalf("List() returned %d items, want %d", len(items), workers)
	}
}
