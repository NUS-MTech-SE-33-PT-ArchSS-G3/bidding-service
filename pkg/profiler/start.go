package profiler

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	IsEnabled bool `json:"IsEnabled"`
	Port      int  `json:"Port"`
}

func Start(cfg *Config, log *zap.Logger) error {
	if !cfg.IsEnabled {
		log.Info("pprof is disabled")
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// local only
	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Port)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	defer func() {
		if r := recover(); r != nil {
			log.Error("panic", zap.Any("panic", r), zap.Stack("stack"))
		}
	}()

	log.Info("pprof listening", zap.String("addr", addr))
	return srv.ListenAndServe()
}
