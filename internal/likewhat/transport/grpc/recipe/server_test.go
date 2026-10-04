package transportgrpc

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	recipeservice "github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/recipe"
	"github.com/Parnishkaspb/LikeWhat/internal/platform/postgres/testpostgres"
	likewhat "github.com/Parnishkaspb/LikeWhat/pkg/like_what"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestMain(m *testing.M) { os.Exit(testpostgres.Main(m)) }

func newServer() *Server {
	return NewServer(recipeservice.NewRecipeService(db.NewStore(db.NewClient(testpostgres.Pool()))))
}

func Test_validationCreateRecipeRequest(t *testing.T) {
	t.Parallel()

	validTobaccos := []*likewhat.RecipeTobacco{
		{TobaccoId: "00000000-0000-0000-0000-000000000001", Percent: 100},
	}
	validSteps := []*likewhat.RecipeStep{{WhatDo: "Зарядить"}}

	tests := []struct {
		name string
		req  *likewhat.CreateRecipeRequest
		want bool // want error?
	}{
		{name: "nil request", req: nil, want: true},
		{name: "empty request", req: &likewhat.CreateRecipeRequest{}, want: true},
		{name: "no tobaccos", req: &likewhat.CreateRecipeRequest{UserId: 1, Title: "T", Steps: validSteps}, want: true},
		{name: "bad tobacco uuid", req: &likewhat.CreateRecipeRequest{
			UserId: 1, Title: "T",
			Tobaccos: []*likewhat.RecipeTobacco{{TobaccoId: "not-a-uuid", Percent: 100}},
			Steps:    validSteps,
		}, want: true},
		{name: "whitespace only title", req: &likewhat.CreateRecipeRequest{UserId: 1, Title: "   ", Tobaccos: validTobaccos, Steps: validSteps}, want: true},
		{name: "step with bad tobacco uuid", req: &likewhat.CreateRecipeRequest{
			UserId: 1, Title: "T",
			Tobaccos: validTobaccos,
			Steps:    []*likewhat.RecipeStep{{TobaccoId: "nope", WhatDo: "X"}},
		}, want: true},
		{name: "valid request", req: &likewhat.CreateRecipeRequest{UserId: 1, Title: "T", Tobaccos: validTobaccos, Steps: validSteps}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validationCreateRecipeRequest(tt.req)
			if got := err != nil; got != tt.want {
				t.Fatalf("validationCreateRecipeRequest() error = %v, want error: %t", err, tt.want)
			}
		})
	}
}

func Test_validationGetRecipeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *likewhat.GetRecipeRequest
		want bool // want error?
	}{
		{name: "nil request", req: nil, want: true},
		{name: "empty id", req: &likewhat.GetRecipeRequest{}, want: true},
		{name: "not a uuid", req: &likewhat.GetRecipeRequest{Id: "abc"}, want: true},
		{name: "valid request", req: &likewhat.GetRecipeRequest{Id: "00000000-0000-0000-0000-000000000001"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validationGetRecipeRequest(tt.req)
			if got := err != nil; got != tt.want {
				t.Fatalf("validationGetRecipeRequest() error = %v, want error: %t", err, tt.want)
			}
		})
	}
}

// seedRecipeData creates the user and tobaccos a gRPC e2e test needs.
func seedRecipeData(t *testing.T) (int64, string, string) {
	t.Helper()

	var userID int64
	if err := testpostgres.Pool().QueryRow(context.Background(),
		`INSERT INTO users (telegram_id, nick_name, name) VALUES ($1, $2, $3) RETURNING id`,
		int64(777), "@grpc", "Grpc Tester").Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	ids := make([]string, 0, 2)
	for _, taste := range []string{"Cherry", "Chocolate"} {
		var manufactureID, id string
		if err := testpostgres.Pool().QueryRow(context.Background(),
			`INSERT INTO manufactures (name) VALUES ($1) RETURNING id`, "Ozon").Scan(&manufactureID); err != nil {
			t.Fatalf("insert manufacture: %v", err)
		}
		if err := testpostgres.Pool().QueryRow(context.Background(),
			`INSERT INTO tobaccos (taste, manufacture_id) VALUES ($1, $2) RETURNING id`, taste, manufactureID).Scan(&id); err != nil {
			t.Fatalf("insert tobacco: %v", err)
		}
		ids = append(ids, id)
	}
	return userID, ids[0], ids[1]
}

func TestRecipeServiceOverGRPC(t *testing.T) {
	t.Parallel()

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	likewhat.RegisterRecipeServiceServer(grpcServer, newServer())
	go func() { _ = grpcServer.Serve(listener) }()
	t.Cleanup(grpcServer.Stop)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := grpc.NewClient("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := likewhat.NewRecipeServiceClient(conn)

	userID, cherry, choco := seedRecipeData(t)

	created, err := client.CreateRecipe(ctx, &likewhat.CreateRecipeRequest{
		UserId: userID,
		Title:  "Шоколадная вишня",
		Tobaccos: []*likewhat.RecipeTobacco{
			{TobaccoId: cherry, Percent: 60},
			{TobaccoId: choco, Percent: 40},
		},
		Steps: []*likewhat.RecipeStep{
			{StepNumber: 1, WhatDo: "Растопите шоколадку"},
			{StepNumber: 2, TobaccoId: cherry, WhatDo: "Зарядите табак в чашу"},
			{StepNumber: 3, WhatDo: "Полейте шоколадом"},
			{StepNumber: 4, WhatDo: "Закройте каллауд"},
		},
	})
	if err != nil {
		t.Fatalf("CreateRecipe() error = %v", err)
	}
	if created.GetId() == "" || created.GetTitle() != "Шоколадная вишня" ||
		len(created.GetTobaccos()) != 2 || len(created.GetSteps()) != 4 {
		t.Fatalf("CreateRecipe() = %+v", created)
	}

	if _, err := client.CreateRecipe(ctx, &likewhat.CreateRecipeRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateRecipe(empty) code = %s, want InvalidArgument", status.Code(err))
	}

	got, err := client.GetRecipe(ctx, &likewhat.GetRecipeRequest{Id: created.GetId()})
	if err != nil || got.GetId() != created.GetId() || len(got.GetSteps()) != 4 {
		t.Fatalf("GetRecipe() = %+v, %v", got, err)
	}

	if _, err := client.GetRecipe(ctx, &likewhat.GetRecipeRequest{Id: "00000000-0000-0000-0000-000000000099"}); status.Code(err) != codes.NotFound {
		t.Fatalf("GetRecipe(missing) code = %s, want NotFound", status.Code(err))
	}

	list, err := client.ListRecipes(ctx, &likewhat.ListRecipesRequest{
		Filter: &likewhat.ListRecipesRequest_Filter{UserId: userID},
	})
	if err != nil || len(list.GetRecipes()) != 1 {
		t.Fatalf("ListRecipes(byUser) = %+v, %v; want one recipe", list, err)
	}

	list, err = client.ListRecipes(ctx, &likewhat.ListRecipesRequest{
		Filter: &likewhat.ListRecipesRequest_Filter{TobaccoIdIn: []string{choco}},
	})
	if err != nil || len(list.GetRecipes()) != 1 || list.GetRecipes()[0].GetId() != created.GetId() {
		t.Fatalf("ListRecipes(byTobacco) = %+v, %v", list, err)
	}
}
