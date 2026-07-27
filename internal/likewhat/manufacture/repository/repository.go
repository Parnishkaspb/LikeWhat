package repository

import (
	"context"
	"errors"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco"
)

var ErrNotFound = errors.New("entity not found")

// ManufactureRepository is the persistence boundary for manufacture records.
type ManufactureRepository interface {
	Create(context.Context, manufacture.Manufacture) (manufacture.Manufacture, error)
	//Get(context.Context, string) (tobacco.Tobacco, error)
	//List(context.Context, tobacco.ListFilter) ([]manufacture.Manufacture, error)
}

func Create(ctx context.Context, item tobacco.Tobacco) error {

}
