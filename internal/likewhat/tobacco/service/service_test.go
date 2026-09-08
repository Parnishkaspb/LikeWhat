package service

import (
	"context"
	"os"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/repository"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newSvc(t *testing.T) *TobaccoService {
	t.Helper()
	return NewTobaccoService(repository.NewPostgres(testpostgres.Pool()))
}

// createManufacture inserts a manufacture row directly, because a tobacco
// record requires a foreign key to manufactures.
func createManufacture(t *testing.T) string {
	t.Helper()
	var id string
	err := testpostgres.Pool().QueryRow(context.Background(),
		`INSERT INTO manufactures (name) VALUES ($1) RETURNING id`, "Ozon").Scan(&id)
	if err != nil {
		t.Fatalf("insert manufacture: %v", err)
	}
	return id
}

func TestCreateNormalizesInput(t *testing.T) {
	svc := newSvc(t)
	mID := createManufacture(t)

	created, err := svc.Create(context.Background(), CreateInput{
		Taste:         "  Vanilla  ",
		Photo:         " https://example.test/photo.jpg ",
		ManufactureID: mID,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" || created.Taste != "Vanilla" || created.Photo != "https://example.test/photo.jpg" ||
		created.Manufacture.ID != mID {
		t.Fatalf("Create() = %+v, want trimmed values", created)
	}
}

func TestGetReturnsCreated(t *testing.T) {
	svc := newSvc(t)
	mID := createManufacture(t)

	created, err := svc.Create(context.Background(), CreateInput{Taste: "Vanilla", ManufactureID: mID})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := svc.Get(context.Background(), created.ID)
	if err != nil || got.ID != created.ID || got.Taste != "Vanilla" {
		t.Fatalf("Get() = %+v, %v", got, err)
	}
}
