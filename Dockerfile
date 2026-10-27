# Build stage
FROM golang:1.27.3-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb AS builder

WORKDIR /app

# Install git for version detection
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with version information
# Accept build args for version info, fall back to git describe if not provided
ARG VERSION
ARG COMMIT
ARG BUILD_DATE

RUN VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")} && \
    COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")} && \
    BUILD_DATE=${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")} && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w \
        -X filesystem-exporter/internal/version.Version=$VERSION \
        -X filesystem-exporter/internal/version.Commit=$COMMIT \
        -X filesystem-exporter/internal/version.BuildDate=$BUILD_DATE" \
    -o filesystem-exporter ./cmd/main.go

# Final stage
FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/filesystem-exporter .

# Create config directory
RUN mkdir -p /root/config

# Expose port
EXPOSE 8080

# Set default config path
ENV CONFIG_PATH=/root/config.yaml

# Run the application
CMD ["./filesystem-exporter"]
