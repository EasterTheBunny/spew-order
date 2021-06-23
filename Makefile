API=spew-order

GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

GOFILES = $(shell find . -name '*.go')
GOPACKAGES = $(shell go list ./...)

all: dependencies build

dependencies:
	go mod download

test: dependencies
	@go test -v -tags !integration $(GOPACKAGES)

benchmark: dependencies fmt
	@go test $(GOPACKAGES) -bench=.

fmt:
	gofmt -w .

openapi:
	go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen && \
	oapi-codegen --config pkg/api/config.yaml pkg/api/openapi.yaml > pkg/api/api.gen.go

build: openapi
	go build -o $(GOBIN)/$(API) ./cmd/$(API)/*.go || exit

build-google-storage-test:
	go build -o $(GOBIN)/tools/google-storage-test ./cmd/tools/google-storage-test/*.go || exit

build-all: fmt test build

run: build-all
	./bin/$(API)

default: build

.PHONY: build project fmt