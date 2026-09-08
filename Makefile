.PHONY: generate test run-server run-client migrate-up migrate-down

MODULE := $(shell go list -m -f '{{.Path}}')
PROTO_FILES := $(shell find api -name '*.proto' -type f)
PROTOC_GEN_PATH := $(shell go env GOPATH)/bin
GOOSE := $(shell go env GOPATH)/bin/goose
MIGRATIONS_DIR := migrations
DATABASE_URL ?= postgres://likewhat:likewhat_dev_password@localhost:5432/likewhat?sslmode=disable

generate:
	PATH="$(PROTOC_GEN_PATH):$(PATH)" protoc -I . --go_out=. --go_opt=module=$(MODULE) --go-grpc_out=. --go-grpc_opt=module=$(MODULE) $(PROTO_FILES)

# DB-backed tests need a reachable Docker daemon. Colima exposes one on a
# unix socket; testcontainers will pick up DOCKER_HOST when it is unset.
test:
	@if [ -z "$$DOCKER_HOST" ] && [ -S "$$HOME/.colima/default/docker.sock" ]; then \
		echo "using colima docker socket $$HOME/.colima/default/docker.sock"; \
		DOCKER_HOST="unix://$$HOME/.colima/default/docker.sock" go test ./...; \
	else \
		go test ./...; \
	fi

run-server:
	go run ./cmd/server

run-client:
	go run ./cmd/client

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down
