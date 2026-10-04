package service

import (
	"context"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
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
	store *db.Store
}

func NewTobaccoService(store *db.Store) *TobaccoService {
	return &TobaccoService{store: store}
}

func (s *TobaccoService) Create(ctx context.Context, input CreateInput) (models.Tobacco, error) {
	return s.store.CreateTobacco(ctx, models.Tobacco{
		Taste: strings.TrimSpace(input.Taste),
		Photo: strings.TrimSpace(input.Photo),
		Manufacture: models.Manufacture{
			ID: strings.TrimSpace(input.ManufactureID),
		},
	})
}

func (s *TobaccoService) Get(ctx context.Context, id string) (models.Tobacco, error) {
	return s.store.GetTobacco(ctx, strings.TrimSpace(id))
}

func (s *TobaccoService) List(ctx context.Context, input ListInput) ([]models.Tobacco, error) {
	return s.store.ListTobaccos(ctx, models.TobaccoFilter{
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
