package main

import (
	"context"
	"errors"
	"flag"
	"kei-services/pkg/config"
	mongoInfra "kei-services/pkg/infra/mongo"
	"kei-services/pkg/infra/redis"
	"kei-services/pkg/logger"
	"kei-services/services/bid-query/internal/cfg"
	"kei-services/services/bid-query/internal/server"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	flag.Parse() // parse -config flag
	_ = godotenv.Load(".env")
	cfg, err := cfg.Load(config.Path())
	if err != nil {
		panic(err)
	}

	log := logger.Init(cfg.Logger, cfg.App)
	if log == nil {
		panic("initialize logger")
	}
	defer func() { _ = log.Sync() }()

	log.Info("Starting App", zap.String("version", cfg.App.Version))

	// mongo
	mc, err := mongoInfra.NewClient(cfg.Mongo, log)
	if err != nil {
		log.Fatal("mongo client", zap.Error(err))
	}
	defer func() {
		if err := mc.Disconnect(context.Background()); err != nil {
			log.Error("disconnect mongo client", zap.Error(err))
		}
	}()

	// Redis
	redisClient, err := redis.Client(cfg.Redis, log)
	if err != nil {
		log.Fatal("redis new client", zap.Error(err))
	}
	defer func() { _ = redisClient.Close() }()

	// Create and start server
	s := server.New(mc.DB, redisClient, cfg, log)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(s, cfg, log)
	}()

	// catch signals
	sigCtx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	select {
	case <-sigCtx.Done():
		log.Info("shutdown signal received")
	case err = <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http: server error", zap.Error(err))
		}
	}

	// graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx, s, log)
}
