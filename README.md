# drone-s3-cache
Caches build artifacts to S3 compatible storage backends

## Build

Build the binary with the following commands:

```
go build
go test
```

## Docker

Build the docker image with the following commands:

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo
docker build --rm=true -t drone-plugins/drone-s3-cache .
```

## Usage

Execute from the working directory:

```
# Rebuild cache
docker run --rm \
  -e PLUGIN_URL=http://minio.company.com \
  -e PLUGIN_ACCESS_KEY=myaccesskey \
  -e PLUGIN_SECRET_KEY=mysecretKey \
  -e PLUGIN_MOUNT=.bundler \
  drone-plugins/drone-s3-cache --rebuild

# Restore from cache
docker run --rm \
  -e PLUGIN_URL=http://minio.company.com \
  -e PLUGIN_ACCESS_KEY=myaccesskey \
  -e PLUGIN_SECRET_KEY=mysecretKey \
  drone-plugins/drone-s3-cache --restore

# Flush cache
docker run --rm \
  -e PLUGIN_URL=http://minio.company.com \
  -e PLUGIN_ACCESS_KEY=myaccesskey \
  -e PLUGIN_SECRET_KEY=mysecretKey \
  drone-plugins/drone-s3-cache --flush
```
