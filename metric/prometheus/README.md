# Prometheus Metric Driver

A Prometheus driver implementation for the svckit metric package. This driver exposes metrics via an HTTP endpoint that can be scraped by Prometheus.

## Features

- **Pull-based metrics**: Exposes metrics via HTTP endpoint for Prometheus scraping
- **Standard metric types**: Counter, Gauge, and Histogram (for timing)
- **Automatic metric naming**: Converts metric names to Prometheus format
- **Configurable options**: Port, path, prefix, namespace, subsystem, and histogram buckets
- **Health check endpoint**: Built-in `/health` endpoint
- **Thread-safe**: Safe for concurrent use
- **Prefix support**: Create metrics with custom prefixes using `WithPrefix()` and `AppendSuffix()`

## Installation

The driver is part of the svckit package. The Prometheus client library is automatically included as a dependency.

## Usage

### Basic Usage

```go
package main

import (
    "github.com/minus5/svckit/metric"
    "github.com/minus5/svckit/metric/prometheus"
)

func main() {
    // Initialize Prometheus driver with defaults
    // - Port: 2112
    // - Path: /metrics
    // - Prefix: AppName.InstanceId
    prometheus.MustDial()
    
    // Use metrics as usual
    metric.Counter("requests_total")
    metric.Counter("requests_total", 5)
    metric.Gauge("active_connections", 42)
    metric.Timing("request_duration", func() {
        // do work
    })
    
    // Metrics available at http://localhost:2112/metrics
    // Health check at http://localhost:2112/health
}
```

### Custom Configuration

```go
prometheus.MustDial(
    prometheus.MetricPrefix("myapp"),
    prometheus.HTTPPort(9090),
    prometheus.MetricsPath("/metrics"),
    prometheus.Namespace("svckit"),
    prometheus.Subsystem("api"),
)
```

### Custom Histogram Buckets

```go
// For timing metrics, customize histogram buckets
prometheus.MustDial(
    prometheus.HistogramBuckets([]float64{0.001, 0.01, 0.1, 1, 10}),
)
```

### Error Handling

```go
// Use Dial() for error handling
if err := prometheus.Dial(); err != nil {
    log.Printf("Failed to initialize Prometheus: %v", err)
}

// Use TryDial() for non-blocking initialization with retry
prometheus.TryDial(
    prometheus.HTTPPort(9090),
)
```

### Graceful Shutdown

```go
defer prometheus.Close()
```

## Configuration Options

### HTTPPort(port int)
Sets the HTTP server port for the metrics endpoint.
- Default: 2112
- Can be overridden via env vars: `SVCKIT_METRIC_PROMETHEUS_PORT` or `PROMETHEUS_PORT`

### MetricsPath(path string)
Sets the HTTP path for the metrics endpoint.
- Default: `/metrics`

### MetricPrefix(prefix string)
Sets the prefix to prepend to every metric.
- Default: `AppName.InstanceId.` (from environment)

### Namespace(namespace string)
Sets the Prometheus namespace for all metrics.
- Default: empty

### Subsystem(subsystem string)
Sets the Prometheus subsystem for all metrics.
- Default: empty

### HistogramBuckets(buckets []float64)
Sets custom histogram buckets for timing metrics.
- Default: `[0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10]` (seconds)

### WithRegistry(registry *prometheus.Registry)
Sets a custom Prometheus registry.
- Default: creates a new registry

## Metric Types

### Counter
Increments a counter metric. Counters only go up.

```go
metric.Counter("requests_total")        // increment by 1
metric.Counter("requests_total", 5)     // increment by 5
metric.Counter("requests_total", 2, 3)  // increment by 5 (sum of values)
```

### Gauge
Sets a gauge metric to a specific value. Gauges can go up and down.

```go
metric.Gauge("active_connections", 42)
metric.Gauge("temperature", -5)
```

### Timing / Histogram
Records timing information as a histogram.

```go
// Timing measures execution time
metric.Timing("request_duration", func() {
    // do work
})

// Time records a duration in nanoseconds
duration := 1500000 // 1.5ms in nanoseconds
metric.Time("operation_duration", duration)
```

## Metric Naming

Metric names are automatically sanitized to comply with Prometheus naming rules:
- Dots (`.`) are replaced with underscores (`_`)
- Invalid characters are replaced with underscores
- Names starting with numbers get an underscore prefix

Examples:
- `http.requests.total` → `http_requests_total`
- `my-metric` → `my_metric`
- `123metric` → `_123metric`

## Prefix Management

Create metrics with custom prefixes:

```go
// Create a metric instance with custom prefix
m := metric.WithPrefix("api")
m.Counter("requests")  // Creates metric: api_requests

// Append to existing prefix
m2 := m.AppendSuffix("v1")
m2.Counter("requests")  // Creates metric: api_v1_requests
```

## Architecture

The Prometheus driver follows the same pattern as the StatsD driver:

1. **Pull-based**: Prometheus scrapes metrics from the application's HTTP endpoint
2. **In-memory storage**: Metrics are stored in memory using Prometheus client library
3. **Lazy initialization**: Metrics are created on first use
4. **Thread-safe**: All operations are protected with mutexes
5. **Shared registry**: All prefix variants share the same registry

## Differences from StatsD

| Feature | StatsD | Prometheus |
|---------|--------|------------|
| Model | Push (UDP) | Pull (HTTP) |
| Protocol | Custom | HTTP + Text format |
| Aggregation | Server-side | Client-side |
| Timing | Timing type | Histogram |
| Labels | Not supported | Not supported (in this implementation) |

## Example Prometheus Configuration

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'myapp'
    static_configs:
      - targets: ['localhost:2112']
```

## Future Enhancements

Potential improvements for future versions:
- Label/tag support using CounterVec, GaugeVec, HistogramVec
- Summary metrics in addition to histograms
- Custom metric help text
- Metric metadata and descriptions
- Integration with Prometheus Pushgateway for batch jobs

## See Also

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [svckit metric package](../metric.go)
- [StatsD driver](../statsd/)
