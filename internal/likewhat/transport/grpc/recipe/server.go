package transportgrpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/models"
	recipeservice "github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/recipe"
	appvalidation "github.com/Parnishkaspb/LikeWhat/internal/platform/validation"
	likewhat "github.com/Parnishkaspb/LikeWhat/pkg/like_what"
	"github.com/Parnishkaspb/LikeWhat/pkg/utils"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server adapts recipe use cases to the generated gRPC contract.
type Server struct {
	likewhat.UnimplementedRecipeServiceServer
	service *recipeservice.RecipeService
}

func NewServer(service *recipeservice.RecipeService) *Server {
	return &Server{service: service}
}

func (s *Server) CreateRecipe(ctx context.Context, req *likewhat.CreateRecipeRequest) (*likewhat.Recipe, error) {
	if err := validationCreateRecipeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationCreateRecipeRequest: %v", err))
	}

	item, err := s.service.Create(ctx, recipeservice.CreateInput{
		UserID:   req.GetUserId(),
		Title:    req.GetTitle(),
		Tobaccos: tobaccosFromProto(req.GetTobaccos()),
		Steps:    stepsFromProto(req.GetSteps()),
	})
	if err != nil {
		return nil, utils.ToStatusError(err)
	}
	return serializeRecipe(item), nil
}

func validationCreateRecipeRequest(req *likewhat.CreateRecipeRequest) error {
	if req == nil {
		return errors.New("request is required")
	}
	if err := validation.ValidateStruct(req,
		validation.Field(&req.UserId, validation.Required),
		validation.Field(&req.Title, validation.Required, appvalidation.RequiredString()),
	); err != nil {
		return err
	}
	if len(req.GetTobaccos()) == 0 {
		return errors.New("tobaccos: at least one tobacco is required")
	}
	for i, t := range req.GetTobaccos() {
		if err := validation.Validate(t.GetTobaccoId(), validation.Required, is.UUID); err != nil {
			return fmt.Errorf("tobaccos[%d].tobacco_id: %w", i, err)
		}
	}
	for i, st := range req.GetSteps() {
		if err := validation.Validate(st.GetWhatDo(), validation.Required); err != nil {
			return fmt.Errorf("steps[%d].what_do: %w", i, err)
		}
		if st.GetTobaccoId() != "" {
			if err := validation.Validate(st.GetTobaccoId(), is.UUID); err != nil {
				return fmt.Errorf("steps[%d].tobacco_id: %w", i, err)
			}
		}
	}
	return nil
}

func (s *Server) GetRecipe(ctx context.Context, req *likewhat.GetRecipeRequest) (*likewhat.Recipe, error) {
	if err := validationGetRecipeRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("validationGetRecipeRequest: %v", err))
	}
	item, err := s.service.Get(ctx, req.GetId())
	if err != nil {
		return nil, utils.ToStatusError(err)
	}
	return serializeRecipe(item), nil
}

func validationGetRecipeRequest(req *likewhat.GetRecipeRequest) error {
	if req == nil {
		return errors.New("request is required")
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required, is.UUID, appvalidation.RequiredString()),
	)
}

func (s *Server) ListRecipes(ctx context.Context, req *likewhat.ListRecipesRequest) (*likewhat.ListRecipesResponse, error) {
	if req == nil {
		req = &likewhat.ListRecipesRequest{}
	}
	if err := validationIdsIn(req.GetFilter().GetTobaccoIdIn()); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("ListRecipes: %v", err))
	}

	var filter recipeservice.ListInput
	if req.GetFilter() != nil {
		filter = recipeservice.ListInput{
			UserID:     req.GetFilter().GetUserId(),
			TobaccoIDs: req.GetFilter().GetTobaccoIdIn(),
		}
	}

	items, err := s.service.List(ctx, filter)
	if err != nil {
		return nil, utils.ToStatusError(err)
	}

	result := make([]*likewhat.Recipe, 0, len(items))
	for _, item := range items {
		result = append(result, serializeRecipe(item))
	}
	return &likewhat.ListRecipesResponse{Recipes: result}, nil
}

// validationIdsIn checks that every identifier in a list filter is a non-empty
// UUID, so a malformed value fails with InvalidArgument instead of a DB error.
func validationIdsIn(ids []string) error {
	for _, id := range ids {
		if err := validation.Validate(id, validation.Required, is.UUID); err != nil {
			return fmt.Errorf("tobacco_id_in: %w", err)
		}
	}
	return nil
}

func tobaccosFromProto(items []*likewhat.RecipeTobacco) []recipeservice.RecipeTobaccoInput {
	result := make([]recipeservice.RecipeTobaccoInput, 0, len(items))
	for _, t := range items {
		if t == nil {
			continue
		}
		result = append(result, recipeservice.RecipeTobaccoInput{
			TobaccoID: strings.TrimSpace(t.GetTobaccoId()),
			Percent:   t.GetPercent(),
		})
	}
	return result
}

func stepsFromProto(items []*likewhat.RecipeStep) []recipeservice.RecipeStepInput {
	result := make([]recipeservice.RecipeStepInput, 0, len(items))
	for _, st := range items {
		if st == nil {
			continue
		}
		result = append(result, recipeservice.RecipeStepInput{
			StepNumber: int(st.GetStepNumber()),
			TobaccoID:  strings.TrimSpace(st.GetTobaccoId()),
			WhatDo:     st.GetWhatDo(),
		})
	}
	return result
}

func serializeRecipe(item models.Recipe) *likewhat.Recipe {
	result := &likewhat.Recipe{
		Id:     item.ID,
		UserId: item.UserID,
		Title:  item.Title,
	}
	for _, t := range item.Tobaccos {
		result.Tobaccos = append(result.Tobaccos, &likewhat.RecipeTobacco{
			TobaccoId: t.TobaccoID,
			Percent:   t.Percent,
		})
	}
	for _, st := range item.Steps {
		result.Steps = append(result.Steps, &likewhat.RecipeStep{
			StepNumber: int64(st.StepNumber),
			TobaccoId:  st.TobaccoID,
			WhatDo:     st.WhatDo,
		})
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
