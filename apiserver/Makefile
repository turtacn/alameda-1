# compute source code references
GIT_COMMIT = $(shell git rev-parse --short HEAD)
_TMP_STR = $(shell git show-ref --abbrev --head | grep "^$(GIT_COMMIT) " | grep -v HEAD | head -1 | cut -c8-)
GIT_REF ?= $(shell [ "$(_TMP_STR)" = "" ] && echo "HEAD" || echo $(_TMP_STR))
GIT_DIRTY = $(shell git diff --quiet || echo '-dirty')
CODE_VERSION = "$(GIT_REF)@$(GIT_COMMIT)$(GIT_DIRTY)"

# Image URL to use all building/pushing image targets
IMG ?= apiserver:v0.0.1

SRC_DIR = $(shell pwd)
INSTALL_ROOT = $(SRC_DIR)/install_root
PRODUCT_ROOT = /opt/alameda/apiserver
DEST_PREFIX = $(INSTALL_ROOT)$(PRODUCT_ROOT)
######################################################################

.PHONY: all test apiserver
all: test apiserver
test2:
	@echo GIT_REFS=$(GIT_REFS)=
	@echo GIT_REF=$(GIT_REF)=
	@echo version=$(CODE_VERSION)=
## test: Run tests.
test: generate fmt vet
	go test ./pkg/... ./cmd/... -coverprofile cover.out

## api-server: Build API server executable binary.
apiserver: generate fmt vet binaries
	@echo Start building API server binary...

binaries:
	@(echo ===binaries: build env start===; env | sort; echo ===binaries: build env end===)
	@(echo ===git references start===; git status; git show-ref --abbrev --head; echo ===git references end===)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
	  -ldflags "-X main.VERSION=$(CODE_VERSION) -X 'main.BUILD_TIME=`date`' -X 'main.GO_VERSION=`go version`'" \
	  -a -o build/build-image/bin/apiserver github.com/containers-ai/alameda/apiserver/cmd

install_dir:
	mkdir -p $(INSTALL_ROOT)/etc/alameda $(DEST_PREFIX)/bin $(DEST_PREFIX)/etc

install: install_dir
	cp -fv $(SRC_DIR)/build/build-image/etc/apiserver.toml $(DEST_PREFIX)/etc/
	ln -sfv $(PRODUCT_ROOT)/etc $(INSTALL_ROOT)/etc/alameda/apiserver
	cp -fv build/build-image/bin/apiserver $(DEST_PREFIX)/bin/
	cd $(INSTALL_ROOT); tar -czvf $(SRC_DIR)/install_root.tgz .; cd -

clean:
	rm -fv build/build-image/bin/apiserver install_root.tgz

clobber: clean
	rm -rf install_root

build: binaries install

.PHONY: run

## run: Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
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
docker-build-alpine:
	@(echo ===docker-build-alpine: build env start===; env | sort; echo ===docker-build-alpine: build env end===)
	docker build --build-arg CODE_VERSION=$(CODE_VERSION) -t ${IMG} -f Dockerfile ..

docker-build-ubi:
	@(echo ===docker-build-ubi: build env start===; env | sort; echo ===docker-build-ubi: build env end===)
	docker build --build-arg CODE_VERSION=$(CODE_VERSION) -t ${IMG} -f Dockerfile.ubi ..

docker-build: docker-build-ubi
