.PHONY: help build test lint clean

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  test     - Run tests"
	@echo "  lint     - Run golangci-lint using official container"
	@echo "  clean    - Clean build artifacts"

# Build the application
build:
	go build -v -ldflags="-s -w" -o filesystem-exporter ./cmd

# Run tests
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run golangci-lint using official container
lint:
	docker run --rm \
		-v "$(PWD):/app" \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run

# Clean build artifacts
clean:
	rm -f filesystem-exporter coverage.txt 