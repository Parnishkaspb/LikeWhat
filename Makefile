.PHONY: generate test run-server run-client

MODULE := $(shell go list -m -f '{{.Path}}')
PROTO_FILES := $(shell find api -name '*.proto' -type f)
PROTOC_GEN_PATH := $(shell go env GOPATH)/bin

generate:
	PATH="$(PROTOC_GEN_PATH):$(PATH)" protoc -I . --go_out=. --go_opt=module=$(MODULE) --go-grpc_out=. --go-grpc_opt=module=$(MODULE) $(PROTO_FILES)

test:
	go test ./...

run-server:
	go run ./cmd/server

run-client:
	go run ./cmd/client
