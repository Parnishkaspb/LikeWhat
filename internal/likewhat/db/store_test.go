package db

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newStore() *Store { return NewStore(NewClient(testpostgres.Pool())) }

func TestUserLifecycle(t *testing.T) {
	st := newStore()
	ctx := context.Background()

	created, err := st.CreateUser(ctx, models.User{TelegramID: 123456789, NickName: "nick", Name: "John"})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if created.ID == 0 || created.TelegramID != 123456789 || created.NickName != "nick" || created.Name != "John" {
		t.Fatalf("CreateUser() = %+v, want assigned ID and fields", created)
	}

	got, err := st.GetUser(ctx, created.ID)
	if err != nil || got.ID != created.ID || got.Name != "John" {
		t.Fatalf("GetUser() = %+v, %v", got, err)
	}

	if _, err := st.CreateUser(ctx, models.User{TelegramID: 987, NickName: "other", Name: "Jane"}); err != nil {
		t.Fatalf("CreateUser() second item error = %v", err)
	}

	items, err := st.ListUsers(ctx, models.UserFilter{NameLike: "john"})
	if err != nil || len(items) != 1 || items[0].Name != "John" {
		t.Fatalf("ListUsers() = %+v, %v; want one John", items, err)
	}

	// LIKE wildcards in input are matched literally, not as wildcards.
	if _, err := st.CreateUser(ctx, models.User{TelegramID: 555, NickName: "x", Name: "100% sure"}); err != nil {
		t.Fatalf("CreateUser() percent name error = %v", err)
	}
	items, err = st.ListUsers(ctx, models.UserFilter{NameLike: "100% sure"})
	if err != nil || len(items) != 1 || items[0].Name != "100% sure" {
		t.Fatalf("ListUsers(percent) = %+v, %v; want literal percent match", items, err)
	}

	if _, err := st.GetUser(ctx, 999999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetUser(missing) error = %v, want ErrNotFound", err)
	}
}

// createManufacture inserts a manufacture row used as a tobacco foreign key.
func createManufacture(t *testing.T, st *Store, ctx context.Context, name string) string {
	t.Helper()
	item, err := st.CreateManufacture(ctx, models.Manufacture{Name: name})
	if err != nil {
		t.Fatalf("create manufacture: %v", err)
	}
	return item.ID
}

func TestTobaccoLifecycle(t *testing.T) {
	st := newStore()
	ctx := context.Background()
	mID := createManufacture(t, st, ctx, "Ozon")

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

	if _, err := st.GetTobacco(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetTobacco(missing) error = %v, want ErrNotFound", err)
	}
}

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
	if _, err := st.GetManufacture(ctx, created.ID); !errors.Is(err, ErrNotFound) {
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
