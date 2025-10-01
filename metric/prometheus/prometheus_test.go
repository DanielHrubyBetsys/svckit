package prometheus

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/minus5/svckit/metric"
)

func TestPrometheusCounter(t *testing.T) {
	// Initialize Prometheus with test port
	err := Dial(
		HTTPPort(19090),
		MetricPrefix("test"),
	)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer Close()

	// Send counter metric
	metric.Counter("test_counter")
	metric.Counter("test_counter", 5)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Scrape metrics
	resp, err := http.Get("http://localhost:19090/metrics")
	if err != nil {
		t.Fatalf("Failed to scrape metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	metrics := string(body)
	if !strings.Contains(metrics, "test_test_counter") {
		t.Errorf("Expected metric 'test_test_counter' not found in output:\n%s", metrics)
	}
}

func TestPrometheusGauge(t *testing.T) {
	// Initialize Prometheus with test port
	err := Dial(
		HTTPPort(19091),
		MetricPrefix("test"),
	)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer Close()

	// Send gauge metric
	metric.Gauge("test_gauge", 42)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Scrape metrics
	resp, err := http.Get("http://localhost:19091/metrics")
	if err != nil {
		t.Fatalf("Failed to scrape metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	metrics := string(body)
	if !strings.Contains(metrics, "test_test_gauge") {
		t.Errorf("Expected metric 'test_test_gauge' not found in output:\n%s", metrics)
	}
	if !strings.Contains(metrics, "42") {
		t.Errorf("Expected value '42' not found in output:\n%s", metrics)
	}
}

func TestPrometheusTiming(t *testing.T) {
	// Initialize Prometheus with test port
	err := Dial(
		HTTPPort(19092),
		MetricPrefix("test"),
	)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer Close()

	// Send timing metric
	metric.Timing("test_timing", func() {
		time.Sleep(10 * time.Millisecond)
	})

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Scrape metrics
	resp, err := http.Get("http://localhost:19092/metrics")
	if err != nil {
		t.Fatalf("Failed to scrape metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	metrics := string(body)
	if !strings.Contains(metrics, "test_test_timing") {
		t.Errorf("Expected metric 'test_test_timing' not found in output:\n%s", metrics)
	}
}

func TestPrometheusWithPrefix(t *testing.T) {
	// Initialize Prometheus with test port
	err := Dial(
		HTTPPort(19093),
		MetricPrefix("base"),
	)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer Close()

	// Create metric with custom prefix
	m := metric.WithPrefix("custom")
	m.Counter("test_counter")

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Scrape metrics
	resp, err := http.Get("http://localhost:19093/metrics")
	if err != nil {
		t.Fatalf("Failed to scrape metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	metrics := string(body)
	if !strings.Contains(metrics, "custom_test_counter") {
		t.Errorf("Expected metric 'custom_test_counter' not found in output:\n%s", metrics)
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with.dots", "with_dots"},
		{"with-dashes", "with_dashes"},
		{"with spaces", "with_spaces"},
		{"123start", "_123start"},
		{"valid_name", "valid_name"},
		{"http.requests.total", "http_requests_total"},
	}

	for _, tt := range tests {
		result := sanitizeName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestHealthEndpoint(t *testing.T) {
	// Initialize Prometheus with test port
	err := Dial(HTTPPort(19094))
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer Close()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Check health endpoint
	resp, err := http.Get("http://localhost:19094/health")
	if err != nil {
		t.Fatalf("Failed to check health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if string(body) != "OK" {
		t.Errorf("Expected 'OK', got %q", string(body))
	}
}
