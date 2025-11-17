package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type ServerMetrics struct {
	Registry *prometheus.Registry

	Inflight   prometheus.Gauge
	ReqTotal   *prometheus.CounterVec
	ReqLatency *prometheus.HistogramVec
	ReqSize    *prometheus.HistogramVec
	RespSize   *prometheus.HistogramVec
	Panics     *prometheus.CounterVec

	handler http.Handler
}

type HTTPOpts struct {
	Namespace string
	//ConstLabels prometheus.Labels
	Buckets []float64
}

func NewHTTPServerMetrics(r *Registry, o HTTPOpts) *ServerMetrics {
	buckets := o.Buckets
	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	inflight := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: o.Namespace,
		Name:      "http_inflight_requests",
		Help:      "Current number of inflight HTTP requests.",
		//ConstLabels: o.ConstLabels,
	})
	reqTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: o.Namespace,
		Name:      "http_requests_total",
		Help:      "Total HTTP requests by method, route, and status.",
		//ConstLabels: o.ConstLabels,
	}, []string{"method", "route", "status"})
	reqLatency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: o.Namespace,
		Name:      "http_request_duration_seconds",
		Help:      "Request duration seconds by method and route.",
		//ConstLabels: o.ConstLabels,
		Buckets: buckets,
	}, []string{"method", "route"})
	reqSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: o.Namespace, Name: "http_request_size_bytes",
		Help: "Approximate request size.",
		//ConstLabels: o.ConstLabels,
		Buckets: []float64{200, 500, 1_000, 5_000, 10_000, 50_000, 100_000, 500_000, 1_000_000},
	}, []string{"method", "route"})
	respSize := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: o.Namespace, Name: "http_response_size_bytes",
		Help: "Response body size.",
		//ConstLabels: o.ConstLabels,
		Buckets: []float64{200, 500, 1_000, 5_000, 10_000, 50_000, 100_000, 500_000, 1_000_000},
	}, []string{"method", "route"})
	panics := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: o.Namespace, Name: "http_panics_total",
		Help: "Total recovered panics by route.",
		//ConstLabels: o.ConstLabels,
	}, []string{"route"})

	r.Reg.MustRegister(inflight, reqTotal, reqLatency, reqSize, respSize, panics)

	return &ServerMetrics{
		Registry:   r.Reg,
		Inflight:   inflight,
		ReqTotal:   reqTotal,
		ReqLatency: reqLatency,
		ReqSize:    reqSize,
		RespSize:   respSize,
		Panics:     panics,
		handler:    r.Handler, // reuse registry's handler for /metrics
	}
}

func (m *ServerMetrics) Handler() http.Handler { return m.handler }

func (m *ServerMetrics) Observe(method, route string, status int, dur time.Duration) {
	m.ReqLatency.WithLabelValues(method, route).Observe(dur.Seconds())
	m.ReqTotal.WithLabelValues(method, route, strconv.Itoa(status)).Inc()
}

// PanicCounterMiddleware is a Gin middleware that counts recovered panics
// Must be placed BEFORE gin recovery
func PanicCounterMiddleware(m *ServerMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				route := c.FullPath()
				if route == "" {
					route = "UNKNOWN"
				}
				m.Panics.WithLabelValues(route).Inc()
				panic(r) // reraise panic
			}
		}()
		c.Next()
	}
}

func GinAdapterMiddleware(m *ServerMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m.Inflight.Inc()
		defer m.Inflight.Dec()
		contentLen := c.Request.ContentLength

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "UNKNOWN"
		}
		if contentLen > 0 {
			m.ReqSize.WithLabelValues(c.Request.Method, route).Observe(float64(contentLen))
		}
		respBytes := c.Writer.Size()
		if respBytes > 0 {
			m.RespSize.WithLabelValues(c.Request.Method, route).Observe(float64(respBytes))
		}
		m.Observe(c.Request.Method, route, c.Writer.Status(), time.Since(start))
	}
}
