package mysql

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

type SqlDbConfig struct {
	User     string
	Password string `json:"-"`
	Host     string `json:"-"`
	Port     int
	DBName   string
	Params   string
	// Pooling
	MaxIdleConns           int
	MaxOpenConns           int
	ConnMaxLifetimeMinutes int

	SSLEnabled bool
	LogLevel   logger.LogLevel // 1 Silent  2 Error 3 Warn 4 Info
}

func BindSqlDb(v *viper.Viper, envPrefix, viperPrefix string) *SqlDbConfig {
	keys := []string{"password", "host"}

	for _, key := range keys {
		envKey := fmt.Sprintf("%s_%s", envPrefix, strings.ToUpper(key))
		viperKey := fmt.Sprintf("%s.%s", viperPrefix, key)
		_ = v.BindEnv(viperKey, envKey)
	}

	var db SqlDbConfig
	_ = v.UnmarshalKey(viperPrefix, &db)
	return &db
}
