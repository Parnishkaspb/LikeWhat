package transportgrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/service"
	likewhat "github.com/Parnishkaspb/LikeWhat/pkg/like_what"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server .
type Server struct {
	likewhat.UnimplementedManufactureServiceServer
	service *likewhat.ManufactureServiceServer
}

func NewServer(service *likewhat.ManufactureServiceServer) *Server {
	return &Server{service: service}
}

func (s *Server) CreateManufacture(ctx context.Context, req *likewhat.CreateManufactureRequest) (*likewhat.Manufacture, error) {
	err := validationCreateManufactureRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationCreateManufactureRequest: %v", err))
	}

	item, err := s.service.Create(ctx, service.CreateInput{
		Taste:         req.GetTaste(),
		Photo:         req.GetPhoto(),
		ManufactureID: req.GetManufactureId(),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return toProto(item), nil
}

func validationCreateManufactureRequest(req *likewhat.CreateManufactureRequest) error {
	if req == nil {
		return errors.New("request is required")
	}

	return validation.ValidateStruct(
		req,
		validation.Field(&req.Name, validation.Required),
	)
}
