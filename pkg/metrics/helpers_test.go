package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNew(t *testing.T) {
	opts := Options{
		Namespace: "test",
		ConstLabels: prometheus.Labels{
			"app": "test-app",
		},
	}

	registry := New(opts)

	if registry == nil {
		t.Fatal("expected non-nil registry")
	}

	if registry.Reg == nil {
		t.Error("expected non-nil Reg")
	}

	if registry.Handler == nil {
		t.Error("expected non-nil Handler")
	}
}

func TestMountMetrics(t *testing.T) {
	opts := Options{
		Namespace: "test",
	}

	registry := New(opts)
	handler := MountMetrics(registry.Handler)

	if handler == nil {
		t.Error("expected non-nil handler")
	}
}

func TestCCounter(t *testing.T) {
	reg := prometheus.NewRegistry()

	counter := CCounter(
		reg,
		"test",
		"test_counter",
		"Test counter help",
		prometheus.Labels{"const": "value"},
		[]string{"label1", "label2"},
	)

	if counter == nil {
		t.Fatal("expected non-nil counter")
	}

	// Test incrementing the counter
	counter.WithLabelValues("val1", "val2").Inc()

	// Verify metric is registered
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("expected at least one metric family")
	}
}

func TestCHistogram(t *testing.T) {
	reg := prometheus.NewRegistry()

	histogram := CHistogram(
		reg,
		"test",
		"test_histogram",
		"Test histogram help",
		prometheus.Labels{"const": "value"},
		[]string{"label1"},
		[]float64{0.1, 0.5, 1.0, 5.0},
	)

	if histogram == nil {
		t.Fatal("expected non-nil histogram")
	}

	// Test observing values
	histogram.WithLabelValues("val1").Observe(0.3)
	histogram.WithLabelValues("val1").Observe(1.5)

	// Verify metric is registered
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("expected at least one metric family")
	}
}

func TestCGauge(t *testing.T) {
	reg := prometheus.NewRegistry()

	gauge := CGauge(
		reg,
		"test",
		"test_gauge",
		"Test gauge help",
		prometheus.Labels{"const": "value"},
		[]string{"label1"},
	)

	if gauge == nil {
		t.Fatal("expected non-nil gauge")
	}

	// Test setting gauge value
	gauge.WithLabelValues("val1").Set(42.5)
	gauge.WithLabelValues("val1").Inc()
	gauge.WithLabelValues("val1").Dec()

	// Verify metric is registered
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("expected at least one metric family")
	}
}
