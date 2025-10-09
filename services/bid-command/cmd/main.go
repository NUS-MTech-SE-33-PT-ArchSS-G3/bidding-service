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
	"kei-services/services/bid-command/internal/infrastructure/cache"
	"kei-services/services/bid-command/internal/infrastructure/mq"
	"kei-services/services/bid-command/internal/server"
	"kei-services/services/bid-command/sqlc"
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

var cfgFlag = flag.String("config", "", "path to config file (optional)")

func main() {
	flag.Parse() // parse -config flag
	_ = godotenv.Load(".env")
	cfg, err := cfg.Load(config.Path())
	if err != nil {
		panic(err)
	}

	log := logger.Init(cfg.Logger, cfg.App)
	if log == nil {
		panic("failed to initialize logger")
	}
	defer func() { _ = log.Sync() }()

	log.Info("Starting App", zap.String("version", cfg.App.Version))

	//// Connect to infrastructures
	// postgres
	db, err := postgres.Client(cfg.Postgres, &cfg.Network.Ssl, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
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
		log.Fatal("Failed to connect to redis", zap.Error(err))
	}
	defer func() { _ = redisClient.Close() }()

	// Kafka
	brokers := []string{"localhost:9092"}
	if host := os.Getenv("KAFKA_BROKERS"); host != "" {
		// allow comma-separated
		brokers = splitCSV(host)
	}
	//openedReader := kafka.NewReader(kafka.ReaderConfig{
	//	Brokers:  brokers,
	//	Topic:    "auction.opened",
	//	GroupID:  "bidding-meta-sync",
	//	MinBytes: 1, MaxBytes: 10 << 20,
	//})
	//defer openedReader.Close()
	//
	//closedReader := kafka.NewReader(kafka.ReaderConfig{
	//	Brokers: brokers,
	//	Topic:   "auction.closed",
	//	GroupID: "bid-command-meta-sync",
	//})
	//defer closedReader.Close()
	openedReader := kafkaInfra.NewReader(kafkaInfra.ReaderConfig{
		Brokers: brokers,
		Topic:   "auction.opened",
		GroupID: "bid-command-meta-sync",
		Offset:  kafkaInfra.OffsetLast, // or OffsetFirst for fresh groups
	})
	defer openedReader.Close()

	closedReader := kafkaInfra.NewReader(kafkaInfra.ReaderConfig{
		Brokers: brokers,
		Topic:   "auction.closed",
		GroupID: "bid-command-meta-sync",
		Offset:  kafkaInfra.OffsetLast,
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
	//writer := &kafka.Writer{
	//	Addr:     kafka.TCP(brokers...),
	//	Topic:    "bids.placed",
	//	Balancer: &kafka.Hash{},
	//}
	writer := kafkaInfra.NewWriter(kafkaInfra.WriterConfig{
		Brokers:  brokers,
		Topic:    "bids.placed",
		ClientID: "bid-command-service",
		Acks:     kafka.RequireAll,
		// Compression: kafka.Snappy,
		// Balancer:    &kafka.Hash{},
	})
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
