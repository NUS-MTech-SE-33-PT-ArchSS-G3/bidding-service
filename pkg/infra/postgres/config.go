package postgres

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

type Config struct {
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
	SSLMode    string // disable|require|verify-ca|verify-full

	LogLevel logger.LogLevel // 1 Silent  2 Error 3 Warn 4 Info
}

// BindPostgresDb loads DB config from viper with env fallbacks
// eg: envPrefix="SQLDB", viperPrefix="sqlDb"
func BindPostgresDb(v *viper.Viper, envPrefix, viperPrefix string) *Config {
	// Bind specific envs for secrets/host
	for _, key := range []string{"user", "password", "host", "port", "dbName", "params"} {
		envKey := fmt.Sprintf("%s_%s", envPrefix, strings.ToUpper(key))
		viperKey := fmt.Sprintf("%s.%s", viperPrefix, key)
		_ = v.BindEnv(viperKey, envKey)
	}
	var db Config
	_ = v.UnmarshalKey(viperPrefix, &db)
	return &db
}
