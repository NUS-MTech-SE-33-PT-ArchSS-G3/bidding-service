package main

import (
	"context"
	"flag"
	"kei-services/pkg/config"
	kafkaInfra "kei-services/pkg/infra/kafka"
	mongoInfra "kei-services/pkg/infra/mongo"
	"kei-services/pkg/logger"
	"kei-services/services/bid-projector/internal/cfg"
	"kei-services/services/bid-projector/internal/events"
	mongoProjection "kei-services/services/bid-projector/internal/projections/mongo"
	"kei-services/services/bid-projector/internal/projector"
	"os"
	"os/signal"
	"syscall"

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

	// setup kafka reader
	// ensure topics
	if err = projector.EnsureTopics(ctx, cfg.KafkaReader.Brokers, cfg.KafkaReader.GroupTopics, 1, 1); err != nil {
		log.Warn("ensure topics", zap.Strings("topics", cfg.KafkaReader.GroupTopics), zap.Error(err))
	}

	bidReader, err := kafkaInfra.NewReader(cfg.KafkaReader)
	if err != nil {
		log.Fatal("kafka reader", zap.Error(err))
	}
	defer bidReader.Close()

	log.Info("kafka reader configured",
		zap.Strings("brokers", cfg.KafkaReader.Brokers),
		zap.String("topic", cfg.KafkaReader.Topic),
		zap.Strings("groupTopics", cfg.KafkaReader.GroupTopics),
		zap.String("groupID", cfg.KafkaReader.GroupID),
		zap.String("startOffset", string(cfg.KafkaReader.Offset)),
	)

	// wire projector
	mongoDbProjection := mongoProjection.NewProjection(mc.DB, log)

	router := &projector.Router{
		Codec:    &events.Codec{},
		Handlers: mongoDbProjection,
	}

	p := projector.New(bidReader, router, log)

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
