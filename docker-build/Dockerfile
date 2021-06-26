FROM golang:1.13.15-alpine3.12

RUN set -xe \
    && apk add git \
    && go get -u github.com/golang/dep/cmd/dep

RUN mkdir -p /go/src/github.com/ktt-ol/status2
WORKDIR /go/src/github.com/ktt-ol/status2
