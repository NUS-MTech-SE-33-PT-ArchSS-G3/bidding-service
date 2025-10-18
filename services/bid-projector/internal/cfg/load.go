package cfg

import (
	"fmt"
	"kei-services/pkg/config"
	"kei-services/pkg/infra/mongo"
	"kei-services/pkg/infra/redis"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func Load(path string) (*Config, error) {
	var cfg Config

	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// env bindings
	config.BindSsl(v)
	redis.BindEnv(v)
	mongo.BindMongoDb(v, "MONGO", "mongo")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config from %q: %w", path, err)
	}

	// for kafka brokers
	if s := os.Getenv("KAFKA_BROKERS"); s != "" {
		parts := splitCSV(s)
		if len(parts) > 0 {
			v.Set("kafka.brokers", parts)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func HotReload(path string, cfg *Config) {
	v := viper.New()
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
	})
}

func splitCSV(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' || r == ' ' })
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
