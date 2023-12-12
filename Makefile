include .env
export $(shell sed 's/=.*//' .env)

SHELL=/bin/bash

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

.PHONY: dev
dev:
	@echo -n "Starting application in hot-reload mode ..." ;\
	air ;\
	
