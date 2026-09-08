package service

import (
	"context"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture"
	manufacturerepo "github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture/repository"
)

// ErrNotFound is returned when a manufacture does not exist.
var ErrNotFound = manufacturerepo.ErrNotFound

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
	repository manufacturerepo.ManufactureRepository
}

func NewManufactureService(repository manufacturerepo.ManufactureRepository) *ManufactureService {
	return &ManufactureService{repository: repository}
}

func (s *ManufactureService) Create(ctx context.Context, input CreateInput) (manufacture.Manufacture, error) {
	return s.repository.Create(ctx, manufacture.Manufacture{Name: strings.TrimSpace(input.Name)})
}

func (s *ManufactureService) Get(ctx context.Context, id string) (manufacture.Manufacture, error) {
	return s.repository.Get(ctx, strings.TrimSpace(id))
}

func (s *ManufactureService) List(ctx context.Context, input ListInput) ([]manufacture.Manufacture, int, error) {
	return s.repository.List(ctx, manufacture.ListFilter{
		NameLike: strings.TrimSpace(input.NameLike),
		IDs:      input.IDs,
		Page:     input.Page,
		PerPage:  input.PerPage,
	})
}

func (s *ManufactureService) Update(ctx context.Context, input UpdateInput) (manufacture.Manufacture, error) {
	return s.repository.Update(ctx, manufacture.Manufacture{
		ID:   strings.TrimSpace(input.ID),
		Name: strings.TrimSpace(input.Name),
	})
}

func (s *ManufactureService) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, strings.TrimSpace(id))
}
