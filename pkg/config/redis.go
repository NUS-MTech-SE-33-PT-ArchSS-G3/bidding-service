package config

import "github.com/spf13/viper"

type Redis struct {
	Addr     string
	Password string `json:"-"`
	PoolSize int
	Port     string
}

func BindRedis(v *viper.Viper) {
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
}
