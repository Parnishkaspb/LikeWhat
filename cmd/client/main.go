package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	likewhat "github.com/Parnishkaspb/LikeWhat/pkg/like_what"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	address := flag.String("addr", "localhost:50051", "gRPC server address")
	taste := flag.String("taste", "vanilla", "tobacco taste")
	photo := flag.String("photo", "", "photo URL")
	manufactureID := flag.String("manufacture-id", "manufacturer-1", "manufacture identifier")
	flag.Parse()

	conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	client := likewhat.NewTobaccoServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	created, err := client.CreateTobacco(ctx, &likewhat.CreateTobaccoRequest{
		Taste:         *taste,
		Photo:         *photo,
		ManufactureId: *manufactureID,
	})
	if err != nil {
		log.Fatalf("create tobacco: %v", err)
	}
	fmt.Printf("created: %s — %s\n", created.GetId(), created.GetTaste())

	tobaccos, err := client.ListTobaccos(ctx, &likewhat.ListTobaccosRequest{})
	if err != nil {
		log.Fatalf("list tobaccos: %v", err)
	}
	for _, tobacco := range tobaccos.GetTobaccos() {
		fmt.Printf("%s | %s | %s\n", tobacco.GetId(), tobacco.GetTaste(), tobacco.GetManufacture().GetId())
	}
}
