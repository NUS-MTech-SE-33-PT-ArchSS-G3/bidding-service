package cfg

import (
	"kei-services/pkg/config"
	"kei-services/pkg/infra/kafka"
	"kei-services/pkg/infra/mysql"
)

type Config struct {
	App *config.App

	Network *config.Network

	Cors *config.Cors

	Pprof *config.Pprof

	Swagger *config.Swagger

	Logger *config.Logger

	SqlDb *mysql.SqlDbConfig

	Redis *config.Redis

	KafkaWriter *kafka.WriterConfig
}
