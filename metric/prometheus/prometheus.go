package prometheus

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/metric"
	"github.com/minus5/svckit/signal"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// invalidCharsRegex matches characters that are not valid in Prometheus metric names
	invalidCharsRegex = regexp.MustCompile(`[^a-zA-Z0-9_:]`)
	// startsWithInvalidRegex matches metric names that don't start with a letter or underscore
	startsWithInvalidRegex = regexp.MustCompile(`^[^a-zA-Z_]`)
)

// Prometheus metric driver.
// Implements metric.Metric interface.
type Prometheus struct {
	prefix     string
	namespace  string
	subsystem  string
	registry   *prometheus.Registry
	counters   map[string]prometheus.Counter
	gauges     map[string]prometheus.Gauge
	histograms map[string]prometheus.Histogram
	buckets    []float64
	mu         sync.RWMutex
	mapLock    sync.Mutex
	prefixes   map[string]*Prometheus
}

// newPrometheus creates a new Prometheus instance
func newPrometheus(prefix, namespace, subsystem string, registry *prometheus.Registry, buckets []float64) *Prometheus {
	return &Prometheus{
		prefix:     prefix,
		namespace:  namespace,
		subsystem:  subsystem,
		registry:   registry,
		counters:   make(map[string]prometheus.Counter),
		gauges:     make(map[string]prometheus.Gauge),
		histograms: make(map[string]prometheus.Histogram),
		buckets:    buckets,
		prefixes:   make(map[string]*Prometheus),
	}
}

// sanitizeName converts a metric name to Prometheus format
func sanitizeName(name string) string {
	// Replace dots with underscores
	name = strings.ReplaceAll(name, ".", "_")
	
	// Remove invalid characters
	name = invalidCharsRegex.ReplaceAllString(name, "_")
	
	// Ensure it starts with a letter or underscore
	if startsWithInvalidRegex.MatchString(name) {
		name = "_" + name
	}
	
	return name
}

// buildMetricName constructs the full metric name with prefix
func (p *Prometheus) buildMetricName(name string) string {
	sanitized := sanitizeName(name)
	if p.prefix != "" {
		return sanitizeName(p.prefix) + sanitized
	}
	return sanitized
}

// getOrCreateCounter gets or creates a counter metric
func (p *Prometheus) getOrCreateCounter(name string) prometheus.Counter {
	p.mu.RLock()
	counter, exists := p.counters[name]
	p.mu.RUnlock()
	
	if exists {
		return counter
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Double-check after acquiring write lock
	if counter, exists := p.counters[name]; exists {
		return counter
	}
	
	metricName := p.buildMetricName(name)
	counter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      metricName,
		Help:      fmt.Sprintf("Counter metric: %s", metricName),
	})
	
	if err := p.registry.Register(counter); err != nil {
		// If already registered (race condition), try to get it
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existingCounter, ok := are.ExistingCollector.(prometheus.Counter); ok {
				p.counters[name] = existingCounter
				return existingCounter
			}
		}
		logger().S("name", metricName).Error(err)
		return counter
	}
	
	p.counters[name] = counter
	return counter
}

// getOrCreateGauge gets or creates a gauge metric
func (p *Prometheus) getOrCreateGauge(name string) prometheus.Gauge {
	p.mu.RLock()
	gauge, exists := p.gauges[name]
	p.mu.RUnlock()
	
	if exists {
		return gauge
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Double-check after acquiring write lock
	if gauge, exists := p.gauges[name]; exists {
		return gauge
	}
	
	metricName := p.buildMetricName(name)
	gauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      metricName,
		Help:      fmt.Sprintf("Gauge metric: %s", metricName),
	})
	
	if err := p.registry.Register(gauge); err != nil {
		// If already registered (race condition), try to get it
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existingGauge, ok := are.ExistingCollector.(prometheus.Gauge); ok {
				p.gauges[name] = existingGauge
				return existingGauge
			}
		}
		logger().S("name", metricName).Error(err)
		return gauge
	}
	
	p.gauges[name] = gauge
	return gauge
}

// getOrCreateHistogram gets or creates a histogram metric
func (p *Prometheus) getOrCreateHistogram(name string) prometheus.Histogram {
	p.mu.RLock()
	histogram, exists := p.histograms[name]
	p.mu.RUnlock()
	
	if exists {
		return histogram
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Double-check after acquiring write lock
	if histogram, exists := p.histograms[name]; exists {
		return histogram
	}
	
	metricName := p.buildMetricName(name)
	histogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: p.namespace,
		Subsystem: p.subsystem,
		Name:      metricName,
		Help:      fmt.Sprintf("Histogram metric: %s", metricName),
		Buckets:   p.buckets,
	})
	
	if err := p.registry.Register(histogram); err != nil {
		// If already registered (race condition), try to get it
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existingHistogram, ok := are.ExistingCollector.(prometheus.Histogram); ok {
				p.histograms[name] = existingHistogram
				return existingHistogram
			}
		}
		logger().S("name", metricName).Error(err)
		return histogram
	}
	
	p.histograms[name] = histogram
	return histogram
}

// Counter increments counter name for sum(values).
// If called without values will increment for 1.
func (p *Prometheus) Counter(name string, values ...int) {
	value := 1
	if len(values) > 0 {
		value = 0
		for _, v := range values {
			value += v
		}
	}
	
	counter := p.getOrCreateCounter(name)
	counter.Add(float64(value))
}

// Gauge submits/updates a gauge type.
func (p *Prometheus) Gauge(name string, value int) {
	gauge := p.getOrCreateGauge(name)
	gauge.Set(float64(value))
}

// Timing measures execution time for f and submits it as histogram type.
func (p *Prometheus) Timing(name string, f func()) {
	stopwatch := metric.NewStopwatch()
	f()
	duration := stopwatch.GetNs()
	p.Time(name, duration)
}

// Time submits a histogram type.
// Duration is in nanoseconds, converted to seconds for Prometheus.
func (p *Prometheus) Time(name string, duration int) {
	histogram := p.getOrCreateHistogram(name)
	// Convert nanoseconds to seconds
	seconds := float64(duration) / 1e9
	histogram.Observe(seconds)
}

// WithPrefix returns a clone of the original metric, but with a different prefix
func (p *Prometheus) WithPrefix(prefix string) metric.Metric {
	p.mapLock.Lock()
	defer p.mapLock.Unlock()
	
	s, ok := p.prefixes[prefix]
	if ok && s != nil {
		return s
	}
	
	// New prefix for cloned instance
	mPrefix := prefix
	if !strings.HasSuffix(mPrefix, ".") {
		mPrefix += "."
	}
	
	// Create new instance sharing the same registry
	p.prefixes[prefix] = newPrometheus(mPrefix, p.namespace, p.subsystem, p.registry, p.buckets)
	return p.prefixes[prefix]
}

// AppendSuffix returns a clone of the original metric, but with the
// suffix appended to the end of the original prefix
func (p *Prometheus) AppendSuffix(suffix string) metric.Metric {
	return p.WithPrefix(p.prefix + suffix)
}

// Dial connects to Prometheus (starts HTTP server) and sets it as metric driver.
// - default port is read from env SVCKIT_METRIC_PROMETHEUS_PORT or PROMETHEUS_PORT, or uses 2112
// - default prefix is AppName.InstanceId
// Examples:
//   - Dial()
//   - Dial(prometheus.MetricPrefix("my_app"))
//   - Dial(prometheus.MetricPrefix("my_app"), prometheus.HTTPPort(9090))
func Dial(opts ...Option) error {
	// default options
	o := &options{
		prefix:  getDefaultPrefix(),
		port:    0, // will be set in Validate
		path:    DefaultMetricsPath,
		buckets: DefaultHistogramBuckets,
	}

	// apply sent options
	for _, optFn := range opts {
		optFn(o)
	}

	// validate options
	if err := o.Validate(); err != nil {
		if !env.InDev() {
			logger().Error(err)
		}
		return err
	}

	// create Prometheus instance without prefix (to support WithPrefix)
	prom := newPrometheus("", o.namespace, o.subsystem, o.registry, o.buckets)

	// set Prometheus as metric driver with default prefix
	metric.Set(prom.WithPrefix(o.prefix))

	// start HTTP server
	if err := startServer(o.port, o.path, o.registry); err != nil {
		logger().Error(err)
		return err
	}

	logger().I("port", o.port).S("path", o.path).S("prefix", o.prefix).Info("started")
	return nil
}

// MustDial same as Dial but raises Fatal on error on failure
func MustDial(opts ...Option) {
	r := 0
again:
	if err := Dial(opts...); err != nil {
		if r > 10 {
			log.Fatal(err)
		}
		r++
		time.Sleep(time.Second)
		goto again
	}
}

// TryDial attempts to dial with exponential backoff
func TryDial(opts ...Option) {
	go func() {
		if err := Dial(opts...); err == nil {
			return
		}
		ctx := context.Background()
		signal.WithBackoff(ctx, func() error {
			return Dial(opts...)
		}, time.Minute, time.Hour*24*365)
	}()
}

func logger() *log.Agregator {
	return log.S("lib", "svckit.metric.prometheus")
}
