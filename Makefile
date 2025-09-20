.PHONY: test

test:
	go test ./... -v

test-cover:
	go test ./... -cover

test-cover-html:
	go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html

build:
	go build -o multipass-exporter ./cmd/multipass-exporter

build-all:
	go build -o multipass-exporter-linux-amd64 ./cmd/multipass-exporter
	GOARCH=amd64 go build -o multipass-exporter-amd64 ./cmd/multipass-exporter
	GOARCH=arm64 go build -o multipass-exporter-arm64 ./cmd/multipass-exporter
	GOARCH=386 go build -o multipass-exporter-386 ./cmd/multipass-exporter

run:
	go run ./cmd/multipass-exporter

clean:
	rm -f multipass-exporter multipass-exporter-* coverage.out coverage.html

fmt:
	go fmt ./...

deps:
	go mod tidy
	go mod download

help:
	@echo "Available commands:"
	@echo "  test          - Run tests with verbose output"
	@echo "  test-cover    - Run tests with coverage"
	@echo "  test-cover-html - Run tests and generate HTML coverage report"
	@echo "  build         - Build the binary for current Linux platform"
	@echo "  build-all     - Build binaries for multiple Linux architectures"
	@echo "  run           - Run the application (uses defaults or config.yaml)"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  help          - Show this help message"
