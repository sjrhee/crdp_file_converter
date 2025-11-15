.PHONY: build test clean run help fmt lint build-cross

# Variables
BINARY_NAME=crdp-file-converter
GO=go
GOFLAGS=-v

# Cross-compilation variables
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
BUILD_DIR := bin

help:
	@echo "Available targets:"
	@echo "  build           - Build the project for current OS"
	@echo "  build-cross     - Build for all platforms (linux, darwin, windows)"
	@echo "  test            - Run tests"
	@echo "  test-cov        - Run tests with coverage"
	@echo "  clean           - Clean build artifacts"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run linter (requires golangci-lint)"
	@echo "  run             - Run the application (requires arguments)"
	@echo "  install         - Install dependencies"
	@echo "  help            - Show this help message"

build:
	$(GO) build -o $(BINARY_NAME) ./cmd

build-cross:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then OUTPUT=$$OUTPUT.exe; fi; \
		echo "Building $$GOOS/$$GOARCH -> $$OUTPUT"; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GO) build -o $$OUTPUT ./cmd; \
	done
	@echo "Build complete! Binaries in $(BUILD_DIR)/"

test:
	$(GO) test $(GOFLAGS) ./...

test-cov:
	$(GO) test -cover ./...
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f e[0-9][0-9]_*\.csv d[0-9][0-9]_*\.csv
	rm -f coverage.out coverage.html
	$(GO) clean

fmt:
	$(GO) fmt ./...
	gofmt -w .

lint:
	golangci-lint run ./...

install:
	$(GO) mod download
	$(GO) mod tidy

run: build
	./$(BINARY_NAME) sample_data.csv --column 1 --operation protect --skip-header
