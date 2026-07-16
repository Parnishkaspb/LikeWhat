package repository

import (
	"context"
	"errors"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco"
)

var ErrNotFound = errors.New("tobacco not found")

// TobaccoRepository is the persistence boundary for tobacco records.
type TobaccoRepository interface {
	Create(context.Context, tobacco.Tobacco) (tobacco.Tobacco, error)
	Get(context.Context, string) (tobacco.Tobacco, error)
	List(context.Context, tobacco.ListFilter) ([]tobacco.Tobacco, error)
}
