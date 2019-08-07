.PHONY: build test

build:
	go build ./cmd/server

test:
	go test ./...
