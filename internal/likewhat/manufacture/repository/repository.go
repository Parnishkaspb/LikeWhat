package repository

import (
	"context"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/errs"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture"
)

// ErrNotFound is returned when a manufacture does not exist.
var ErrNotFound = errs.ErrNotFound

// ManufactureRepository is the persistence boundary for manufacture records.
type ManufactureRepository interface {
	Create(context.Context, manufacture.Manufacture) (manufacture.Manufacture, error)
	Get(context.Context, string) (manufacture.Manufacture, error)
	List(context.Context, manufacture.ListFilter) ([]manufacture.Manufacture, int, error)
	Update(context.Context, manufacture.Manufacture) (manufacture.Manufacture, error)
	Delete(context.Context, string) error
}
