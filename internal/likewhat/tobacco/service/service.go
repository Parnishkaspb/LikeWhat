package service

import (
	"context"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/repository"
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
// Input validation happens at the transport layer; this service only orchestrates.
type TobaccoService struct {
	repository repository.TobaccoRepository
}

func NewTobaccoService(repository repository.TobaccoRepository) *TobaccoService {
	return &TobaccoService{repository: repository}
}

func (s *TobaccoService) Create(ctx context.Context, input CreateInput) (tobacco.Tobacco, error) {
	return s.repository.Create(ctx, tobacco.Tobacco{
		Taste: strings.TrimSpace(input.Taste),
		Photo: strings.TrimSpace(input.Photo),
		Manufacture: tobacco.Manufacture{
			ID: strings.TrimSpace(input.ManufactureID),
		},
	})
}

func (s *TobaccoService) Get(ctx context.Context, id string) (tobacco.Tobacco, error) {
	return s.repository.Get(ctx, strings.TrimSpace(id))
}

func (s *TobaccoService) List(ctx context.Context, input ListInput) ([]tobacco.Tobacco, error) {
	return s.repository.List(ctx, tobacco.ListFilter{
		Taste:          strings.TrimSpace(input.Taste),
		ManufactureIDs: normalizeIDs(input.ManufactureIDs),
	})
}

// normalizeIDs trims and de-duplicates an identifier list.
func normalizeIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
