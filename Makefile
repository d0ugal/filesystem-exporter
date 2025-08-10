.PHONY: build test run clean docker-build docker-run docker-push help

# Default target
.DEFAULT_GOAL := help

# Build variables
BINARY_NAME := filesystem-exporter
DOCKER_IMAGE := ghcr.io/d0ugal/filesystem-exporter
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run with custom config
run-config:
	@echo "Running $(BINARY_NAME) with custom config..."
	./$(BINARY_NAME) -config config.yaml

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -f coverage.html

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):latest .
	docker tag $(DOCKER_IMAGE):latest $(DOCKER_IMAGE):$(VERSION)

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 \
		-v /:/host:ro \
		-v $(PWD)/config.yaml:/root/config.yaml:ro \
		$(DOCKER_IMAGE):latest

# Build and run in Docker
docker: docker-build docker-run

# Push Docker image
docker-push:
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):$(VERSION)

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Generate go.sum
sum:
	@echo "Generating go.sum..."
	go mod tidy
	go mod verify

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Check code quality
check: fmt lint test
	@echo "Code quality check completed"

# Release build (optimized)
release-build:
	@echo "Building release version..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BINARY_NAME) ./cmd/main.go

# Run CI checks locally
ci: fmt lint test security-scan
	@echo "CI checks completed successfully"

# Run security scan
security-scan:
	@echo "Running security scan..."
	@if command -v trivy >/dev/null 2>&1; then \
		trivy fs .; \
	else \
		echo "Trivy not found. Install with: go install github.com/aquasecurity/trivy/cmd/trivy@latest"; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  run            - Build and run the application"
	@echo "  run-config     - Run with custom config"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker         - Build and run Docker container"
	@echo "  docker-push    - Push Docker image to registry"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  deps           - Install dependencies"
	@echo "  sum            - Generate go.sum"
	@echo "  install-tools  - Install development tools"
	@echo "  check          - Run fmt, lint, and test"
	@echo "  release-build  - Build optimized release version"
	@echo "  ci             - Run CI checks locally"
	@echo "  security-scan  - Run security scan"
	@echo "  help           - Show this help message" 