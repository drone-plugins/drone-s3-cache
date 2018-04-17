# drone-s3-cache

[![Build Status](http://beta.drone.io/api/badges/drone-plugins/drone-s3-cache/status.svg)](http://beta.drone.io/drone-plugins/drone-s3-cache)
[![Join the discussion at https://discourse.drone.io](https://img.shields.io/badge/discourse-forum-orange.svg)](https://discourse.drone.io)
[![Drone questions at https://stackoverflow.com](https://img.shields.io/badge/drone-stackoverflow-orange.svg)](https://stackoverflow.com/questions/tagged/drone.io)
[![Go Doc](https://godoc.org/github.com/drone-plugins/drone-s3-cache?status.svg)](http://godoc.org/github.com/drone-plugins/drone-s3-cache)
[![Go Report](https://goreportcard.com/badge/github.com/drone-plugins/drone-s3-cache)](https://goreportcard.com/report/github.com/drone-plugins/drone-s3-cache)
[![](https://images.microbadger.com/badges/image/plugins/s3-cache.svg)](https://microbadger.com/images/plugins/s3-cache "Get your own image badge on microbadger.com")

Drone plugin that allows you to cache directories within the build workspace, this plugin is backed by S3 compatible storages. For the usage information and a listing of the available options please take a look at [the docs](http://plugins.drone.io/drone-plugins/drone-s3-cache/).

## Build

Build the binary with the following commands:

```
go build
```

## Docker

Build the Docker image with the following commands:

```
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -o release/linux/amd64/drone-s3-cache
docker build --rm -t plugins/s3-cache .
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -e PLUGIN_FLUSH=true \
  -e PLUGIN_URL="http://minio.company.com" \
  -e PLUGIN_ACCESS_KEY="myaccesskey" \
  -e PLUGIN_SECRET_KEY="mysecretKey" \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  plugins/s3-cache

docker run --rm \
  -e PLUGIN_RESTORE=true \
  -e PLUGIN_URL="http://minio.company.com" \
  -e PLUGIN_ACCESS_KEY="myaccesskey" \
  -e PLUGIN_SECRET_KEY="mysecretKey" \
  -e DRONE_REPO_OWNER="foo" \
  -e DRONE_REPO_NAME="bar" \
  -e DRONE_COMMIT_BRANCH="test" \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  plugins/s3-cache

docker run -it --rm \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  alpine:latest sh -c "mkdir -p cache && echo 'testing cache' >> cache/test && cat cache/test"

docker run --rm \
  -e PLUGIN_REBUILD=true \
  -e PLUGIN_MOUNT=".bundler" \
  -e PLUGIN_URL="http://minio.company.com" \
  -e PLUGIN_ACCESS_KEY="myaccesskey" \
  -e PLUGIN_SECRET_KEY="mysecretKey" \
  -e DRONE_REPO_OWNER="foo" \
  -e DRONE_REPO_NAME="bar" \
  -e DRONE_COMMIT_BRANCH="test" \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  plugins/s3-cache
```
