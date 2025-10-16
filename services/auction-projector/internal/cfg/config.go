package cfg

import (
	"kei-services/pkg/config"
	"kei-services/pkg/infra/kafka"
	"kei-services/pkg/infra/redis"
	"kei-services/pkg/logger"
)

type Config struct {
	App *config.App

	Network *config.Network

	Logger *logger.Config

	Redis *redis.Config

	KafkaReader *kafka.ReaderConfig
}
