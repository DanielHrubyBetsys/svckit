package prometheus

// Usage examples for Prometheus metric driver
//
// Basic usage:
//
//	import "github.com/minus5/svckit/metric/prometheus"
//
//	func main() {
//		// Initialize with defaults (port 2112, path /metrics)
//		prometheus.MustDial()
//
//		// Use metrics as usual
//		metric.Counter("requests_total")
//		metric.Gauge("active_connections", 42)
//		metric.Timing("request_duration", func() {
//			// do work
//		})
//
//		// Metrics available at http://localhost:2112/metrics
//	}
//
// Custom configuration:
//
//	prometheus.MustDial(
//		prometheus.MetricPrefix("myapp"),
//		prometheus.HTTPPort(9090),
//		prometheus.MetricsPath("/custom/metrics"),
//		prometheus.Namespace("svckit"),
//		prometheus.Subsystem("api"),
//	)
//
// With custom histogram buckets:
//
//	prometheus.MustDial(
//		prometheus.HistogramBuckets([]float64{0.001, 0.01, 0.1, 1, 10}),
//	)
//
// Using TryDial for non-blocking initialization with retry:
//
//	prometheus.TryDial(
//		prometheus.HTTPPort(9090),
//	)
//
// Using Dial for error handling:
//
//	if err := prometheus.Dial(); err != nil {
//		log.Printf("Failed to initialize Prometheus: %v", err)
//	}
//
// Graceful shutdown:
//
//	defer prometheus.Close()
