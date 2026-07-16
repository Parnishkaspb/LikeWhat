package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	ideasv1 "github.com/example/ideas-grpc-service/gen/ideas/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IdeaServer is a concurrency-safe, in-memory implementation of IdeaService.
// Replace its map with a repository implementation when persistent storage is needed.
type IdeaServer struct {
	ideasv1.UnimplementedIdeaServiceServer

	mu     sync.RWMutex
	ideas  map[string]*ideasv1.Idea
	nextID uint64
	now    func() time.Time
}

func NewIdeaServer() *IdeaServer {
	return &IdeaServer{ideas: make(map[string]*ideasv1.Idea), now: time.Now}
}

func (s *IdeaServer) CreateIdea(_ context.Context, req *ideasv1.CreateIdeaRequest) (*ideasv1.Idea, error) {
	if err := validateContent(req.GetTitle(), req.GetDescription()); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	now := timestamppb.New(s.now().UTC())
	idea := &ideasv1.Idea{
		Id:          fmt.Sprintf("idea-%d", s.nextID),
		Title:       strings.TrimSpace(req.GetTitle()),
		Description: strings.TrimSpace(req.GetDescription()),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.ideas[idea.Id] = idea
	return cloneIdea(idea), nil
}

func (s *IdeaServer) GetIdea(_ context.Context, req *ideasv1.GetIdeaRequest) (*ideasv1.Idea, error) {
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	idea, ok := s.ideas[req.GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "idea not found")
	}
	return cloneIdea(idea), nil
}

func (s *IdeaServer) ListIdeas(context.Context, *ideasv1.ListIdeasRequest) (*ideasv1.ListIdeasResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ideas := make([]*ideasv1.Idea, 0, len(s.ideas))
	for _, idea := range s.ideas {
		ideas = append(ideas, cloneIdea(idea))
	}
	sort.Slice(ideas, func(i, j int) bool { return ideas[i].GetId() < ideas[j].GetId() })
	return &ideasv1.ListIdeasResponse{Ideas: ideas}, nil
}

func (s *IdeaServer) UpdateIdea(_ context.Context, req *ideasv1.UpdateIdeaRequest) (*ideasv1.Idea, error) {
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := validateContent(req.GetTitle(), req.GetDescription()); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	idea, ok := s.ideas[req.GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "idea not found")
	}
	idea.Title = strings.TrimSpace(req.GetTitle())
	idea.Description = strings.TrimSpace(req.GetDescription())
	idea.UpdatedAt = timestamppb.New(s.now().UTC())
	return cloneIdea(idea), nil
}

func (s *IdeaServer) DeleteIdea(_ context.Context, req *ideasv1.DeleteIdeaRequest) (*emptypb.Empty, error) {
	if strings.TrimSpace(req.GetId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.ideas[req.GetId()]; !ok {
		return nil, status.Error(codes.NotFound, "idea not found")
	}
	delete(s.ideas, req.GetId())
	return &emptypb.Empty{}, nil
}

func validateContent(title, description string) error {
	if strings.TrimSpace(title) == "" {
		return status.Error(codes.InvalidArgument, "title is required")
	}
	if strings.TrimSpace(description) == "" {
		return status.Error(codes.InvalidArgument, "description is required")
	}
	return nil
}

func cloneIdea(idea *ideasv1.Idea) *ideasv1.Idea {
	return &ideasv1.Idea{
		Id:          idea.GetId(),
		Title:       idea.GetTitle(),
		Description: idea.GetDescription(),
		CreatedAt:   timestamppb.New(idea.GetCreatedAt().AsTime()),
		UpdatedAt:   timestamppb.New(idea.GetUpdatedAt().AsTime()),
	}
}
