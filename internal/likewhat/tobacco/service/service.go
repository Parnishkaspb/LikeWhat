package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/repository"
	"github.com/Parnishkaspb/LikeWhat/internal/models"
)

var (
	ErrTasteRequired       = errors.New("taste is required")
	ErrManufactureRequired = errors.New("manufacture_id is required")
)

// CreateInput contains the business fields accepted when a tobacco record is created.
type CreateInput struct {
	Taste         string
	Photo         string
	ManufactureID string
}

// ListInput contains optional search criteria.
type ListInput struct {
	Taste          string
	ManufactureIDs []string
}

// TobaccoService contains tobacco use cases and is independent of transports.
type TobaccoService struct {
	repository repository.TobaccoRepository
}

func NewTobaccoService(repository repository.TobaccoRepository) *TobaccoService {
	return &TobaccoService{repository: repository}
}

func (s *TobaccoService) Create(ctx context.Context, input CreateInput) (models.Tobacco, error) {
	taste := strings.TrimSpace(input.Taste)
	if taste == "" {
		return models.Tobacco{}, ErrTasteRequired
	}

	manufactureID := strings.TrimSpace(input.ManufactureID)
	if manufactureID == "" {
		return models.Tobacco{}, ErrManufactureRequired
	}

	return s.repository.Create(ctx, models.Tobacco{
		Taste: strings.TrimSpace(input.Taste),
		Photo: strings.TrimSpace(input.Photo),
		Manufacture: models.Manufacture{
			ID: manufactureID,
		},
	})
}

func (s *TobaccoService) Get(ctx context.Context, id string) (models.Tobacco, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return models.Tobacco{}, repository.ErrNotFound
	}
	return s.repository.Get(ctx, id)
}

func (s *TobaccoService) List(ctx context.Context, input ListInput) ([]models.Tobacco, error) {
	manufactureIDs := make([]string, 0, len(input.ManufactureIDs))
	seen := make(map[string]struct{}, len(input.ManufactureIDs))
	for _, id := range input.ManufactureIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		manufactureIDs = append(manufactureIDs, id)
	}

	return s.repository.List(ctx, models.ListFilter{
		Taste:          strings.TrimSpace(input.Taste),
		ManufactureIDs: manufactureIDs,
	})
}
