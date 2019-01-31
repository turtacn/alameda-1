
# Image URL to use all building/pushing image targets
IMG ?= operator:latest

.PHONY: all test manager test-log 
all: test manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... ./services/... -coverprofile cover.out

# Run tests
test-log: generate fmt vet manifests
	go test ./pkg/utils/log/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -ldflags "-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``[[ ! -z $(git status -s -uno) ]] && echo \"-dirty\"` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'" -o bin/manager github.com/containers-ai/alameda/operator/cmd/manager

.PHONY: run install deploy manifests

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD
manifests:
	go run github.com/containers-ai/alameda/vendor/sigs.k8s.io/controller-tools/cmd/controller-gen crd

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
	docker build ./.. -t ${IMG} -f Dockerfile
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}
