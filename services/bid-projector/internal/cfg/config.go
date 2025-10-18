package cfg

import (
	"kei-services/pkg/config"
	"kei-services/pkg/infra/kafka"
	"kei-services/pkg/infra/mongo"
	"kei-services/pkg/logger"
)

type Config struct {
	App *config.App

	Network *config.Network

	Logger *logger.Config

	Mongo *mongo.Config

	KafkaReader *kafka.ReaderConfig
}
