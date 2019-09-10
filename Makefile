.PHONY: build clean test help default

BIN_NAME=rivet

VERSION := $(shell grep "const Version " version/version.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "-dirty" || true)
BUILD_DATE=$(shell date --utc '+%Y-%m-%d-%H:%M:%S')

default: test

help:
	@echo 'Management commands for rivet:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make test            Run tests on a compiled project.'
	@echo '    make clean           Clean the directory tree.'
	@echo

build:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build \
			-ldflags "-X github.com/danlamanna/rivet/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/danlamanna/rivet/version.BuildDate=${BUILD_DATE}" \
		  -gcflags "all=-trimpath=${GOPATH}" \
	    -o bin/${BIN_NAME}-darwin
	GOOS=linux GOARCH=amd64 go build \
			-ldflags "-X github.com/danlamanna/rivet/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/danlamanna/rivet/version.BuildDate=${BUILD_DATE}" \
		  -gcflags "all=-trimpath=${GOPATH}" \
	    -o bin/${BIN_NAME}-linux

clean:
	rm -rf bin/

test:
	go test ./...
