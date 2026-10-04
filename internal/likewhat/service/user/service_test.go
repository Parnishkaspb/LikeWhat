package service

import (
	"context"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	userstore "github.com/Parnishkaspb/LikeWhat/internal/likewhat/db/user"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newSvc() *UserService {
	return NewUserService(userstore.NewStore(db.NewClient(testpostgres.Pool())))
}

func TestCreateNormalizesInput(t *testing.T) {
	svc := newSvc()

	created, err := svc.Create(context.Background(), CreateInput{TelegramID: 123, NickName: "  nick  ", Name: "  John  "})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == 0 || created.TelegramID != 123 || created.NickName != "nick" || created.Name != "John" {
		t.Fatalf("Create() = %+v, want trimmed values", created)
	}
}

func TestGetReturnsCreated(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateInput{TelegramID: 123, NickName: "nick", Name: "John"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := svc.Get(ctx, created.ID)
	if err != nil || got.ID != created.ID || got.Name != "John" {
		t.Fatalf("Get() = %+v, %v", got, err)
	}
}
