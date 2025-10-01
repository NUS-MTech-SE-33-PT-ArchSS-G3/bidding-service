package main

import (
	"context"
	"errors"
	"kei-services/internal/bidding/cfg"
	"kei-services/internal/bidding/infrastructure/cache"
	"kei-services/internal/bidding/infrastructure/mq"
	"kei-services/internal/bidding/server"
	"kei-services/pkg/config"
	"kei-services/pkg/infra/mysql"
	"kei-services/pkg/infra/redis"
	"kei-services/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"

	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load(".env")
	cfg := cfg.Load(config.Path())

	log := logger.Init(cfg.Logger, cfg.App)
	if log == nil {
		panic("failed to initialize logger")
	}
	defer func() { _ = log.Sync() }()

	log.Info("Starting App", zap.String("version", cfg.App.Version))

	//// Connect to infrastructures
	// Database
	db, err := mysql.Client(cfg.SqlDb, &cfg.Network.Ssl, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if sqlDB, derr := db.DB(); derr == nil {
			_ = sqlDB.Close()
		}
	}()

	// Redis
	redisClient, err := redis.Client(cfg.Redis, log)
	if err != nil {
		log.Fatal("Failed to connect to redis", zap.Error(err))
	}
	defer func() { _ = redisClient.Close() }()

	// Kafka
	// readers for meta sync (host dev uses localhost:9092; if containerize, use kafka:9092)
	brokers := []string{"localhost:9092"}
	if host := os.Getenv("KAFKA_BROKERS"); host != "" {
		// allow comma-separated
		brokers = splitCSV(host)
	}
	openedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "auction.opened",
		GroupID:  "bidding-meta-sync",
		MinBytes: 1, MaxBytes: 10 << 20,
	})
	defer openedReader.Close()

	closedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "auction.closed",
		GroupID: "bidding-meta-sync",
	})
	defer closedReader.Close()

	sub := &mq.AuctionMetaSubscriber{
		Log:          log,
		OpenedReader: openedReader,
		ClosedReader: closedReader,
		Store:        cache.AuctionMetadataCache{R: redisClient},
		DefaultTTL:   24 * time.Hour,
	}

	bg, stopBG := context.WithCancel(context.Background())
	go func() {
		if err := sub.Run(bg); err != nil {
			log.Error("auction meta subscriber stopped", zap.Error(err))
		}
	}()

	// Kafka writer for publishing BidPlaced (if you have a helper, use that)
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    "bids.placed",
		Balancer: &kafka.Hash{},
	}
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx, s, log)
	case err = <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http: server error", zap.Error(err))
		}
	}

	// graceful shutdown
	stopBG() // stop subscriber
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx, s, log)

}

func splitCSV(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' || r == ' ' })
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
