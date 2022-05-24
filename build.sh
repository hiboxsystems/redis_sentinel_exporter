#!/bin/bash

set -eux

VERSION="$(cat VERSION)"

export CGO_ENABLED=0
export GOOS=linux

if [ "$(lscpu | grep Architecture | awk '{print $2}')" = "aarch64" ]; then 
    echo "arm64"
    export GOARCH=arm64
else 
    echo "amd64"
    export GOARCH=amd64
fi

go mod tidy
go mod verify
go test ./...
go build -ldflags "-X github.com/prometheus/common/version.Version=${VERSION}"
