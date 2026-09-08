package transportgrpc

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture/repository"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/manufacture/service"
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
	return NewServer(service.NewManufactureService(repository.NewPostgres(testpostgres.Pool())))
}

func Test_validationCreateManufactureRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *likewhat.CreateManufactureRequest
		want bool // want error?
	}{
		{name: "nil request", req: nil, want: true},
		{name: "empty request", req: &likewhat.CreateManufactureRequest{}, want: true},
		{name: "whitespace only name", req: &likewhat.CreateManufactureRequest{Name: "   "}, want: true},
		{name: "valid request", req: &likewhat.CreateManufactureRequest{Name: "Ozon"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validationCreateManufactureRequest(tt.req)
			if got := err != nil; got != tt.want {
				t.Fatalf("validationCreateManufactureRequest() error = %v, want error: %t", err, tt.want)
			}
		})
	}
}

func Test_validationEditManufactureRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *likewhat.EditManufactureRequest
		want bool // want error?
	}{
		{name: "nil request", req: nil, want: true},
		{name: "missing id", req: &likewhat.EditManufactureRequest{Name: "Ozon"}, want: true},
		{name: "missing name", req: &likewhat.EditManufactureRequest{Id: "1"}, want: true},
		{name: "whitespace only name", req: &likewhat.EditManufactureRequest{Id: "1", Name: "   "}, want: true},
		{name: "valid request", req: &likewhat.EditManufactureRequest{Id: "1", Name: "Ozon"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validationEditManufactureRequest(tt.req)
			if got := err != nil; got != tt.want {
				t.Fatalf("validationEditManufactureRequest() error = %v, want error: %t", err, tt.want)
			}
		})
	}
}

func TestManufactureServiceOverGRPC(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	likewhat.RegisterManufactureServiceServer(grpcServer, newServer())
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

	client := likewhat.NewManufactureServiceClient(conn)

	created, err := client.CreateManufacture(ctx, &likewhat.CreateManufactureRequest{Name: "Ozon"})
	if err != nil {
		t.Fatalf("CreateManufacture() error = %v", err)
	}
	if created.GetId() == "" || created.GetName() != "Ozon" {
		t.Fatalf("CreateManufacture() = %+v", created)
	}

	if _, err := client.CreateManufacture(ctx, &likewhat.CreateManufactureRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateManufacture(empty name) code = %s, want InvalidArgument", status.Code(err))
	}

	edited, err := client.EditManufacture(ctx, &likewhat.EditManufactureRequest{Id: created.GetId(), Name: "Ozon Force"})
	if err != nil || edited.GetName() != "Ozon Force" {
		t.Fatalf("EditManufacture() = %+v, %v", edited, err)
	}

	list, err := client.ListManufactures(ctx, &likewhat.ListManufacturesRequest{
		Filter: &likewhat.ListManufacturesRequest_Filter{NameLike: "ozon"},
	})
	if err != nil || list.GetTotalCount() != 1 || len(list.GetManufactures()) != 1 {
		t.Fatalf("ListManufactures() = %+v, %v", list, err)
	}

	if _, err := client.DeleteManufacture(ctx, &likewhat.DeleteManufactureRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteManufacture() error = %v", err)
	}

	list, err = client.ListManufactures(ctx, &likewhat.ListManufacturesRequest{})
	if err != nil || list.GetTotalCount() != 0 {
		t.Fatalf("ListManufactures() after delete = %+v, %v; want 0", list, err)
	}
}
