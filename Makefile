# Add GOPATH/bin to PATH for this Makefile
export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: test lint

test:
	go test ./... -v

test-cover:
	go test ./... -cover

test-cover-html:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html

build:
	go build -o multipass-exporter ./cmd/multipass-exporter

build-all:
	GOARCH=amd64 go build -o multipass-exporter-linux-amd64 ./cmd/multipass-exporter
	GOARCH=arm64 go build -o multipass-exporter-linux-arm64 ./cmd/multipass-exporter

run:
	go run ./cmd/multipass-exporter

clean:
	rm -f multipass-exporter multipass-exporter-* coverage.out coverage.html

lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1; \
	fi
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout=5m

lint-fast:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1; \
	fi
	@echo "Running golangci-lint (fast mode)..."
	@golangci-lint run --fast

fmt:
	go fmt ./...

deps:
	go mod tidy
	go mod download

help:
	@echo "Available commands:"
	@echo "  test             - Run tests with verbose output"
	@echo "  test-cover       - Run tests with coverage"
	@echo "  test-cover-html  - Run tests and generate HTML coverage report"
	@echo "  build            - Build the binary for current Linux platform"
	@echo "  build-all        - Build binaries for multiple Linux architectures"
	@echo "  run              - Run the application (uses defaults or config.yaml)"
	@echo "  clean            - Clean build artifacts"
	@echo "  lint             - Run linter (auto-installs if needed)"
	@echo "  lint-fast        - Run linter in fast mode"
	@echo "  fmt              - Format code"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  help             - Show this help message"
