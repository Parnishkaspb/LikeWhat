package service

import (
	"context"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
)

// ErrNotFound is returned when a manufacture does not exist.
var ErrNotFound = db.ErrNotFound

// CreateInput contains the business fields accepted when a manufacture is created.
type CreateInput struct {
	Name string
}

// UpdateInput contains the fields available when a manufacture is edited.
type UpdateInput struct {
	ID   string
	Name string
}

// ListInput contains optional search and pagination criteria.
type ListInput struct {
	NameLike string
	IDs      []string
	Page     uint64
	PerPage  uint64
}

// ManufactureService contains manufacture use cases and is independent of transports.
// Input validation happens at the transport layer; this service only orchestrates.
type ManufactureService struct {
	store *db.Store
}

func NewManufactureService(store *db.Store) *ManufactureService {
	return &ManufactureService{store: store}
}

func (s *ManufactureService) Create(ctx context.Context, input CreateInput) (models.Manufacture, error) {
	return s.store.CreateManufacture(ctx, models.Manufacture{Name: strings.TrimSpace(input.Name)})
}

func (s *ManufactureService) Get(ctx context.Context, id string) (models.Manufacture, error) {
	return s.store.GetManufacture(ctx, strings.TrimSpace(id))
}

func (s *ManufactureService) List(ctx context.Context, input ListInput) ([]models.Manufacture, int, error) {
	return s.store.ListManufactures(ctx, models.ManufactureFilter{
		NameLike: strings.TrimSpace(input.NameLike),
		IDs:      input.IDs,
		Page:     input.Page,
		PerPage:  input.PerPage,
	})
}

func (s *ManufactureService) Update(ctx context.Context, input UpdateInput) (models.Manufacture, error) {
	return s.store.UpdateManufacture(ctx, models.Manufacture{
		ID:   strings.TrimSpace(input.ID),
		Name: strings.TrimSpace(input.Name),
	})
}

func (s *ManufactureService) Delete(ctx context.Context, id string) error {
	return s.store.DeleteManufacture(ctx, strings.TrimSpace(id))
}
