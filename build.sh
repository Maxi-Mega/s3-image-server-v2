#!/bin/bash -eu

echo "Testing ..."

go test -race ./...

echo "Building ..."

VERSION="dev"
ENV_FILE="build-dev.env"
BUILD_TIME=$(date "+%Y-%m-%d %H:%M:%S")
PROD="false"
MINIFY="false"

if [ $# -eq 1 ]; then
    VERSION="$1"
fi

if [ "$VERSION" != "dev" ]; then
    ENV_FILE="build-prod.env"
    PROD="true"
    MINIFY="esbuild"
fi

BINARY_FILENAME="S3ImageServer-$VERSION"

test -f "${ENV_FILE}" && source "${ENV_FILE}" # Load $BASE_URL

echo "Building front-end with base URL '$BASE_URL' ..."

(cd frontend && yarn build --base="$BASE_URL" --minify="$MINIFY" && touch dist/.gitkeep)

echo "Building binary ..."

go generate ./...
go build -ldflags="-X 'main.version=$VERSION' -X 'main.buildTime=$BUILD_TIME' -X 'main.prod=$PROD' -extldflags=-static" -tags osusergo,netgo -o "$BINARY_FILENAME" .

$PROD && upx --best "$BINARY_FILENAME" # Only compress when building for prod

echo "Successfully built the app under the name $BINARY_FILENAME"
