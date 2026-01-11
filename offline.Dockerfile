FROM s3-image-server-builder:v2.2.7 AS builder

ARG BASE_URL="/"

WORKDIR /go/src/app

COPY config/ config/
COPY frontend/ frontend/
COPY internal/ internal/
COPY resources/ resources/
COPY utils/ utils/
COPY *.go go.* ./
COPY gqlgen.yml schema.graphqls build.sh ./

ENV GOPROXY=off

RUN BASE_URL="$BASE_URL" ./build.sh "$VERSION"

RUN mv "S3ImageServer-$VERSION" S3ImageServer

FROM scratch

COPY --from=builder /go/src/app/S3ImageServer /app/S3ImageServer

ENTRYPOINT ["/app/S3ImageServer"]
