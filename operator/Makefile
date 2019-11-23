
# Image URL to use all building/pushing image targets
IMG ?= operator:latest
# Produce CRDs that work back to Kubernetes 1.16 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all
all: manager

# Run tests
.PHONY: test
#test: generate fmt vet manifests
test: fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
.PHONY: manager
#manager: generate fmt vet
manager: fmt vet
	go build -ldflags "-X main.VERSION=`git rev-parse --abbrev-ref HEAD`-`git rev-parse --short HEAD``git diff --quiet || echo '-dirty'` -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'" -o bin/manager github.com/containers-ai/alameda/operator/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
#run: generate fmt vet manifests
run: fmt vet manifests
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
.PHONY: install
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD
.PHONY: manifests
manifests: controller-gen
	go run github.com/containers-ai/alameda/vendor/sigs.k8s.io/controller-tools/cmd/controller-gen crd

# Run go fmt against code
.PHONY: fmt
fmt:
	go fmt ./...

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

# Generate code
.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
.PHONY: docker-build
docker-build: test
	docker build ./.. -t ${IMG} -f Dockerfile
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
.PHONY: docker-push
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.2 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
