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
	Reg     *prometheus.Registry
	Handler http.Handler
}

func New(o Options) *Registry {
	r := prometheus.NewRegistry()
	r.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return &Registry{
		Reg:     r,
		Handler: promhttp.HandlerFor(r, promhttp.HandlerOpts{EnableOpenMetrics: true}),
	}
}

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
