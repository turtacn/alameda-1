.PHONY: all test license-utils
all: test license-utils

# Run tests
test: generate fmt vet
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build license-utils executable binary
license-utils: generate fmt vet
	go build -ldflags "-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'" -o bin/license-utils github.com/containers-ai/alameda/datahub/tools/license-utils/cmd

.PHONY: run

# Run license-utils in interactive mode
run: generate fmt vet
	go run ./cmd/main.go interactive

.PHONY: fmt vet generate

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...
