package repository

import (
	"context"
	"errors"

	"github.com/Parnishkaspb/LikeWhat/internal/models"
)

var ErrNotFound = errors.New("tobacco not found")

// TobaccoRepository is the persistence boundary for tobacco records.
type TobaccoRepository interface {
	Create(context.Context, models.Tobacco) (models.Tobacco, error)
	Get(context.Context, string) (models.Tobacco, error)
	List(context.Context, models.ListFilter) ([]models.Tobacco, error)
}
