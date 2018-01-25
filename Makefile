.DEFAULT_GOAL := build
.PHONY: test.exmaples

SOURCES        = $(shell find . -name '*.go')
GOPKGS        = $(shell go list ./... | grep -v /vendor/)

build: $(SOURCES)
	go build .

fmt:
	go fmt $(GOPKGS)

check:
	golint $(GOPKGS)
	go vet $(GOPKGS)

test: build
	go test -v ./

test.exmaples: build
	go test -v ./test-examples/...

test.all: test test.exmaples