package service

import (
	manufacturerepo "github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture/repository"
)

// ManufactureService contains manufacture use cases and is independent of transports.
type ManufactureService struct {
	repository manufacturerepo.ManufactureRepository
}

func NewManufactureService(repository manufacturerepo.ManufactureRepository) *ManufactureService {
	return &ManufactureService{repository: repository}
}
