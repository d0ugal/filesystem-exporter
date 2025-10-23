# Filesystem Exporter

A lightweight Go-based metrics collection service for Prometheus that monitors filesystem and directory sizes.

**Image**: `ghcr.io/d0ugal/filesystem-exporter:v1.21.0`

## Metrics

### Filesystem Metrics
- `filesystem_exporter_volume_size_bytes`: Total size of filesystem in bytes
- `filesystem_exporter_volume_available_bytes`: Available space on filesystem in bytes
- `filesystem_exporter_volume_used_ratio`: Ratio of used space (0.0 to 1.0)

### Directory Metrics
- `filesystem_exporter_directory_size_bytes`: Size of directory in bytes

### Collection Metrics
- `filesystem_exporter_collection_duration_seconds`: Duration of collection in seconds
- `filesystem_exporter_collection_success_total`: Total number of successful collections
- `filesystem_exporter_collection_failed_total`: Total number of failed collections

### Endpoints
- `GET /`: HTML dashboard with service status and metrics information
- `GET /metrics`: Prometheus metrics endpoint
- `GET /health`: Health check endpoint

## Quick Start

### Docker Compose

```yaml
version: '3.8'
services:
  filesystem-exporter:
    image: ghcr.io/d0ugal/filesystem-exporter:v1.21.0
    ports:
      - "8080:8080"
    volumes:
      - /:/host:ro
      - ./config.yaml:/root/config.yaml:ro
    restart: unless-stopped
```

1. Create a `config.yaml` file (see Configuration section)
2. Run: `docker-compose up -d`
3. Access metrics: `curl http://localhost:8080/metrics`

## Configuration

Create a `config.yaml` file to configure filesystems and directories to monitor:

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

## Deployment

### Docker Compose (Environment Variables)

```yaml
version: '3.8'
services:
  filesystem-exporter:
    image: ghcr.io/d0ugal/filesystem-exporter:v1.21.0
    ports:
      - "8080:8080"
    volumes:
      - /:/host:ro
    environment:
      - FILESYSTEM_EXPORTER_CONFIG_FROM_ENV=true
      - FILESYSTEM_EXPORTER_FILESYSTEMS=root:/host:sda1:1m
      - FILESYSTEM_EXPORTER_DIRECTORIES=home:/host/home:1:10m,logs:/host/var/log:0:5m
    restart: unless-stopped
```

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
        image: ghcr.io/d0ugal/filesystem-exporter:v1.21.0
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

## Prometheus Integration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'filesystem-exporter'
    static_configs:
      - targets: ['filesystem-exporter:8080']
```

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