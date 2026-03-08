#!/bin/bash -e

LINTER_VERSION=v2.11.1-alpine

VOLUME_NAME="s3-image-server-pkg-cache"
if docker volume ls | grep -q $VOLUME_NAME; then
    GO_PKG_CACHE="-v $VOLUME_NAME:/go/pkg"
else
    echo "No volume for go pkg cache found. You can create one with 'docker volume create $VOLUME_NAME'"
fi

echo "Testing ..."

docker run --rm -v "$(pwd)":/app ${GO_PKG_CACHE} -e HOME=/go/pkg \
    -w /app golangci/golangci-lint:${LINTER_VERSION} \
    sh -exc "
    go test -coverpkg=./... -coverprofile=coverage.out -covermode=count ./...
    go test -race ./...
    go tool cover -html=coverage.out -o coverage.html
    "

echo "Linting ..."

docker run --rm -v "$(pwd)":/app ${GO_PKG_CACHE} -e HOME=/go/pkg \
    -e GOOS=linux -e GOARCH=amd64 -w /app golangci/golangci-lint:${LINTER_VERSION} \
    sh -ec "
    git config --global --add safe.directory /app
    golangci-lint run
    "

echo "Done !"
