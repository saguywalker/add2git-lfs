FROM golang:1.12-alpine

LABEL MAINTAINER = saguywalker@protonmail.com

RUN apk update && apk add --no-cache git \
    && go get -u gitlab.com/saguywalker/add2git-lfs \
    && go get -u github.com/labstack/echo \
    && go get -u github.com/labstack/echo/middleware

WORKDIR $GOPATH/src/gitlab.com/saguywalker/add2git-lfs

RUN go run main.go