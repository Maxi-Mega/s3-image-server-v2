#!/bin/bash -e

LINTER_VERSION=v2.7.2-alpine

if docker volume ls | grep -q s3-image-server-pkg-cache; then
    GO_PKG_CACHE="-v s3-image-server-pkg-cache:/go/pkg"
else
    echo "No volume for go pkg cache found. You can create one with 'docker volume create s3-image-server-pkg-cache'"
fi

echo "Testing ..."

docker run --rm -v "$(pwd)":/app ${GO_PKG_CACHE} -e HOME=/go/pkg \
    -w /app golangci/golangci-lint:${LINTER_VERSION} \
    sh -exc "
    go test ./...
    go test -race ./...
    "

echo "Linting ..."

docker run --rm -v "$(pwd)":/app ${GO_PKG_CACHE} -e HOME=/go/pkg \
    -e GOOS=linux -e GOARCH=amd64 -w /app golangci/golangci-lint:${LINTER_VERSION} \
    sh -ec "
    git config --global --add safe.directory /app
    golangci-lint run
    "


echo "Done !"
