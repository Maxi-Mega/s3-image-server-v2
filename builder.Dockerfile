# syntax=docker/dockerfile:1.19

FROM golang:1.25-alpine3.22 AS backend-builder

WORKDIR /go/src/app

COPY --exclude=frontend . .

RUN go mod download
RUN go generate ./...

FROM golang:1.25-alpine3.22

COPY --from=backend-builder /go/pkg /go/pkg

WORKDIR /go/src/app/frontend

COPY frontend/package.json frontend/yarn.lock frontend/.yarnrc.yml ./
COPY frontend/.yarn/ .yarn/

RUN apk update && apk add --no-cache \
    nodejs \
    npm \
    git \
    bash

RUN npm install -g corepack
RUN corepack enable

RUN yarn config set --home enableTelemetry 0

RUN yarn install --immutable --inline-builds

ARG VERSION="dev"
ENV VERSION="$VERSION"
