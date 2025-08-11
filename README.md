# Filesystem Exporter

A lightweight, high-performance Go-based metrics collection service for monitoring filesystem usage and directory sizes. This service provides Prometheus metrics for comprehensive filesystem monitoring and directory size tracking across any Linux environment.

## Features

- **Filesystem Metrics**: Collects filesystem usage statistics using `df` command
- **Directory Metrics**: Collects directory sizes using `du` command with configurable depth
- **High Performance**: Efficient collection with individual tickers per filesystem/directory
- **Configurable**: YAML-based configuration with human-readable duration formats
- **Prometheus Integration**: Exports metrics in Prometheus format
- **Health Monitoring**: Built-in health check endpoint
- **Graceful Shutdown**: Proper signal handling and cleanup
- **Structured Logging**: JSON and text logging with configurable levels
- **Retry Logic**: Exponential backoff retry for failed collections

## Quick Start

### Using Docker

```bash
# Pull the image
docker pull ghcr.io/d0ugal/filesystem-exporter:latest

# Run with default configuration
docker run -d \
  --name filesystem-exporter \
  -p 8080:8080 \
  -v /:/host:ro \
  -v $(pwd)/config.yaml:/root/config.yaml:ro \
  ghcr.io/d0ugal/filesystem-exporter:latest

# Access metrics
curl http://localhost:8080/metrics
```

### Using Docker Compose

```yaml
version: '3.8'
services:
  filesystem-exporter:
    image: ghcr.io/d0ugal/filesystem-exporter:latest
    ports:
      - "8080:8080"
    volumes:
      - /:/host:ro
      - ./config.yaml:/root/config.yaml:ro
    restart: unless-stopped
```

### From Source

```bash
# Clone the repository
git clone https://github.com/d0ugal/filesystem-exporter.git
cd filesystem-exporter

# Build and run
make build
./filesystem-exporter
```

## Configuration

The application can be configured via `config.yaml` file or environment variables. Copy `config.example.yaml` to `config.yaml` and customize for your environment.

### Configuration Methods

1. **YAML Configuration File** (default): Use `config.yaml` for complex configurations
2. **Environment Variables**: Use environment variables for simple configurations or containerized deployments
3. **Hybrid Mode**: Use environment variables to override specific settings

### Environment Variable Configuration

For containerized deployments, you can configure the application entirely through environment variables, similar to Prometheus. This is especially useful for Kubernetes, Docker Compose, and other container orchestration systems.

#### Environment Variable Format

All environment variables are prefixed with `FILESYSTEM_EXPORTER_`:

- `FILESYSTEM_EXPORTER_CONFIG_FROM_ENV=true` - Force environment-only configuration mode
- `FILESYSTEM_EXPORTER_SERVER_HOST` - Server host (default: "0.0.0.0")
- `FILESYSTEM_EXPORTER_SERVER_PORT` - Server port (default: 8080)
- `FILESYSTEM_EXPORTER_LOG_LEVEL` - Log level: debug, info, warn, error (default: "info")
- `FILESYSTEM_EXPORTER_LOG_FORMAT` - Log format: json, text (default: "json")
- `FILESYSTEM_EXPORTER_METRICS_DEFAULT_INTERVAL` - Default collection interval (default: "5m")

#### Filesystems Configuration

Configure filesystems using the `FILESYSTEM_EXPORTER_FILESYSTEMS` environment variable:

```
FILESYSTEM_EXPORTER_FILESYSTEMS=name1:mount1:device1[:interval1],name2:mount2:device2[:interval2]
```

**Format**: `name:mount:device[:interval]`

**Examples**:
```bash
# Basic filesystem
FILESYSTEM_EXPORTER_FILESYSTEMS=root:/:sda1

# Multiple filesystems with intervals
FILESYSTEM_EXPORTER_FILESYSTEMS=root:/:sda1:1m,data:/data:sdb1:2m

# Host filesystem for Docker
FILESYSTEM_EXPORTER_FILESYSTEMS=host:/host:sda1:1m
```

#### Directories Configuration

Configure directories using the `FILESYSTEM_EXPORTER_DIRECTORIES` environment variable:

```
FILESYSTEM_EXPORTER_DIRECTORIES=name1:path1:levels1[:interval1],name2:path2:levels2[:interval2]
```

**Format**: `name:path:levels[:interval]`

**Examples**:
```bash
# Basic directory
FILESYSTEM_EXPORTER_DIRECTORIES=home:/home:1

# Multiple directories with intervals
FILESYSTEM_EXPORTER_DIRECTORIES=home:/home:1:10m,logs:/var/log:0:5m,apps:/opt/apps:2:30m

# Docker host directories
FILESYSTEM_EXPORTER_DIRECTORIES=containers:/host/var/lib/docker/containers:1:5m,volumes:/host/var/lib/docker/volumes:1:5m
```

#### Docker Compose Example

```yaml
version: '3.8'
services:
  filesystem-exporter:
    image: ghcr.io/d0ugal/filesystem-exporter:latest
    ports:
      - "8080:8080"
    volumes:
      - /:/host:ro
    environment:
      - FILESYSTEM_EXPORTER_CONFIG_FROM_ENV=true
      - FILESYSTEM_EXPORTER_FILESYSTEMS=root:/host:sda1:1m,data:/host/data:sdb1:2m
      - FILESYSTEM_EXPORTER_DIRECTORIES=home:/host/home:1:10m,logs:/host/var/log:0:5m
    restart: unless-stopped
```

#### Kubernetes Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: filesystem-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: filesystem-exporter
  template:
    metadata:
      labels:
        app: filesystem-exporter
    spec:
      containers:
      - name: filesystem-exporter
        image: ghcr.io/d0ugal/filesystem-exporter:latest
        ports:
        - containerPort: 8080
        env:
        - name: FILESYSTEM_EXPORTER_CONFIG_FROM_ENV
          value: "true"
        - name: FILESYSTEM_EXPORTER_FILESYSTEMS
          value: "root:/host:sda1:1m"
        - name: FILESYSTEM_EXPORTER_DIRECTORIES
          value: "home:/host/home:1:10m,logs:/host/var/log:0:5m"
        volumeMounts:
        - name: host-root
          mountPath: /host
          readOnly: true
      volumes:
      - name: host-root
        hostPath:
          path: /
          type: Directory
```

### Duration Format
- `"60s"` - 60 seconds
- `"5m"` - 5 minutes  
- `"2h"` - 2 hours
- `"1h30m"` - 1 hour 30 minutes
- `"1h30m45s"` - 1 hour 30 minutes 45 seconds

### Basic Configuration

```yaml
server:
  host: "0.0.0.0"
  port: 8080

logging:
  level: "info"
  format: "json"

metrics:
  collection:
    default_interval: "5m"

filesystems:
  - name: "root"
    mount_point: "/"
    device: "sda1"
    interval: "1m"

directories:
  home:
    path: "/home"
    subdirectory_levels: 1
```

## Metrics

### Filesystem Metrics
- `filesystem_exporter_volume_size_bytes`: Total size of filesystem in bytes
- `filesystem_exporter_volume_available_bytes`: Available space on filesystem in bytes
- `filesystem_exporter_volume_used_ratio`: Ratio of used space (0.0 to 1.0)

### Directory Metrics
- `filesystem_exporter_directory_size_bytes`: Size of directory in bytes

### Collection Metrics
- `filesystem_exporter_collection_duration_seconds`: Duration of collection in seconds
- `filesystem_exporter_collection_timestamp`: Timestamp of last collection
- `filesystem_exporter_collection_success_total`: Total number of successful collections
- `filesystem_exporter_collection_failed_total`: Total number of failed collections

### Processing Metrics
- `filesystem_exporter_directories_processed_total`: Total number of directories processed
- `filesystem_exporter_directories_failed_total`: Total number of directories that failed to process

## Endpoints

- `GET /`: HTML dashboard with service status and metrics information
- `GET /metrics`: Prometheus metrics endpoint
- `GET /health`: Health check endpoint

## Use Cases

### Synology NAS Monitoring
```yaml
filesystems:
  - name: "volume1"
    mount_point: "/volume1"
    device: "sda1"
    interval: "1m"
  - name: "usb1"
    mount_point: "/volumeUSB1/usbshare"
    device: "usb1p1"
    interval: "2m"

directories:
  nas:
    path: "/volume1/nas/"
    subdirectory_levels: 1
  media:
    path: "/volume1/nas/Media/"
    subdirectory_levels: 2
```

### Docker Host Monitoring
```yaml
filesystems:
  - name: "host"
    mount_point: "/host"
    device: "host"
    interval: "1m"

directories:
  containers:
    path: "/host/var/lib/docker/containers"
    subdirectory_levels: 1
  volumes:
    path: "/host/var/lib/docker/volumes"
    subdirectory_levels: 1
```

### Kubernetes Monitoring
```yaml
filesystems:
  - name: "node"
    mount_point: "/host"
    device: "node"
    interval: "1m"

directories:
  pods:
    path: "/host/var/lib/kubelet/pods"
    subdirectory_levels: 2
  logs:
    path: "/host/var/log"
    subdirectory_levels: 1
```

## Development

### Prerequisites
- Go 1.21 or later
- Docker (optional)

### Local Development

1. Install dependencies:
   ```bash
   make deps
   ```

2. Run tests:
   ```bash
   make test
   ```

3. Build the application:
   ```bash
   make build
   ```

4. Run the application:
   ```bash
   make run
   ```

### Docker Development

1. Build the Docker image:
   ```bash
   make docker-build
   ```

2. Run the container:
   ```bash
   make docker-run
   ```

Or build and run in one command:
   ```bash
   make docker
   ```

## Project Structure
```
.
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── metrics/             # Prometheus metrics registry
│   ├── collectors/          # Metrics collectors
│   ├── server/              # HTTP server
│   └── logging/             # Logging configuration
├── config.yaml              # Configuration file
├── config.example.yaml      # Example configuration
├── go.mod                   # Go module file
├── Dockerfile               # Docker configuration
└── Makefile                 # Build and development tasks
```

## Testing

Run all tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

## Code Quality

Format code:
```bash
make fmt
```

Lint code:
```bash
make lint
```

## Deployment

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: filesystem-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: filesystem-exporter
  template:
    metadata:
      labels:
        app: filesystem-exporter
    spec:
      containers:
      - name: filesystem-exporter
        image: ghcr.io/d0ugal/filesystem-exporter:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: host-root
          mountPath: /host
          readOnly: true
        - name: config
          mountPath: /root/config.yaml
          subPath: config.yaml
      volumes:
      - name: host-root
        hostPath:
          path: /
      - name: config
        configMap:
          name: filesystem-exporter-config
```

### Prometheus Integration
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'filesystem-exporter'
    static_configs:
      - targets: ['filesystem-exporter:8080']
```

## Monitoring

The application provides several metrics for monitoring its own health:

- Collection success/failure rates
- Collection duration
- Processing success/failure rates
- Last collection timestamps

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure the container has read access to the mounted volumes
2. **Command Not Found**: The application requires `df` and `du` commands to be available
3. **Configuration Errors**: Check the YAML syntax in `config.yaml`
4. **High Memory Usage**: Large directories with many subdirectories can consume significant memory

### Logs
The application uses structured logging with JSON format. Log levels can be configured in the YAML configuration.

### Debug Mode
Enable debug logging to troubleshoot issues:
```yaml
logging:
  level: "debug"
  format: "json"
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License. 