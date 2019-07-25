FROM golang:1.12-alpine

LABEL maintainer = "saguywalker at protonmail dot com"

RUN apk update && apk add --no-cache git \
    && go get -u gitlab.com/saguywalker/add2git-lfs \
    && go get -u github.com/labstack/echo \
    && go get -u github.com/labstack/echo/middleware \
    && cd $GOPATH/src/gitlab.com/saguywalker/add2git-lfs \
    && go build

WORKDIR $GOPATH/src/gitlab.com/saguywalker/add2git-lfs

EXPOSE 12358

cmd ["./add2git-lfs"]