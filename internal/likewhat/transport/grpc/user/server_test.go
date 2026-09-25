package transportgrpc

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/db"
	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/service/user"
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
	return NewServer(service.NewUserService(db.NewStore(db.NewClient(testpostgres.Pool()))))
}

func Test_validationCreateUserRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *likewhat.CreateUserRequest
		want bool // want error?
	}{
		{name: "nil request", req: nil, want: true},
		{name: "empty request", req: &likewhat.CreateUserRequest{}, want: true},
		{name: "zero telegram_id", req: &likewhat.CreateUserRequest{NickName: "nick", Name: "John"}, want: true},
		{name: "whitespace only nick", req: &likewhat.CreateUserRequest{TelegramId: 1, NickName: "   ", Name: "John"}, want: true},
		{name: "valid request", req: &likewhat.CreateUserRequest{TelegramId: 1, NickName: "nick", Name: "John"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validationCreateUserRequest(tt.req)
			if got := err != nil; got != tt.want {
				t.Fatalf("validationCreateUserRequest() error = %v, want error: %t", err, tt.want)
			}
		})
	}
}

func Test_validationGetUserRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *likewhat.GetUserRequest
		want bool // want error?
	}{
		{name: "nil request", req: nil, want: true},
		{name: "zero id", req: &likewhat.GetUserRequest{}, want: true},
		{name: "valid request", req: &likewhat.GetUserRequest{Id: 1}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validationGetUserRequest(tt.req)
			if got := err != nil; got != tt.want {
				t.Fatalf("validationGetUserRequest() error = %v, want error: %t", err, tt.want)
			}
		})
	}
}

func TestUserServiceOverGRPC(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	likewhat.RegisterUserServiceServer(grpcServer, newServer())
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

	client := likewhat.NewUserServiceClient(conn)

	created, err := client.CreateUser(ctx, &likewhat.CreateUserRequest{TelegramId: 111, NickName: "nick", Name: "John"})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if created.GetId() == 0 || created.GetName() != "John" || created.GetTelegramId() != 111 {
		t.Fatalf("CreateUser() = %+v", created)
	}

	if _, err := client.CreateUser(ctx, &likewhat.CreateUserRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateUser(empty) code = %s, want InvalidArgument", status.Code(err))
	}

	got, err := client.GetUser(ctx, &likewhat.GetUserRequest{Id: created.GetId()})
	if err != nil || got.GetId() != created.GetId() || got.GetName() != "John" {
		t.Fatalf("GetUser() = %+v, %v", got, err)
	}

	if _, err := client.GetUser(ctx, &likewhat.GetUserRequest{Id: 999999}); status.Code(err) != codes.NotFound {
		t.Fatalf("GetUser(missing) code = %s, want NotFound", status.Code(err))
	}

	list, err := client.ListUsers(ctx, &likewhat.ListUsersRequest{})
	if err != nil || len(list.GetUsers()) != 1 {
		t.Fatalf("ListUsers() = %+v, %v; want one user", list, err)
	}

	if _, err := client.CreateUser(ctx, &likewhat.CreateUserRequest{TelegramId: 111, NickName: "nick2", Name: "Jane"}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("CreateUser(duplicate telegram_id) code = %s, want AlreadyExists", status.Code(err))
	}
}
