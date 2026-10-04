package service

import (
	"context"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
)

// ErrNotFound is returned when a user does not exist.
var ErrNotFound = db.ErrNotFound

// CreateInput contains the business fields accepted when a user is created.
type CreateInput struct {
	TelegramID int64
	NickName   string
	Name       string
}

// ListInput contains optional search criteria.
type ListInput struct {
	NameLike string
}

// UserService contains user use cases and is independent of transports.
// Input validation happens at the transport layer; this service only orchestrates.
type UserService struct {
	store *db.Store
}

func NewUserService(store *db.Store) *UserService {
	return &UserService{store: store}
}

func (s *UserService) Create(ctx context.Context, input CreateInput) (models.User, error) {
	return s.store.CreateUser(ctx, models.User{
		TelegramID: input.TelegramID,
		NickName:   strings.TrimSpace(input.NickName),
		Name:       strings.TrimSpace(input.Name),
	})
}

func (s *UserService) Get(ctx context.Context, id int64) (models.User, error) {
	return s.store.GetUser(ctx, id)
}

func (s *UserService) List(ctx context.Context, input ListInput) ([]models.User, error) {
	return s.store.ListUsers(ctx, models.UserFilter{NameLike: strings.TrimSpace(input.NameLike)})
}
