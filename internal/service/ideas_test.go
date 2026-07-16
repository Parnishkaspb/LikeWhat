package service

import (
	"context"
	"net"
	"testing"
	"time"

	ideasv1 "github.com/example/ideas-grpc-service/gen/ideas/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestIdeaLifecycle(t *testing.T) {
	t.Parallel()
	server := NewIdeaServer()
	ctx := context.Background()

	created, err := server.CreateIdea(ctx, &ideasv1.CreateIdeaRequest{Title: "First", Description: "Test idea"})
	if err != nil {
		t.Fatalf("CreateIdea() error = %v", err)
	}
	if created.GetId() != "idea-1" {
		t.Fatalf("CreateIdea() id = %q, want idea-1", created.GetId())
	}

	updated, err := server.UpdateIdea(ctx, &ideasv1.UpdateIdeaRequest{Id: created.GetId(), Title: "Updated", Description: "Updated description"})
	if err != nil {
		t.Fatalf("UpdateIdea() error = %v", err)
	}
	if updated.GetTitle() != "Updated" {
		t.Fatalf("UpdateIdea() title = %q", updated.GetTitle())
	}

	listed, err := server.ListIdeas(ctx, &ideasv1.ListIdeasRequest{})
	if err != nil || len(listed.GetIdeas()) != 1 {
		t.Fatalf("ListIdeas() = %+v, %v; want one idea", listed, err)
	}

	if _, err := server.DeleteIdea(ctx, &ideasv1.DeleteIdeaRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteIdea() error = %v", err)
	}
	if _, err := server.GetIdea(ctx, &ideasv1.GetIdeaRequest{Id: created.GetId()}); status.Code(err) != codes.NotFound {
		t.Fatalf("GetIdea() error code = %s, want NotFound", status.Code(err))
	}
}

func TestCreateIdeaRejectsEmptyTitle(t *testing.T) {
	t.Parallel()
	_, err := NewIdeaServer().CreateIdea(context.Background(), &ideasv1.CreateIdeaRequest{Description: "Description"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateIdea() error code = %s, want InvalidArgument", status.Code(err))
	}
}

func TestGRPCTransport(t *testing.T) {
	t.Parallel()
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	ideasv1.RegisterIdeaServiceServer(grpcServer, NewIdeaServer())
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("DialContext() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	idea, err := ideasv1.NewIdeaServiceClient(conn).CreateIdea(ctx, &ideasv1.CreateIdeaRequest{
		Title:       "Transport test",
		Description: "This request crossed the gRPC transport.",
	})
	if err != nil {
		t.Fatalf("CreateIdea() error = %v", err)
	}
	if idea.GetId() != "idea-1" {
		t.Fatalf("CreateIdea() id = %q, want idea-1", idea.GetId())
	}
}
