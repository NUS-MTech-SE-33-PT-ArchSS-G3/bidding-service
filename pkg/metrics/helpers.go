package metrics

import "github.com/prometheus/client_golang/prometheus"

func CCounter(reg *prometheus.Registry, ns, name, help string, consts prometheus.Labels, labels []string) *prometheus.CounterVec {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: ns, Name: name, Help: help, ConstLabels: consts}, labels)
	reg.MustRegister(cv)
	return cv
}

func CHistogram(reg *prometheus.Registry, ns, name, help string, consts prometheus.Labels, labels []string, buckets []float64) *prometheus.HistogramVec {
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{Namespace: ns, Name: name, Help: help, ConstLabels: consts, Buckets: buckets}, labels)
	reg.MustRegister(hv)
	return hv
}

func CGauge(reg *prometheus.Registry, ns, name, help string, consts prometheus.Labels, labels []string) *prometheus.GaugeVec {
	gv := prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: ns, Name: name, Help: help, ConstLabels: consts}, labels)
	reg.MustRegister(gv)
	return gv
}
