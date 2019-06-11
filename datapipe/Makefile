API_SERVER_VERSION=$(shell cat build/build-image/VERSION)
API_SERVER_COMMITID=$(shell git rev-parse HEAD)

# Image URL to use all building/pushing image targets
IMG ?= datapipe:v0.0.1

.PHONY: all test datapipe
all: test datapipe

## test: Run tests.
test: generate fmt
	go test ./pkg/... ./cmd/... -coverprofile cover.out

## datapipe: Build datapipe executable binary.
datapipe:
	@echo Start building datapipe binary...
	go build -ldflags "-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'" -o build/build-image/bin/datapipe github.com/containers-ai/alameda/datapipe/cmd

.PHONY: run

## run: Run against the configured Kubernetes cluster in ~/.kube/config
run:
	go run ./cmd/main.go run

.PHONY: fmt vet generate docker-build docker-push

## fmt: Run go fmt against code.
fmt:
	go fmt ./pkg/... ./cmd/...

## vet: Run go vet against code.
vet:
	go vet ./pkg/... ./cmd/...

## generate: Generate code
generate:
	go generate ./pkg/... ./cmd/...

## docker-build: Build the docker image.
docker-build:
	docker build ./.. -t ${IMG} -f build/build-image/Dockerfile
