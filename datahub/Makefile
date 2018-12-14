
# Image URL to use all building/pushing image targets
IMG ?= datahub:latest

.PHONY: all test datahub 
all: test manager

# Run tests
test: generate fmt vet
	go test ./pkg/... ./cmd/... ./services/... -coverprofile cover.out

# Build datahub binary
datahub: generate fmt vet
	go build -o bin/datahub github.com/containers-ai/alameda/datahub/cmd

.PHONY: run

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/main.go

.PHONY: fmt vet generate docker-build docker-push

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}
