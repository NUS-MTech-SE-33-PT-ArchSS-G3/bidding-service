package main

import (
	"context"
	"flag"
	"kei-services/pkg/config"
	kafkaInfra "kei-services/pkg/infra/kafka"
	redisInfra "kei-services/pkg/infra/redis"
	"kei-services/pkg/logger"
	"kei-services/services/auction-projector/internal/cfg"
	"kei-services/services/auction-projector/internal/events"
	redisProjection "kei-services/services/auction-projector/internal/projections/redis"
	"kei-services/services/auction-projector/internal/projector"
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

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	// redis
	redisClient, err := redisInfra.Client(cfg.Redis, log)
	if err != nil {
		log.Fatal("connect to redis", zap.Error(err))
	}
	defer func() { _ = redisClient.Close() }()

	// setup kafka reader
	// ensure topics
	if err = projector.EnsureTopics(ctx, cfg.KafkaReader.Brokers, cfg.KafkaReader.GroupTopics, 1, 1); err != nil {
		log.Warn("ensure topics", zap.Strings("topics", cfg.KafkaReader.GroupTopics), zap.Error(err))
	}

	auctionReader, err := kafkaInfra.NewReader(cfg.KafkaReader)
	if err != nil {
		log.Fatal("kafka reader", zap.Error(err))
	}
	defer auctionReader.Close()

	log.Info("kafka reader configured",
		zap.Strings("brokers", cfg.KafkaReader.Brokers),
		zap.String("topic", cfg.KafkaReader.Topic),
		zap.Strings("groupTopics", cfg.KafkaReader.GroupTopics),
		zap.String("groupID", cfg.KafkaReader.GroupID),
		zap.String("startOffset", string(cfg.KafkaReader.Offset)),
	)

	// wire projector
	cache := redisProjection.NewAuctionMetadataProjection(redisClient, log)
	redisProjection := redisProjection.NewProjection(cache, log, 15*time.Minute)

	router := &projector.Router{
		Codec:    &events.Codec{},
		Handlers: redisProjection,
	}

	p := projector.New(auctionReader, router, log)

	// run projector
	go func() {
		if err = p.Run(ctx); err != nil {
			log.Error("projector stopped", zap.Error(err))
		}
	}()

	// block until signal
	<-ctx.Done()
	log.Info("shutdown signal received")
}
