package transportgrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/repository"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/service"
	likewhat "github.com/Parnishkaspb/LikeWhat/pkg/like_what"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server adapts tobacco use cases to the generated gRPC contract.
type Server struct {
	likewhat.UnimplementedTobaccoServiceServer
	service *service.TobaccoService
}

func NewServer(service *service.TobaccoService) *Server {
	return &Server{service: service}
}

func (s *Server) CreateTobacco(ctx context.Context, req *likewhat.CreateTobaccoRequest) (*likewhat.Tobacco, error) {
	err := validationCreateTobaccoRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationCreateTobaccoRequest: %v", err))
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

func validationCreateTobaccoRequest(req *likewhat.CreateTobaccoRequest) error {
	if req == nil {
		return errors.New("request is required")
	}

	return validation.ValidateStruct(
		req,
		validation.Field(&req.Taste, validation.Required),
		validation.Field(&req.ManufactureId, validation.Required, is.UUID),
	)
}

func (s *Server) GetTobacco(ctx context.Context, req *likewhat.GetTobaccoRequest) (*likewhat.Tobacco, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	item, err := s.service.Get(ctx, req.GetId())
	if err != nil {
		return nil, toStatusError(err)
	}
	return toProto(item), nil
}

func (s *Server) ListTobaccos(ctx context.Context, req *likewhat.ListTobaccosRequest) (*likewhat.ListTobaccosResponse, error) {
	if req == nil {
		req = &likewhat.ListTobaccosRequest{}
	}

	items, err := s.service.List(ctx, service.ListInput{
		Taste:          req.GetTaste(),
		ManufactureIDs: req.GetManufactureId(),
	})
	if err != nil {
		return nil, toStatusError(err)
	}

	result := make([]*likewhat.Tobacco, 0, len(items))
	for _, item := range items {
		result = append(result, toProto(item))
	}
	return &likewhat.ListTobaccosResponse{Tobaccos: result}, nil
}

func toProto(item tobacco.Tobacco) *likewhat.Tobacco {
	result := &likewhat.Tobacco{
		Id:    item.ID,
		Taste: item.Taste,
		Proto: item.Proto,
		Manufacture: &likewhat.Manufacture{
			Id:   item.Manufacture.ID,
			Name: item.Manufacture.Name,
		},
	}
	if !item.CreatedAt.IsZero() {
		result.CreatedAt = timestamppb.New(item.CreatedAt)
	}
	if !item.UpdatedAt.IsZero() {
		result.UpdatedAt = timestamppb.New(item.UpdatedAt)
	}
	if item.DeletedAt != nil {
		result.DeletedAt = timestamppb.New(*item.DeletedAt)
	}
	return result
}

func toStatusError(err error) error {
	switch {
	case errors.Is(err, service.ErrTasteRequired), errors.Is(err, service.ErrManufactureRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, repository.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
