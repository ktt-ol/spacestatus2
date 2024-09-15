#!/bin/sh

GIT_VERSION=$(git describe --always --abbrev=8  --dirty --broken)

go version
CGO_ENABLED=0 go build -ldflags "-X main.buildVersion=${GIT_VERSION}" cmd/spaceStatus/status2.go
