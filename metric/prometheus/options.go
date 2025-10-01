package prometheus

import (
	"fmt"
	"os"
	"strconv"

	"github.com/minus5/svckit/env"
	promclient "github.com/prometheus/client_golang/prometheus"
)

const (
	// DefaultHTTPPort is the default port for Prometheus metrics HTTP server
	DefaultHTTPPort = 2112

	// DefaultMetricsPath is the default path for metrics endpoint
	DefaultMetricsPath = "/metrics"
)

var (
	// DefaultHistogramBuckets are time-based buckets in nanoseconds
	// Covers from 1ms (1,000,000 ns) to 10 seconds (10,000,000,000 ns)
	DefaultHistogramBuckets = []float64{
		1_000_000,      // 1ms
		2_500_000,      // 2.5ms
		5_000_000,      // 5ms
		10_000_000,     // 10ms
		25_000_000,     // 25ms
		50_000_000,     // 50ms
		100_000_000,    // 100ms
		250_000_000,    // 250ms
		500_000_000,    // 500ms
		1_000_000_000,  // 1s
		2_500_000_000,  // 2.5s
		5_000_000_000,  // 5s
		10_000_000_000, // 10s
	}

	staticPrometheusEnvVars = []string{
		"SVCKIT_METRIC_PROMETHEUS_PORT",
		"PROMETHEUS_PORT",
	}
)

// options is set of configurable options
type options struct {
	prefix    string
	port      int
	path      string
	namespace string
	subsystem string
	buckets   []float64
	registry  *promclient.Registry
}

// Validate options before start
func (o *options) Validate() error {
	// Set default port if not specified
	if o.port == 0 {
		o.port = getPortFromEnv()
		if o.port == 0 {
			o.port = DefaultHTTPPort
		}
	}

	// Set default path if not specified
	if o.path == "" {
		o.path = DefaultMetricsPath
	}

	// Set default buckets if not specified
	if o.buckets == nil {
		o.buckets = DefaultHistogramBuckets
	}

	// Create default registry if not provided
	if o.registry == nil {
		o.registry = promclient.NewRegistry()
	}

	// Set default prefix if empty
	if o.prefix == "" {
		o.prefix = getDefaultPrefix()
	}

	return nil
}

// getPortFromEnv tries to get Prometheus port from environment variables
func getPortFromEnv() int {
	for _, envVar := range staticPrometheusEnvVars {
		if portStr := os.Getenv(envVar); portStr != "" {
			if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
				return port
			}
		}
	}
	return 0
}

// Option is type for option implementation
type Option func(o *options)

// HTTPPort sets the HTTP server port for metrics endpoint
func HTTPPort(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

// MetricsPath sets the HTTP path for metrics endpoint
func MetricsPath(path string) Option {
	return func(o *options) {
		o.path = path
	}
}

// MetricPrefix is prefix to prepend to every metric being sent
func MetricPrefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

// Namespace sets the Prometheus namespace for all metrics
func Namespace(namespace string) Option {
	return func(o *options) {
		o.namespace = namespace
	}
}

// Subsystem sets the Prometheus subsystem for all metrics
func Subsystem(subsystem string) Option {
	return func(o *options) {
		o.subsystem = subsystem
	}
}

// HistogramBuckets sets custom histogram buckets for timing metrics
func HistogramBuckets(buckets []float64) Option {
	return func(o *options) {
		o.buckets = buckets
	}
}

// WithRegistry sets a custom Prometheus registry
func WithRegistry(registry *promclient.Registry) Option {
	return func(o *options) {
		o.registry = registry
	}
}

// getDefaultPrefix returns the default metric prefix based on app name and instance ID
func getDefaultPrefix() string {
	appName := env.AppName()
	instanceID := env.InstanceId()
	if appName != "" && instanceID != "" {
		return fmt.Sprintf("%s.%s.", appName, instanceID)
	}
	if appName != "" {
		return fmt.Sprintf("%s.", appName)
	}
	return ""
}
