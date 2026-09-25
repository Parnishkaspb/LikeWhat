package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	manufactureservice "github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/manufacture"
	recipeservice "github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/recipe"
	tobaccoservice "github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/tobacco"
	userservice "github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/user"
	manufacturetransport "github.com/Parnishkaspb/LikeWhat/internal/likewhat/transport/grpc/manufacture"
	recipetransport "github.com/Parnishkaspb/LikeWhat/internal/likewhat/transport/grpc/recipe"
	tobacotransport "github.com/Parnishkaspb/LikeWhat/internal/likewhat/transport/grpc/tobacco"
	usertransport "github.com/Parnishkaspb/LikeWhat/internal/likewhat/transport/grpc/user"
	likewhat "github.com/Parnishkaspb/LikeWhat/pkg/like_what"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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

	store := db.NewStore(db.NewClient(pool))

	server := grpc.NewServer()
	likewhat.RegisterTobaccoServiceServer(server, tobacotransport.NewServer(tobaccoservice.NewTobaccoService(store)))
	likewhat.RegisterManufactureServiceServer(server, manufacturetransport.NewServer(manufactureservice.NewManufactureService(store)))
	likewhat.RegisterUserServiceServer(server, usertransport.NewServer(userservice.NewUserService(store)))
	likewhat.RegisterRecipeServiceServer(server, recipetransport.NewServer(recipeservice.NewRecipeService(store)))
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	// Serve reflection so tools like grpcurl and GraphQL gateways can
	// introspect the API without the compiled proto.
	reflection.Register(server)

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
