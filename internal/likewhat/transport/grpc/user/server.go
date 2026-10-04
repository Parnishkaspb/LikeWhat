package transportgrpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/user"
	appvalidation "github.com/Parnishkaspb/LikeWhat/internal/platform/validation"
	likewhat "github.com/Parnishkaspb/LikeWhat/pkg/like_what"
	"github.com/Parnishkaspb/LikeWhat/pkg/utils"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server adapts user use cases to the generated gRPC contract.
type Server struct {
	likewhat.UnimplementedUserServiceServer
	service *service.UserService
}

func NewServer(service *service.UserService) *Server {
	return &Server{service: service}
}

func (s *Server) CreateUser(ctx context.Context, req *likewhat.CreateUserRequest) (*likewhat.User, error) {
	if err := validationCreateUserRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationCreateUserRequest: %v", err))
	}

	item, err := s.service.Create(ctx, service.CreateInput{
		TelegramID: req.GetTelegramId(),
		NickName:   strings.TrimSpace(req.GetNickName()),
		Name:       strings.TrimSpace(req.GetName()),
	})
	if err != nil {
		return nil, utils.ToStatusError(err)
	}
	return serializeUser(item), nil
}

func (s *Server) GetUser(ctx context.Context, req *likewhat.GetUserRequest) (*likewhat.User, error) {
	if err := validationGetUserRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationGetUserRequest: %v", err))
	}

	item, err := s.service.Get(ctx, req.GetId())
	if err != nil {
		return nil, utils.ToStatusError(err)
	}
	return serializeUser(item), nil
}

func (s *Server) ListUsers(ctx context.Context, req *likewhat.ListUsersRequest) (*likewhat.ListUsersResponse, error) {
	if req == nil {
		req = &likewhat.ListUsersRequest{}
	}

	users, err := s.service.List(ctx, service.ListInput{})
	if err != nil {
		return nil, utils.ToStatusError(err)
	}

	result := lo.Map(users, func(item models.User, _ int) *likewhat.User {
		return serializeUser(item)
	})

	return &likewhat.ListUsersResponse{Users: result}, nil
}

func validationCreateUserRequest(req *likewhat.CreateUserRequest) error {
	if req == nil {
		return errors.New("request is required")
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.TelegramId, validation.Required),
		validation.Field(&req.NickName, validation.Required, appvalidation.RequiredString()),
		validation.Field(&req.Name, validation.Required, appvalidation.RequiredString()),
	)
}

func validationGetUserRequest(req *likewhat.GetUserRequest) error {
	if req == nil {
		return errors.New("request is required")
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required),
	)
}

func serializeUser(item models.User) *likewhat.User {
	return &likewhat.User{
		Id:         item.ID,
		TelegramId: item.TelegramID,
		NickName:   item.NickName,
		Name:       item.Name,
	}
}
