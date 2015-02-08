FROM golang:latest
MAINTAINER peter.edge@gmail.com

RUN mkdir -p /go/src/github.com/peter-edge/go-exec
WORKDIR /go/src/github.com/peter-edge/go-exec
ADD Makefile /go/src/github.com/peter-edge/go-exec/
RUN make clean
RUN make testdeps
ADD . /go/src/github.com/peter-edge/go-exec/
