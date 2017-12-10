.PHONY: build install test doc fmt lint vet

export GOPATH

VERSION ?= $(shell git describe --tags)

default: build

build: install vet compile
	echo "Build"

compile:
	go build -v -o ./bin/beagle \
	-ldflags "-X github.com/blent/beagle/src/core.Version=${VERSION}" \
	./src/main.go

install:
	glide install

test:
	go test ./src/... -v

doc:
	godoc -http=:6060 -index

# http://golang.org/cmd/go/#hdr-Run_gofmt_on_package_sources
fmt:
	go fmt ./src/...

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	golint ./src

# http://godoc.org/code.google.com/p/go.tools/cmd/vet
# go get code.google.com/p/go.tools/cmd/vet
vet:
	go vet ./src/...