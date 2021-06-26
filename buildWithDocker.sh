#!/bin/bash

set -e

# create/update build image
docker build -t status2-build docker-build/

# move current vendor out of the way
if [ -d vendor ]; then
  mv vendor vendor_local
fi

mkdir -p .docker-build/dep-cache
mkdir -p .docker-build/vendor

mv .docker-build/vendor vendor

docker run -it --rm \
  --user=$(id -u):$(id -g)  \
  -v $(pwd):/go/src/github.com/ktt-ol/status2 \
  -v $(pwd)/.docker-build/dep-cache:/go/pkg/dep \
  status2-build \
 dep ensure -v -vendor-only && ./build.sh

# cleanup
mv vendor .docker-build/vendor
if [ -d vendor_local ]; then
  mv vendor_local vendor
fi
