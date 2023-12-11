#!make
include .env

.PHONY: default
default: test lint

.PHONY: build
build:
	@echo "Building application ..."
	CGO_ENABLED=0 GOARCH=amd64 go build -o build/collection ./cmd/collection

.PHONY: test
test:
	go test ./... -cover -race

.PHONY: lint
lint:
	golangci-lint run

.PHONY: watch
watch:
	templ generate -watch

.PHONY: air
air:
	air
