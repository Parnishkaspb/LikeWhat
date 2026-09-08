package user

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

	if _, err := st.GetUser(ctx, 999999); !errors.Is(err, db.ErrNotFound) {
		t.Fatalf("GetUser(missing) error = %v, want ErrNotFound", err)
	}
}
