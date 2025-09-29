package profiler

import (
	"errors"
	"fmt"
	"kei-services/pkg/config"
	"net/http"
	"net/http/pprof"
	"time"

	"go.uber.org/zap"
)

func Start(cfg *config.Pprof, log *zap.Logger) {
	if !cfg.IsEnabled {
		log.Info("pprof is disabled")
		return
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

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic", zap.Any("panic", r), zap.Stack("stack"))
			}
		}()

		log.Info("pprof listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Warn("pprof server error", zap.Error(err))
		}
	}()
}
