package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	ideasv1 "github.com/example/ideas-grpc-service/gen/ideas/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	address := flag.String("addr", "localhost:50051", "gRPC server address")
	title := flag.String("title", "gRPC migration", "idea title")
	description := flag.String("description", "Replace the HTTP transport with gRPC", "idea description")
	flag.Parse()

	conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	client := ideasv1.NewIdeaServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	created, err := client.CreateIdea(ctx, &ideasv1.CreateIdeaRequest{Title: *title, Description: *description})
	if err != nil {
		log.Fatalf("create idea: %v", err)
	}
	fmt.Printf("created: %s — %s\n", created.GetId(), created.GetTitle())

	ideas, err := client.ListIdeas(ctx, &ideasv1.ListIdeasRequest{})
	if err != nil {
		log.Fatalf("list ideas: %v", err)
	}
	for _, idea := range ideas.GetIdeas() {
		fmt.Printf("%s | %s | %s\n", idea.GetId(), idea.GetTitle(), idea.GetDescription())
	}
}
