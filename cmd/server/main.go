package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	ideasv1 "github.com/example/ideas-grpc-service/gen/ideas/v1"
	"github.com/example/ideas-grpc-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	address := os.Getenv("GRPC_ADDR")
	if address == "" {
		address = ":50051"
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("listen on %s: %v", address, err)
	}

	server := grpc.NewServer()
	ideasv1.RegisterIdeaServiceServer(server, service.NewIdeaServer())
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	go func() {
		log.Printf("gRPC server listening on %s", address)
		if err := server.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	done := make(chan struct{})
	go func() { server.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		server.Stop()
	}
}
