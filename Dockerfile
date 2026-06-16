# Build stage
FROM golang:1.26.4-alpine@sha256:7a3e50096189ad57c9f9f865e7e4aa8585ed1585248513dc5cda498e2f41812c AS builder

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
FROM alpine:3.24.1@sha256:bec4ccd3817e7c824eb0388971a0b83fab111d586285511ba0266b77e8dc65a9

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
