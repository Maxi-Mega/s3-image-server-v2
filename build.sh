#!/bin/bash -eu

VERSION="dev"
PROD="false"
MINIFY="false"

if [ $# -eq 1 ]; then
    VERSION="$1"
fi

if [ "$VERSION" != "dev" ]; then
    PROD="true"
    MINIFY="esbuild"
fi

BINARY_FILENAME="S3ImageServer-$VERSION"

BASE_URL="/"

echo "Building front-end with base URL '$BASE_URL' ..."

(cd frontend && yarn build --base="$BASE_URL" --minify="$MINIFY" && touch dist/.gitkeep)

echo "Building binary ..."

go generate ./...
go build -ldflags="-X 'main.version=$VERSION' -X 'main.prod=$PROD' -extldflags=-static" -tags osusergo,netgo -o "$BINARY_FILENAME" .

echo "Successfully built the app under the name $BINARY_FILENAME"
