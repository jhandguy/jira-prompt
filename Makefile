.PHONY: tidy lint vet staticcheck check build test

VERSION:=$(shell git describe --tags 2> /dev/null || git rev-parse HEAD)

tidy:
	go mod tidy --diff

lint:
	revive -config revive/config.toml -formatter friendly ./...

vet:
	go vet ./...

staticcheck:
	staticcheck ./...

check: tidy lint vet staticcheck

build:
	go build -ldflags "-X main.version=${VERSION}" -o jp

test:
	go test -v ./... -race -covermode=atomic -shuffle=on
