package main

import (
	"context"
	"errors"
	"flag"
	"kei-services/pkg/config"
	kafkaInfra "kei-services/pkg/infra/kafka"
	mongoInfra "kei-services/pkg/infra/mongo"
	"kei-services/pkg/infra/redis"
	"kei-services/pkg/logger"
	"kei-services/services/bid-query/internal/cfg"
	"kei-services/services/bid-query/internal/infrastructure/mq"
	"kei-services/services/bid-query/internal/server"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
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
		panic("initialize logger")
	}
	defer func() { _ = log.Sync() }()

	log.Info("Starting App", zap.String("version", cfg.App.Version))

	//// Connect to infrastructures
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

	// Redis // todo: use it
	redisClient, err := redis.Client(cfg.Redis, log)
	if err != nil {
		log.Fatal("redis new client", zap.Error(err))
	}
	defer func() { _ = redisClient.Close() }()

	// projector start
	// Kafka for projector to read model (todo: move to a projector service, should be in the same consumer group)
	ctx := context.Background()
	if err := mq.EnsureTopic(ctx, cfg.KafkaReader.Brokers, cfg.KafkaReader.Topic, 1, 1); err != nil {
		log.Warn("kafka ensure topic failed", zap.Error(err))
	}

	bidPlacedReader, err := kafkaInfra.NewReader(cfg.KafkaReader)
	if err != nil {
		log.Fatal("bid placed kafka reader", zap.Error(err))
	}
	defer bidPlacedReader.Close()
	log.Info("kafka reader configured",
		zap.Strings("brokers", cfg.KafkaReader.Brokers),
		zap.String("topic", cfg.KafkaReader.Topic),
		zap.Strings("groupTopics", cfg.KafkaReader.GroupTopics),
		zap.String("groupID", cfg.KafkaReader.GroupID),
		zap.String("startOffset", string(cfg.KafkaReader.Offset)),
	)

	proj := mq.NewBidPlacedProjector(bidPlacedReader, mc.DB, 100, log)

	bg, stopBG := context.WithCancel(context.Background())
	go func() {
		if err := proj.Run(bg); err != nil {
			log.Error("bids.placed subscriber stopped", zap.Error(err))
		}
	}()
	// projector end
	//direct := kafka.NewReader(kafka.ReaderConfig{
	//	Brokers:   []string{"kafka:9092"},
	//	Topic:     "bids.placed",
	//	Partition: 0,
	//	MinBytes:  1,
	//	MaxBytes:  10 << 20,
	//})
	//defer direct.Close()
	//_ = direct.SetOffset(kafka.FirstOffset) // or LastOffset
	//
	//m, err := direct.ReadMessage(context.Background())
	//log.Info("direct read", zap.Any("msg", m), zap.Error(err))

	//// Create and start server
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
	stopBG() // stop subscriber
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx, s, log)

}
