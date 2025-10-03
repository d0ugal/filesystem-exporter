# Quick Start Guide

Get filesystem-exporter running in minutes!

## Option 1: Docker (Recommended)

### 1. Pull the image
```bash
docker pull ghcr.io/d0ugal/filesystem-exporter:v1.14.4
```

### 2. Create a configuration file
```bash
# Copy the example configuration to create your own config
cp config.example.yaml config.yaml

# Edit the configuration for your environment
nano config.yaml
```

### 3. Run the container
```bash
docker run -d \
  --name filesystem-exporter \
  -p 8080:8080 \
  -v /:/host:ro \
  -v $(pwd)/config.yaml:/root/config.yaml:ro \
  ghcr.io/d0ugal/filesystem-exporter:v1.14.4
```

### 4. Verify it's working
```bash
# Check the health endpoint
curl http://localhost:8080/health

# View metrics
curl http://localhost:8080/metrics
```

## Option 2: From Source

### 1. Clone the repository
```bash
git clone https://github.com/d0ugal/filesystem-exporter.git
cd filesystem-exporter
```

### 2. Build and run
```bash
# Build the application
make build

# Run with default configuration
./filesystem-exporter
```

## Option 3: Docker Compose

### 1. Create docker-compose.yml
```yaml
version: '3.8'
services:
  filesystem-exporter:
    image: ghcr.io/d0ugal/filesystem-exporter:v1.14.4
    ports:
      - "8080:8080"
    volumes:
      - /:/host:ro
      - ./config.yaml:/root/config.yaml:ro
    restart: unless-stopped
```

### 2. Run with Docker Compose
```bash
docker-compose up -d
```

## Configuration Examples

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

### Synology NAS Configuration
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

## Next Steps

1. **Configure for your environment** - Edit `config.yaml` with your filesystems and directories
2. **Set up monitoring** - Add the metrics endpoint to your Prometheus configuration
3. **Create dashboards** - Build Grafana dashboards using the available metrics
4. **Set up alerts** - Configure alerts for disk space and directory size thresholds

## Troubleshooting

### Common Issues

**Permission Denied**
- Ensure the container has read access to mounted volumes
- Use `:ro` (read-only) flag for volume mounts

**Configuration Errors**
- Check YAML syntax in your config file
- Validate paths exist and are accessible

**High Memory Usage**
- Reduce `subdirectory_levels` for large directory trees
- Increase collection intervals for less frequent monitoring

### Getting Help

- **Documentation**: See [README.md](README.md) for detailed documentation
- **Issues**: Report bugs and request features on [GitHub](https://github.com/d0ugal/filesystem-exporter/issues)
- **Examples**: Check [config.example.yaml](config.example.yaml) for more configuration examples
