#!/bin/bash -eu

VERSION="dev"

if [ $# -eq 1 ]; then
    VERSION="$1"
fi

BINARY_FILENAME="S3ImageServerV2-$VERSION"

BASE_URL="/"

echo "Building front-end with base URL '$BASE_URL' ..."

(cd frontend && yarn build --base="$BASE_URL" && touch dist/.gitkeep)

echo "Building binary ..."

go generate ./...
go build -ldflags="-X 'main.version=$VERSION' -extldflags=-static" -tags osusergo,netgo -o "$BINARY_FILENAME" .

echo "Successfully built the app under the name $BINARY_FILENAME"
