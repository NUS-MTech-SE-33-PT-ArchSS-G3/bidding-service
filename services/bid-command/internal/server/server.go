package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"kei-services/pkg/middleware"
	swagger "kei-services/pkg/swagger"
	"kei-services/services/bid-command/internal/cfg"
	"kei-services/services/bid-command/openapi"

	"kei-services/pkg/profiler"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	ginzap "github.com/gin-contrib/zap"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	srv    *http.Server
	engine *gin.Engine
	cfg    *cfg.Config
	log    *zap.Logger
}

func New(db *gorm.DB, redis *redis.Client, w *kafka.Writer, cfg *cfg.Config, log *zap.Logger) *Server {
	if cfg.App.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	r.Use(
		ginzap.Ginzap(log, time.RFC3339, true),
		ginzap.RecoveryWithZap(log, true),

		middleware.RequestID(),
		middleware.WithRequestLogger(log),
		middleware.RequestLogger(log),
		middleware.Cors(cfg.Cors, log),
		middleware.MaxBody(10<<20), // 10 mb
	)

	go profiler.Start(cfg.Pprof, log)
	swagger.Serve(func() (*openapi3.T, error) {
		return openapi.GetSwagger()
	}, r, cfg.Swagger, log)

	registerHealthroutes(r, db, redis, log)
	registerProtectedRoutes(r, initDependencies(db, redis, w, cfg, log), cfg, log)

	r.NoRoute(func(c *gin.Context) { c.JSON(404, gin.H{"error": "not found"}) })
	r.NoMethod(func(c *gin.Context) { c.JSON(405, gin.H{"error": "method not allowed"}) })

	addr := fmt.Sprintf(":%d", cfg.Network.Port)
	if cfg.Network.IsLocalHost {
		addr = fmt.Sprintf("127.0.0.1:%d", cfg.Network.Port)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1mb
	}

	return &Server{srv: srv, engine: r, cfg: cfg, log: log}
}

func Start(s *Server, cfg *cfg.Config, log *zap.Logger) error {
	if !cfg.Network.Ssl.IsEnabled {
		log.Info("Starting server without TLS", zap.String("address", s.srv.Addr))
		return s.srv.ListenAndServe()
	} else {
		s.srv.TLSConfig = tls13OnlyConfig()

		log.Info("Starting server with TLS 1.3 only",
			zap.String("address", s.srv.Addr),
			zap.String("cert", cfg.Network.Ssl.CertFile),
			zap.String("key", cfg.Network.Ssl.KeyFile),
		)

		return s.srv.ListenAndServeTLS(cfg.Network.Ssl.CertFile, cfg.Network.Ssl.KeyFile)
	}
}

func Shutdown(ctx context.Context, s *Server, log *zap.Logger) error {
	log.Info("Shutting down server")
	return s.srv.Shutdown(ctx)
}

func tls13OnlyConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		NextProtos: []string{"h2", "http/1.1"},
	}
}
