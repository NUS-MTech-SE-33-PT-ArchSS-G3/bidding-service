package main

import (
	"context"
	"errors"
	"flag"
	"kei-services/pkg/config"
	kafkaInfra "kei-services/pkg/infra/kafka"
	"kei-services/pkg/infra/postgres"
	"kei-services/pkg/infra/redis"
	"kei-services/pkg/logger"
	"kei-services/services/bid-command/internal/cfg"
	"kei-services/services/bid-command/internal/server"
	"kei-services/services/bid-command/sqlc"
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

	//// Connect to infrastructures
	// postgres
	db, err := postgres.Client(cfg.Postgres, &cfg.Network.Ssl, log)
	if err != nil {
		log.Fatal("connect to database", zap.Error(err))
	}
	sqlDB, _ := db.DB()
	if err := sqlc.EnsureSchema(context.Background(), sqlDB); err != nil {
		log.Fatal("apply schema", zap.Error(err))
	}
	defer func() {
		if sqlDB, derr := db.DB(); derr == nil {
			_ = sqlDB.Close()
		}
	}()

	// Redis
	redisClient, err := redis.Client(cfg.Redis, log)
	if err != nil {
		log.Fatal("connect to redis", zap.Error(err))
	}
	defer func() { _ = redisClient.Close() }()

	// Kafka
	writer := kafkaInfra.NewWriter(cfg.KafkaWriter, log)
	defer writer.Close()

	//// Create and start server
	s := server.New(db, redisClient, writer, cfg, log)

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
	//stopBG() // stop subscriber
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx, s, log)
}
