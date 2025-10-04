ifneq (,$(wildcard .env))
    include .env
    export
endif

SHELL := /bin/bash

# Build variables
APP_NAME := collection
BUILD_DIR := build
TMP_DIR := tmp
CMD_DIR := cmd/collection
MAIN_FILE := $(CMD_DIR)/main.go
GO_BUILD_FLAGS := CGO_ENABLED=0 GOARCH=amd64
BUILD_OUTPUT := $(BUILD_DIR)/$(APP_NAME)

.PHONY: dev
dev:
	@echo "🚀 Starting development environment with hot-reload on port 7331..."
	@make live

.PHONY: live
live:
	@make -j5 live/templ live/tailwind live/server live/assets

.PHONY: live/templ
live/templ:
	@echo "👁️  Starting templ watcher..."
	@templ generate --watch --proxy="http://localhost:9080" --open-browser=false -v

.PHONY: live/tailwind
live/tailwind:
	@echo "🎨 Starting Tailwind watcher..."
	@npx @tailwindcss/cli -i ./assets/css/tailwind-input.css -o ./assets/css/tailwind-output.gen.css --minify --watch

.PHONY: live/server
live/server:
	@echo "🔥 Starting server with hot-reload..."
	@go run github.com/air-verse/air@latest \
		--build.cmd "go build -o $(TMP_DIR)/bin/main $(MAIN_FILE) && templ generate --notify-proxy" \
		--build.bin "$(TMP_DIR)/bin/main" \
		--build.delay "100" \
		--build.exclude_dir "node_modules" \
		--build.include_ext "go" \
		--build.stop_on_error "false" \
		--misc.clean_on_exit true

.PHONY: live/assets
live/assets:
	@echo "📁 Starting assets watcher..."
	@go run github.com/air-verse/air@latest \
		--build.cmd "templ generate --notify-proxy" \
		--build.bin "true" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_dir "assets" \
		--build.include_ext "js,css"

.PHONY: build
build: clean
	@echo "🔨 Building application..."
	@mkdir -p $(BUILD_DIR)
	@$(GO_BUILD_FLAGS) go build -ldflags="-s -w" -o $(BUILD_OUTPUT) $(MAIN_FILE)
	@echo "✅ Build complete: $(BUILD_OUTPUT)"

.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test ./... -cover -race

.PHONY: lint
lint:
	@echo "🔍 Running linter..."
	@golangci-lint run

.PHONY: assets/build
assets/build:
	@echo "🎨 Building frontend assets..."
	@templ generate
	@npx @tailwindcss/cli -i ./assets/css/tailwind-input.css -o ./assets/css/tailwind-output.gen.css --minify
	@echo "✅ Assets built"

.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(TMP_DIR) coverage.out coverage.html
	@find . -name "*.gen.go" -delete
	@find . -name "*.gen.css" -delete
	@echo "✅ Clean complete"

.PHONY: deps
deps:
	@echo "📦 Managing dependencies..."
	@go mod download
	@go mod tidy
	@go mod verify

.PHONY: deps/update
deps/update:
	@echo "⬆️  Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@go mod vendor
	@echo "✅ Dependencies updated"
