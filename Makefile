.PHONY: generate test run-server run-client

generate:
	PATH="$(HOME)/go/bin:$(PATH)" protoc --go_out=. --go_opt=module=github.com/example/ideas-grpc-service --go-grpc_out=. --go-grpc_opt=module=github.com/example/ideas-grpc-service -I proto -I /opt/homebrew/include proto/ideas/v1/ideas.proto

test:
	go test ./...

run-server:
	go run ./cmd/server

run-client:
	go run ./cmd/client
