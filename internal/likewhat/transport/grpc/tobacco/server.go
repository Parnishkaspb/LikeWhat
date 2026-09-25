package transportgrpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/tobacco"
	appvalidation "github.com/Parnishkaspb/LikeWhat/internal/platform/validation"
	likewhat "github.com/Parnishkaspb/LikeWhat/pkg/like_what"
	"github.com/Parnishkaspb/LikeWhat/pkg/utils"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/samber/lo"
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
	if err := validationCreateTobaccoRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationCreateTobaccoRequest: %v", err))
	}

	item, err := s.service.Create(ctx, service.CreateInput{
		Taste:         strings.TrimSpace(req.GetTaste()),
		Photo:         strings.TrimSpace(req.GetPhoto()),
		ManufactureID: strings.TrimSpace(req.GetManufactureId()),
	})
	if err != nil {
		return nil, utils.ToStatusError(err)
	}
	return serializeTobacco(item), nil
}

func validationCreateTobaccoRequest(req *likewhat.CreateTobaccoRequest) error {
	if req == nil {
		return errors.New("request is required")
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.Taste, validation.Required, appvalidation.RequiredString()),
		validation.Field(&req.ManufactureId, validation.Required, appvalidation.RequiredString(), is.UUID),
	)
}

func (s *Server) GetTobacco(ctx context.Context, req *likewhat.GetTobaccoRequest) (*likewhat.Tobacco, error) {
	err := validationGetTobaccoRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationGetTobaccoRequest: %v", err))
	}
	item, err := s.service.Get(ctx, req.GetId())
	if err != nil {
		return nil, utils.ToStatusError(err)
	}
	return serializeTobacco(item), nil
}

func validationGetTobaccoRequest(req *likewhat.GetTobaccoRequest) error {
	if req == nil {
		return errors.New("request is required")
	}

	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required, is.UUID, appvalidation.RequiredString()),
	)
}

func (s *Server) ListTobaccos(ctx context.Context, req *likewhat.ListTobaccosRequest) (*likewhat.ListTobaccosResponse, error) {
	if req == nil {
		req = &likewhat.ListTobaccosRequest{}
	}

	if err := validationIdsIn(req.GetManufactureId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("ListTobaccos: %v", err))
	}

	tobaccos, err := s.service.List(ctx, service.ListInput{
		Taste:          req.GetTaste(),
		ManufactureIDs: req.GetManufactureId(),
	})
	if err != nil {
		return nil, utils.ToStatusError(err)
	}

	result := lo.Map(tobaccos, func(item models.Tobacco, index int) *likewhat.Tobacco {
		return serializeTobacco(item)
	})

	return &likewhat.ListTobaccosResponse{Tobaccos: result}, nil
}

// validationIdsIn checks that every identifier in a list filter is a non-empty
// UUID, so a malformed value fails with InvalidArgument instead of a DB error.
func validationIdsIn(ids []string) error {
	for _, id := range ids {
		if err := validation.Validate(id, validation.Required, is.UUID); err != nil {
			return fmt.Errorf("manufacture_id: %w", err)
		}
	}
	return nil
}

func serializeTobacco(item models.Tobacco) *likewhat.Tobacco {
	result := &likewhat.Tobacco{
		Id:    item.ID,
		Taste: item.Taste,
		Photo: item.Photo,
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
