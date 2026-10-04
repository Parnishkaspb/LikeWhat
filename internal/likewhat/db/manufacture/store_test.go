package manufacture

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newStore() *Store { return NewStore(db.NewClient(testpostgres.Pool())) }

func TestManufactureLifecycle(t *testing.T) {
	st := newStore()
	ctx := context.Background()

	created, err := st.CreateManufacture(ctx, models.Manufacture{Name: "Ozon"})
	if err != nil {
		t.Fatalf("CreateManufacture() error = %v", err)
	}
	if created.ID == "" || created.Name != "Ozon" || created.CreatedAt.IsZero() {
		t.Fatalf("CreateManufacture() = %+v", created)
	}

	if _, err := st.CreateManufacture(ctx, models.Manufacture{Name: "Starline"}); err != nil {
		t.Fatalf("CreateManufacture() second item error = %v", err)
	}

	items, total, err := st.ListManufactures(ctx, models.ManufactureFilter{NameLike: "tarli"})
	if err != nil || total != 1 || len(items) != 1 || items[0].Name != "Starline" {
		t.Fatalf("ListManufactures(name_like) = %+v total=%d, %v", items, total, err)
	}

	ids, total, err := st.ListManufactures(ctx, models.ManufactureFilter{IDs: []string{created.ID}})
	if err != nil || total != 1 || len(ids) != 1 || ids[0].ID != created.ID {
		t.Fatalf("ListManufactures(ids) = %+v total=%d, %v", ids, total, err)
	}

	updated, err := st.UpdateManufacture(ctx, models.Manufacture{ID: created.ID, Name: "Ozon Force"})
	if err != nil || updated.Name != "Ozon Force" {
		t.Fatalf("UpdateManufacture() = %+v, %v", updated, err)
	}

	if err := st.DeleteManufacture(ctx, created.ID); err != nil {
		t.Fatalf("DeleteManufacture() error = %v", err)
	}
	if _, err := st.GetManufacture(ctx, created.ID); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("GetManufacture(deleted) error = %v, want ErrNotFound", err)
	}
}

func TestManufacturePagination(t *testing.T) {
	st := newStore()
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if _, err := st.CreateManufacture(ctx, models.Manufacture{Name: "M"}); err != nil {
			t.Fatalf("CreateManufacture() error = %v", err)
		}
	}

	items, total, err := st.ListManufactures(ctx, models.ManufactureFilter{Page: 1, PerPage: 2})
	if err != nil || total != 5 || len(items) != 2 {
		t.Fatalf("ListManufactures(page=1,per=2) = %d items total=%d, %v; want 2/5", len(items), total, err)
	}

	items, total, err = st.ListManufactures(ctx, models.ManufactureFilter{Page: 3, PerPage: 2})
	if err != nil || total != 5 || len(items) != 1 {
		t.Fatalf("ListManufactures(page=3,per=2) = %d items total=%d, %v; want 1/5", len(items), total, err)
	}
}
