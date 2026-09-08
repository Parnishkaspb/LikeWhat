package transportgrpc

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/repository"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco/service"
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
	return NewServer(service.NewTobaccoService(repository.NewPostgres(testpostgres.Pool())))
}

func Test_validationCreateTobaccoRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *likewhat.CreateTobaccoRequest
		want bool // want error?
	}{
		{name: "nil request", req: nil, want: true},
		{name: "empty request", req: &likewhat.CreateTobaccoRequest{}, want: true},
		{name: "whitespace only taste", req: &likewhat.CreateTobaccoRequest{ManufactureId: "3422b448-2460-4fd2-9183-8000de6f8343"}, want: true},
		{name: "missing manufacture_id", req: &likewhat.CreateTobaccoRequest{Taste: "Vanilla"}, want: true},
		{name: "invalid manufacture_id not uuid", req: &likewhat.CreateTobaccoRequest{Taste: "Vanilla", ManufactureId: "123"}, want: true},
		{name: "valid request", req: &likewhat.CreateTobaccoRequest{Taste: "Vanilla", ManufactureId: "3422b448-2460-4fd2-9183-8000de6f8343"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validationCreateTobaccoRequest(tt.req)
			if got := err != nil; got != tt.want {
				t.Fatalf("validationCreateTobaccoRequest() error = %v, want error: %t", err, tt.want)
			}
		})
	}
}

// createManufacture inserts a manufacture row directly, because a tobacco
// record requires a foreign key to manufactures.
func createManufacture(t *testing.T) string {
	t.Helper()
	var id string
	err := testpostgres.Pool().QueryRow(context.Background(),
		`INSERT INTO manufactures (name) VALUES ($1) RETURNING id`, "Ozon").Scan(&id)
	if err != nil {
		t.Fatalf("insert manufacture: %v", err)
	}
	return id
}

func TestTobaccoServiceOverGRPC(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	likewhat.RegisterTobaccoServiceServer(grpcServer, newServer())
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

	mID := createManufacture(t)
	client := likewhat.NewTobaccoServiceClient(conn)

	created, err := client.CreateTobacco(ctx, &likewhat.CreateTobaccoRequest{Taste: "Vanilla", ManufactureId: mID})
	if err != nil {
		t.Fatalf("CreateTobacco() error = %v", err)
	}
	if created.GetId() == "" || created.GetManufacture().GetId() != mID {
		t.Fatalf("CreateTobacco() = %+v", created)
	}

	items, err := client.ListTobaccos(ctx, &likewhat.ListTobaccosRequest{Taste: "vanilla"})
	if err != nil || len(items.GetTobaccos()) != 1 {
		t.Fatalf("ListTobaccos() = %+v, %v; want one item", items, err)
	}

	if _, err := client.GetTobacco(ctx, &likewhat.GetTobaccoRequest{Id: "00000000-0000-0000-0000-000000000000"}); status.Code(err) != codes.NotFound {
		t.Fatalf("GetTobacco() code = %s, want NotFound", status.Code(err))
	}
}
