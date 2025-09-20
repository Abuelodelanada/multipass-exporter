# Multipass Exporter

A Prometheus exporter for [Multipass](https://multipass.run/) that exposes metrics about your Multipass instances.

## Overview

This exporter provides metrics about Multipass virtual machines, making it easy to monitor your local development environment with Prometheus and visualize the data with Grafana.

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `multipass_instances_total` | Gauge | Total number of Multipass instances |
| `multipass_instances_running` | Gauge | Number of currently running Multipass instances |
| `multipass_instances_stopped` | Gauge | Number of currently stopped Multipass instances |
| `multipass_error` | Gauge | Error indicator (1 when collection fails, 0 otherwise) |

## Installation

### From Source

```bash
git clone https://github.com/Abuelodelanada/multipass-exporter.git
cd multipass-exporter
go build -o multipass-exporter ./cmd/multipass-exporter
```

### Requirements

- Go 1.23 or later
- Multipass installed and accessible in your PATH
- Linux system

## Configuration

The exporter uses a YAML configuration file. By default, it looks for `config.yaml` in the current directory.

### Configuration File

Create a `config.yaml` file:

```yaml
# Port to listen on (default: 8080)
port: 9090

# Metrics endpoint path (default: /metrics)
metrics_path: /metrics

# Timeout for multipass commands in seconds (default: 5)
timeout_seconds: 5
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `port` | 8080 | TCP port for the HTTP server |
| `metrics_path` | /metrics | HTTP path for metrics endpoint |
| `timeout_seconds` | 5 | Timeout for multipass command execution |

## Usage

### Running the Exporter

```bash
# Show help
./multipass-exporter --help

# Run with default configuration (port 8080, /metrics endpoint)
./multipass-exporter

# Run with config file
./multipass-exporter --config config.yaml

# Run with custom config file
./multipass-exporter --config /path/to/custom-config.yaml
```

### Accessing Metrics

Once running, access the metrics at:

```
http://localhost:9090/metrics
```

Example output:

```
# HELP multipass_instances_total Total number of Multipass instances
# TYPE multipass_instances_total gauge
multipass_instances_total 3

# HELP multipass_instances_running Total number of Multipass running instances
# TYPE multipass_instances_running gauge
multipass_instances_running 2

# HELP multipass_instances_stopped Total number of Multipass stopped instances
# TYPE multipass_instances_stopped gauge
multipass_instances_stopped 1

# HELP multipass_error Error collecting metrics from Multipass
# TYPE multipass_error gauge
multipass_error 0
```

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'multipass'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 30s
```


## Development

### Building

```bash
# Build for current Linux platform
go build -o multipass-exporter ./cmd/multipass-exporter

# Build for different Linux architectures
GOARCH=amd64 go build -o multipass-exporter-amd64 ./cmd/multipass-exporter
GOARCH=arm64 go build -o multipass-exporter-arm64 ./cmd/multipass-exporter
```

### Testing

```bash
go test ./...
```

### Running Locally

```bash
# Run with default configuration
go run ./cmd/multipass-exporter

# Run with config file
go run ./cmd/multipass-exporter --config config.yaml
```

## Security Considerations

- The exporter runs commands as the same user who executes it
- Ensure the config file has appropriate permissions
- Consider running behind a reverse proxy with authentication in production environments

## Troubleshooting

### Common Issues

1. **"multipass command not found"**: Ensure Multipass is installed and in your PATH
2. **"Permission denied"**: Check file permissions and user access to multipass
3. **"Connection refused"**: Verify the port is not already in use

### Debug Mode

For troubleshooting, you can increase logging verbosity by checking the application output.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Support

For issues and questions:
- [GitHub Issues](https://github.com/Abuelodelanada/multipass-exporter/issues)
- [Multipass Documentation](https://multipass.run/docs)

---

**Note**: This exporter requires Multipass to be installed and accessible. It only monitors local Multipass instances and does not support remote Multipass installations.