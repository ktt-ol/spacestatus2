#!/bin/bash

set -e

# create/update build image
docker build -t status2-build docker-build/

mkdir -p .docker-build/dep-cache
mkdir -p .docker-build/vendor

docker run -it --rm \
  --env GOCACHE=/go/src/github.com/ktt-ol/status2/.docker-build/build-cache \
  --user=$(id -u):$(id -g)  \
  -v $(pwd):/go/src/github.com/ktt-ol/status2 \
  -v $(pwd)/.docker-build/dep-cache:/go/pkg/dep \
  status2-build \
  /bin/sh -c "dep ensure -v -vendor-only && ./build.sh"
