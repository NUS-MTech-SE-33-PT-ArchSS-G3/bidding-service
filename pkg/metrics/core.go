package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Options struct {
	Namespace   string
	ConstLabels prometheus.Labels
}
type Registry struct {
	Reg         *prometheus.Registry  // the raw registry (no labels)
	Registerer  prometheus.Registerer // the registerer WITH ConstLabels applied
	Handler     http.Handler          // http handler for /metrics
	ConstLabels prometheus.Labels
	DefaultNS   string
}

func New(o Options) *Registry {
	// 1) Base registry
	base := prometheus.NewRegistry()

	// 2) Wrap the registerer with your const labels (applies to everything registered through it)
	regWith := prometheus.WrapRegistererWith(o.ConstLabels, base)

	// 3) Register default collectors INTO THE WRAPPED REGISTERER
	regWith.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
	)

	// 4) Expose handler from the base registry (it sees everything)
	h := promhttp.HandlerFor(base, promhttp.HandlerOpts{EnableOpenMetrics: true})

	return &Registry{
		Reg:         base,
		Registerer:  regWith,
		Handler:     h,
		ConstLabels: o.ConstLabels,
		DefaultNS:   o.Namespace,
	}
}

//type Registry struct {
//	Reg     *prometheus.Registry
//	Handler http.Handler
//}
//
//func New(o Options) *Registry {
//	r := prometheus.NewRegistry()
//	r.MustRegister(
//		collectors.NewGoCollector(),
//		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
//
//	return &Registry{
//		Reg:     r,
//		Handler: promhttp.HandlerFor(r, promhttp.HandlerOpts{EnableOpenMetrics: true}),
//	}
//}

// Serve starts a HTTP server to serve metrics
//
//	Usage: go func() {
//	 _ = Serve(ctx, ":8080", metrics.MountMetrics(metricsReg.Handler)) }()
func Serve(ctx context.Context, addr string, h http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	return srv.ListenAndServe()
}

func MountMetrics(h http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", h)
	return mux
}
