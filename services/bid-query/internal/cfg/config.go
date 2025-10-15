package cfg

import (
	"kei-services/pkg/config"
	"kei-services/pkg/infra/kafka"
	"kei-services/pkg/infra/mongo"
	"kei-services/pkg/infra/redis"
	"kei-services/pkg/logger"
	"kei-services/pkg/profiler"
	swagger "kei-services/pkg/swagger"
)

type Config struct {
	App *config.App

	Network *config.Network

	Cors *config.Cors

	Pprof *profiler.Config

	Swagger *swagger.Config

	Logger *logger.Config

	Mongo *mongo.Config

	Redis *redis.Config

	KafkaReader *kafka.ReaderConfig
}
