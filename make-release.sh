#!/bin/bash

VERSION=0.0.1
GIT_VERSION=$(git describe --always --long --dirty)

GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$GIT_VERSION" -gcflags "all=-trimpath=/Users/dan/go" -o rivet-$VERSION-darwin-amd64
GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$GIT_VERSION" -gcflags "all=-trimpath=/Users/dan/go" -o rivet-$VERSION-linux-amd64
