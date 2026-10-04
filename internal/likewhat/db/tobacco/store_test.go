package tobacco

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	manufacturestore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/manufacture"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newStore() *Store { return NewStore(db.NewClient(testpostgres.Pool())) }

// createManufacture inserts a manufacture row used as a tobacco foreign key.
func createManufacture(t *testing.T, ctx context.Context, name string) string {
	t.Helper()
	st := manufacturestore.NewStore(db.NewClient(testpostgres.Pool()))
	item, err := st.CreateManufacture(ctx, models.Manufacture{Name: name})
	if err != nil {
		t.Fatalf("create manufacture: %v", err)
	}
	return item.ID
}

func TestTobaccoLifecycle(t *testing.T) {
	st := newStore()
	ctx := context.Background()
	mID := createManufacture(t, ctx, "Ozon")

	created, err := st.CreateTobacco(ctx, models.Tobacco{
		Taste:       "Vanilla",
		Photo:       "https://example.test/photo.jpg",
		Manufacture: models.Manufacture{ID: mID},
	})
	if err != nil {
		t.Fatalf("CreateTobacco() error = %v", err)
	}
	if created.ID == "" || created.Taste != "Vanilla" || created.Manufacture.ID != mID {
		t.Fatalf("CreateTobacco() = %+v", created)
	}

	got, err := st.GetTobacco(ctx, created.ID)
	if err != nil || got.Taste != "Vanilla" || got.Manufacture.ID != mID {
		t.Fatalf("GetTobacco() = %+v, %v", got, err)
	}
	if got.Manufacture.Name != "Ozon" {
		t.Fatalf("GetTobacco().Manufacture.Name = %q, want Ozon", got.Manufacture.Name)
	}

	if _, err := st.CreateTobacco(ctx, models.Tobacco{Taste: "Cherry", Manufacture: models.Manufacture{ID: mID}}); err != nil {
		t.Fatalf("CreateTobacco() second item error = %v", err)
	}

	items, err := st.ListTobaccos(ctx, models.TobaccoFilter{Taste: "vanilla"})
	if err != nil || len(items) != 1 || items[0].Taste != "Vanilla" {
		t.Fatalf("ListTobaccos() = %+v, %v; want one Vanilla", items, err)
	}

	if _, err := st.GetTobacco(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("GetTobacco(missing) error = %v, want ErrNotFound", err)
	}
}
