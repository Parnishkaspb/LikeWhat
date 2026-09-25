package transportgrpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/manufacture"
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

// defaultPerPage is the page size used when a client does not request one.
const defaultPerPage = 20

// Server adapts manufacture use cases to the generated gRPC contract.
type Server struct {
	likewhat.UnimplementedManufactureServiceServer
	service *service.ManufactureService
}

func NewServer(service *service.ManufactureService) *Server {
	return &Server{service: service}
}

func (s *Server) CreateManufacture(ctx context.Context, req *likewhat.CreateManufactureRequest) (*likewhat.Manufacture, error) {
	if err := validationCreateManufactureRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationCreateManufactureRequest: %v", err))
	}

	item, err := s.service.Create(ctx, service.CreateInput{Name: strings.TrimSpace(req.GetName())})
	if err != nil {
		return nil, utils.ToStatusError(err)
	}
	return serializeManufacture(item), nil
}

func (s *Server) ListManufactures(ctx context.Context, req *likewhat.ListManufacturesRequest) (*likewhat.ListManufacturesResponse, error) {
	if req == nil {
		req = &likewhat.ListManufacturesRequest{}
	}

	const maxPerPage = 100
	perPage := req.GetPerPage()
	if perPage == 0 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("ListManufactures: per_page must not exceed %d", maxPerPage))
	}

	filter := req.GetFilter()
	page := max(req.GetPage(), 1)

	if err := validationIdsIn(filter.GetIdsIn()); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("ListManufactures: %v", err))
	}

	manufactures, total, err := s.service.List(ctx, service.ListInput{
		NameLike: filter.GetNameLike(),
		IDs:      filter.GetIdsIn(),
		Page:     page,
		PerPage:  req.GetPerPage(),
	})
	if err != nil {
		return nil, utils.ToStatusError(err)
	}

	result := lo.Map(manufactures, func(item models.Manufacture, _ int) *likewhat.Manufacture {
		return serializeManufacture(item)
	})

	return &likewhat.ListManufacturesResponse{
		Manufactures: result,
		TotalCount:   uint64(total),
		// There is a next page while the fetched page does not reach the end
		// of the matched set.
		NextPage: uint64(page)*req.GetPerPage()+uint64(len(manufactures)) < uint64(total),
	}, nil
}

func (s *Server) EditManufacture(ctx context.Context, req *likewhat.EditManufactureRequest) (*likewhat.Manufacture, error) {
	if err := validationEditManufactureRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationEditManufactureRequest: %v", err))
	}

	item, err := s.service.Update(ctx, service.UpdateInput{
		ID:   strings.TrimSpace(req.GetId()),
		Name: strings.TrimSpace(req.GetName()),
	})
	if err != nil {
		return nil, utils.ToStatusError(err)
	}
	return serializeManufacture(item), nil
}

func (s *Server) DeleteManufacture(ctx context.Context, req *likewhat.DeleteManufactureRequest) (*likewhat.DeleteManufactureResponse, error) {
	err := validationDeleteManufactureRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationDeleteManufactureRequest: %v", err))
	}
	if err := s.service.Delete(ctx, req.GetId()); err != nil {
		return nil, utils.ToStatusError(err)
	}
	return &likewhat.DeleteManufactureResponse{IsDeleted: true, Message: "manufacture deleted"}, nil
}

func validationDeleteManufactureRequest(req *likewhat.DeleteManufactureRequest) error {
	if req == nil {
		return errors.New("request is required")
	}

	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required, is.UUID),
	)
}

func validationCreateManufactureRequest(req *likewhat.CreateManufactureRequest) error {
	if req == nil {
		return errors.New("request is required")
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.Name, validation.Required, appvalidation.RequiredString()),
	)
}

func validationEditManufactureRequest(req *likewhat.EditManufactureRequest) error {
	if req == nil {
		return errors.New("request is required")
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required, is.UUID),
		validation.Field(&req.Name, validation.Required, appvalidation.RequiredString()),
	)
}

// validationIdsIn checks that every identifier in a list filter is a non-empty
// UUID, so a malformed value fails with InvalidArgument instead of a DB error.
func validationIdsIn(ids []string) error {
	for _, id := range ids {
		if err := validation.Validate(id, validation.Required, is.UUID); err != nil {
			return fmt.Errorf("ids_in: %w", err)
		}
	}
	return nil
}

func serializeManufacture(item models.Manufacture) *likewhat.Manufacture {
	result := &likewhat.Manufacture{
		Id:   item.ID,
		Name: item.Name,
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
