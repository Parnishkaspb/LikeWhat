package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	manufacturerepo "github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture/repository"
	manufactureservice "github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture/service"
	manufacturetransport "github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture/transport/grpc"
	tobaccorepo "github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/repository"
	tobaccoservice "github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/service"
	tobacotransport "github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/transport/grpc"
	likewhat "github.com/Parnishkaspb/LikeWhat/pkg/like_what"
	"github.com/jackc/pgx/v5/pgxpool"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatalf("DATABASE_URL is required (e.g. %q)", "postgres://likewhat:likewhat_dev_password@localhost:5432/likewhat?sslmode=disable")
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("parse database config: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("connect database: %v", err)
	}

	tobaccoRepo := tobaccorepo.NewPostgres(pool)
	manufactureRepo := manufacturerepo.NewPostgres(pool)

	server := grpc.NewServer()
	likewhat.RegisterTobaccoServiceServer(server, tobacotransport.NewServer(tobaccoservice.NewTobaccoService(tobaccoRepo)))
	likewhat.RegisterManufactureServiceServer(server, manufacturetransport.NewServer(manufactureservice.NewManufactureService(manufactureRepo)))
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	go func() {
		log.Printf("gRPC server listening on %s", address)
		if err := server.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	<-ctx.Done()

	done := make(chan struct{})
	go func() { server.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		server.Stop()
	}
}
