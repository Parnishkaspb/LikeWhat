package repository

import (
	"context"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/errs"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco"
)

// ErrNotFound is returned when a tobacco record does not exist.
var ErrNotFound = errs.ErrNotFound

// TobaccoRepository is the persistence boundary for tobacco records.
type TobaccoRepository interface {
	Create(context.Context, tobacco.Tobacco) (tobacco.Tobacco, error)
	Get(context.Context, string) (tobacco.Tobacco, error)
	List(context.Context, tobacco.ListFilter) ([]tobacco.Tobacco, error)
}
