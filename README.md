# S3 Image Server V2

### Web UI for browsing images from S3 buckets

## How to build

### From source

1. Clone to repo
2. Install the dependencies (`go mod download`)
3. Copy build-dev.env.sample & build-prod.env.sample to build-dev.env & build-prod.env
4. Run `./build.sh <version>`, with `<version>` being 'dev' for unminified, or any other value for a minified and compressed build.

### From builder image

1. Clone the repo
2. Download the builder image tar from the release page
3. Run `docker load -i s3-image-server-builder-vX.X.X.tar.gz`
4. Run `docker build -f offline.Dockerfile -t s3-image-server:vX.X.X --build-arg BASE_URL=/ .`
    (Note that the `BASE_URL` argument is optional, and defaults to `/`.)
