#!/bin/bash -eu

VERSION="dev"

if [ $# -eq 1 ]; then
    VERSION="$1"
fi

BINARY_FILENAME="S3ImageServerV2-$VERSION"

echo "Building front-end ..."

(cd src/frontend && yarn build)

echo "Building binary ..."

go build -ldflags="-X 'main.version=$VERSION' -extldflags=-static" -tags osusergo,netgo -o "$BINARY_FILENAME" ./src

echo "Successfully built the app under the name $BINARY_FILENAME"
