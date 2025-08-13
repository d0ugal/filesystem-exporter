# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o filesystem-exporter ./cmd/main.go

# Final stage
FROM alpine:3.22.1

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