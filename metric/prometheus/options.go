package prometheus

import (
	"fmt"
	"os"
	"strconv"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/signal"
	promclient "github.com/prometheus/client_golang/prometheus"
)

const (
	// DefaultHTTPPort is the default port for Prometheus metrics HTTP server
	DefaultHTTPPort = 2112
	
	// DefaultMetricsPath is the default path for metrics endpoint
	DefaultMetricsPath = "/metrics"
)

var (
	// DefaultHistogramBuckets are time-based buckets suitable for timing metrics
	// Covers from 1ms to 10 seconds
	DefaultHistogramBuckets = []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}
	
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

// tryGetPortFromServiceDiscovery attempts to get port from service discovery
func tryGetPortFromServiceDiscovery() (int, error) {
	var addr dcy.Address
	err := signal.WithExponentialBackoff(func() error {
		var err error
		addr, err = dcy.Service("prometheus")
		return err
	})
	if err != nil {
		return 0, err
	}
	return addr.Port, nil
}
