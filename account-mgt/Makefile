API_SERVER_VERSION=$(shell cat build/build-image/VERSION)
API_SERVER_COMMITID=$(shell git rev-parse HEAD)

# Image URL to use all building/pushing image targets
IMG ?= openldap:1.2.4

.PHONY: all docker-build

all: docker-build

.PHONY: docker-build

## docker-build: Build the docker image.
docker-build:
	docker build build/build-image/ -t ${IMG} -f build/build-image/Dockerfile
