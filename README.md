# S3 Image Server V2

### Web UI for browsing images from S3 buckets

## How to build

1. Clone to repo
2. Install the dependencies (`go mod download`)
3. Copy build-dev.env.sample & build-prod.env.sample to build-dev.env & build-prod.env
4. Run `./build.sh <version>`, with `<version>` being 'dev' for unminified, or any other value for a minified build.
