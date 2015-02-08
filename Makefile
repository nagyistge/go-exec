.PHONY: all deps build test cov install container containertest doc clean

all: test install

deps:
	go get -d -v code.google.com/p/go-uuid

build: deps
	go build ./...

test: deps
	go test -test.v ./...

cov: deps
	go get -v github.com/axw/gocov/gocov
	gocov test | gocov report

install: deps
	go install ./...

container: deps
	docker build -t peteredge/goexec .

containertest: container
	docker run peteredge/goexec make test

doc:
	go get -v github.com/robertkrimen/godocdown/godocdown
	cp .readme.header README.md
	godocdown | tail -n +6 >> README.md

clean:
	go clean -i ./...
