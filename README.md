<h1 align="center">
  <img src="images/logo.png?raw=true" alt="Multipass Exporter" width="20%">
  <br />
  Multipass Exporter
</h1>


A Prometheus exporter for [Multipass](https://multipass.run/) that exposes metrics about your Multipass instances.

## Overview

This exporter provides metrics about Multipass virtual machines, making it easy to monitor your local development environment with Prometheus.

## Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `multipass_instances_total` | Gauge | Total number of Multipass instances |
| `multipass_instances_running` | Gauge | Number of currently running Multipass instances |
| `multipass_instances_stopped` | Gauge | Number of currently stopped Multipass instances |
| `multipass_instances_deleted` | Gauge | Number of deleted Multipass instances |
| `multipass_instances_suspended` | Gauge | Number of suspended Multipass instances |
| `multipass_instance_memory_bytes` | Gauge | Memory usage of Multipass instances in bytes (with `name` and `release` labels) |
| `multipass_instance_cpu_total` | Gauge | Total number of CPUs in Multipass instances (with `name` and `release` labels) |
| `multipass_instance_load_1m` | Gauge | Average number of processes running on CPU or in queue waiting for CPU time in the last minute (with `name` and `release` labels) |
| `multipass_instance_load_5m` | Gauge | Average number of processes running on CPU or in queue waiting for CPU time in the last 5 minutes (with `name` and `release` labels) |
| `multipass_instance_load_15m` | Gauge | Average number of processes running on CPU or in queue waiting for CPU time in the last 15 minutes (with `name` and `release` labels) |
| `multipass_error` | Gauge | Error indicator (1 when collection fails, 0 otherwise) |

## Installation

### From Source

```bash
git clone https://github.com/Abuelodelanada/multipass-exporter.git
cd multipass-exporter
make build
```

### Requirements

- Go 1.23 or later
- Multipass installed and accessible in your PATH

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

# Log level (default: info). Available levels: debug, info, warn, error, fatal
# Logs are formatted as: LEVEL timestamp message fields
log_level: debug
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `port` | 1986 | TCP port for the HTTP server |
| `metrics_path` | /metrics | HTTP path for metrics endpoint |
| `timeout_seconds` | 5 | Timeout for multipass command execution |
| `log_level` | info | Log level (debug, info, warn, error, fatal) |

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
http://localhost:1986/metrics
```

Example output:

```
# HELP multipass_instances_total Total number of Multipass instances
# TYPE multipass_instances_total gauge
multipass_instances_total 8

# HELP multipass_instances_running Total number of Multipass running instances
# TYPE multipass_instances_running gauge
multipass_instances_running 6

# HELP multipass_instances_stopped Total number of Multipass stopped instances
# TYPE multipass_instances_stopped gauge
multipass_instances_stopped 1

# HELP multipass_instances_deleted Total number of Multipass deleted instances
# TYPE multipass_instances_deleted gauge
multipass_instances_deleted 1

# HELP multipass_instances_suspended Total number of Multipass suspended instances
# TYPE multipass_instances_suspended gauge
multipass_instances_suspended 0

# HELP multipass_instance_memory_bytes Memory usage of Multipass instances in bytes
# TYPE multipass_instance_memory_bytes gauge
multipass_instance_memory_bytes{name="charm-dev-36",release="Ubuntu 24.04.3 LTS"} 3.388157952e+09
multipass_instance_memory_bytes{name="coslite",release="Ubuntu 22.04.2 LTS"} 2.986962944e+09
multipass_instance_memory_bytes{name="edp",release="Ubuntu 23.10"} 1.530298368e+09

# HELP multipass_instance_cpu_total Total number of CPUs in Multipass instances
# TYPE multipass_instance_cpu_total gauge
multipass_instance_cpu_total{name="charm-dev-36",release="Ubuntu 24.04.3 LTS"} 2
multipass_instance_cpu_total{name="coslite",release="Ubuntu 22.04.2 LTS"} 1
multipass_instance_cpu_total{name="edp",release="Ubuntu 23.10"} 4

# HELP multipass_instance_load_1m Average number of processes running on CPU or in queue waiting for CPU time in the last minute
# TYPE multipass_instance_load_1m gauge
multipass_instance_load_1m{name="charm-dev-36",release="Ubuntu 24.04.3 LTS"} 0.11
multipass_instance_load_1m{name="coslite",release="Ubuntu 22.04.2 LTS"} 0.23
multipass_instance_load_1m{name="edp",release="Ubuntu 23.10"} 0.15

# HELP multipass_instance_load_5m Average number of processes running on CPU or in queue waiting for CPU time in the last 5 minutes
# TYPE multipass_instance_load_5m gauge
multipass_instance_load_5m{name="charm-dev-36",release="Ubuntu 24.04.3 LTS"} 0.23
multipass_instance_load_5m{name="coslite",release="Ubuntu 22.04.2 LTS"} 0.3
multipass_instance_load_5m{name="edp",release="Ubuntu 23.10"} 0.28

# HELP multipass_instance_load_15m Average number of processes running on CPU or in queue waiting for CPU time in the last 15 minutes
# TYPE multipass_instance_load_15m gauge
multipass_instance_load_15m{name="charm-dev-36",release="Ubuntu 24.04.3 LTS"} 0.3
multipass_instance_load_15m{name="coslite",release="Ubuntu 22.04.2 LTS"} 0.4
multipass_instance_load_15m{name="edp",release="Ubuntu 23.10"} 0.35

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
      - targets: ['localhost:1986']
    scrape_interval: 60s
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

### Linting

```bash
# Run full linter (auto-installs golangci-lint if needed)
make lint

# Run fast linting mode
make lint-fast
```

### Testing

```bash
make test
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

## Troubleshooting

### Common Issues

1. **"multipass command not found"**: Ensure Multipass is installed and in your PATH
2. **"Permission denied"**: Check file permissions and user access to multipass
3. **"Connection refused"**: Verify the port is not already in use

### Debug Mode

For troubleshooting, you can configure debug logging through the configuration file:

```yaml
log_level: debug
```

Log format: `LEVEL timestamp message fields`

Available log levels: `debug`, `info`, `warn`, `error`, `fatal`

Example debug output:
```
INFO   [2025-09-22T23:35:03-03:00] Starting metrics collection
DEBUG  [2025-09-22T23:35:03-03:00] Executing multipass info command
INFO   [2025-09-22T23:35:03-03:00] Successfully parsed multipass info            instance_count=8
DEBUG  [2025-09-22T23:35:03-03:00] Collecting instance total                     count=8
DEBUG  [2025-09-22T23:35:03-03:00] Collecting instance running                   count=6
DEBUG  [2025-09-22T23:35:03-03:00] Collecting instance stopped                   count=1
DEBUG  [2025-09-22T23:35:03-03:00] Collecting instance deleted                   count=1
DEBUG  [2025-09-22T23:35:03-03:00] Collecting instance suspended                 count=0
INFO   [2025-09-22T23:35:03-03:00] Collecting memory metrics                     instance_count=8
DEBUG  [2025-09-22T23:35:03-03:00] Collecting CPU metrics                        instance_count=8
```

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
