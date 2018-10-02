TAG?=latest
NAME?=alameda
GOOS?=linux

.PHONY: all
all: build

.PHONY: deps
deps:
	go get -u github.com/golang/dep/cmd/dep

.PHONY: build
build: clean deps
	dep ensure
	GOOS=$(GOOS) go build -o ${NAME}

.PHONY: docker 
docker:
	docker build -t ${NAME}:${TAG} .

.PHONY: release
release: build docker

.PHONY: clean
clean:
	rm -f ${NAME}	
