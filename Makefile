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

live/templ:
	templ generate --watch --proxy="http://localhost:9080" --open-browser=false -v

live/tailwind:
	npx tailwindcss -i ./assets/css/tailwind-input.css -o ./assets/css/tailwind-output.gen.css --minify --watch

live/esbuild:
	npx esbuild js/index.ts --bundle --outdir=assets/ --watch

live/server:
	go run github.com/air-verse/air@latest \
	--build.cmd "go build -o tmp/bin/main ./cmd/collection/main.go && templ generate --notify-proxy" \
	--build.bin "tmp/bin/main" \
	--build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

live/assets:
	go run github.com/air-verse/air@latest \
	--build.cmd "templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "assets" \
	--build.include_ext "js,css"

live: 
	make -j5 live/templ live/tailwind live/server live/assets

.PHONY: dev
dev:
	@echo -n "Starting application in hot-reload mode on port 7331 ..." ;\
	make live

	
