.PHONY: all build test clean proto deps local dev

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=auth-server
BINARY_UNIX=$(BINARY_NAME)_unix

# Proto parameters
PROTOC=protoc
PROTO_DIR=api
PROTO_FILES=$(wildcard $(PROTO_DIR)/*.proto)
PROTO_GO_DIR=internal/auth

all: deps proto build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/auth-server

test:
	$(GOTEST) -v ./...

clean:
	$(GOCMD) clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(PROTO_GO_DIR)/*.pb.go

deps:
	$(GOMOD) download
	$(GOMOD) tidy

proto:
	$(PROTOC) \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

local:
	docker-compose up -d
	./$(BINARY_NAME)

dev:
	air -c .air.toml

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/auth-server

# Docker commands
docker-build:
	docker build -t polyid/auth-server:latest .

docker-run:
	docker run -p 8080:8080 polyid/auth-server:latest

# Kubernetes commands
k8s-deploy:
	kubectl apply -f deployments/kubernetes/

k8s-delete:
	kubectl delete -f deployments/kubernetes/

# Development tools
install-tools:
	go install github.com/cosmtrek/air@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Help
help:
	@echo "Available targets:"
	@echo "  all            - Run deps, proto, and build"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  proto          - Generate protobuf code"
	@echo "  deps           - Download dependencies"
	@echo "  local          - Run with local dependencies"
	@echo "  dev            - Run with hot reload"
	@echo "  build-linux    - Build for Linux"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  k8s-deploy     - Deploy to Kubernetes"
	@echo "  k8s-delete     - Delete from Kubernetes"
	@echo "  install-tools  - Install development tools" 