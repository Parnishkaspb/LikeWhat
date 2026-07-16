package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/repository"
)

func TestCreateNormalizesInput(t *testing.T) {
	t.Parallel()
	svc := NewTobaccoService(repository.NewMemory())

	created, err := svc.Create(context.Background(), CreateInput{
		Taste:         "  Vanilla  ",
		Photo:         " https://example.test/photo.jpg ",
		ManufactureID: " manufacturer-1 ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Taste != "Vanilla" || created.Photo != "https://example.test/photo.jpg" || created.Manufacture.ID != "manufacturer-1" {
		t.Fatalf("Create() = %+v, want trimmed values", created)
	}
}

func TestCreateValidatesRequiredFields(t *testing.T) {
	t.Parallel()
	svc := NewTobaccoService(repository.NewMemory())

	if _, err := svc.Create(context.Background(), CreateInput{ManufactureID: "m-1"}); !errors.Is(err, ErrTasteRequired) {
		t.Fatalf("Create() error = %v, want ErrTasteRequired", err)
	}
	if _, err := svc.Create(context.Background(), CreateInput{Taste: "Vanilla"}); !errors.Is(err, ErrManufactureRequired) {
		t.Fatalf("Create() error = %v, want ErrManufactureRequired", err)
	}
}
