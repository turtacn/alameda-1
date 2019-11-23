
# Image URL to use all building/pushing image targets
IMG ?= ai-dispatcher 
COMMIT := $(shell git rev-parse --short HEAD)

.PHONY: all test ai-dispatcher 
all: test ai-dispatcher 

# Run tests
test: generate fmt vet
	go test . -coverprofile cover.out

# Build ai-dispatcher binary
ai-dispatcher: generate fmt vet
	go build -ldflags "-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'" -o bin/ai-dispatcher github.com/containers-ai/alameda/ai-dispatcher/cmd

.PHONY: run

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/main.go

.PHONY: fmt vet generate docker-build docker-push

# Run go fmt against code
fmt:
	go fmt .

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate:
	go generate .

# Build the docker image
docker-build: test
	docker build ./.. -t ${IMG}:${COMMIT} -f Dockerfile
