package mongo

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Config struct {
	URI                string // e.g. "mongodb://mongo:27017"
	DBName             string // e.g. "bid_query"
	User               string
	Password           string `json:"-"`
	TLS                bool
	ConnectTimeoutSec  int
	ServerSelectionSec int
	MaxPoolSize        uint64
	MinPoolSize        uint64
}

func BindMongoDb(v *viper.Viper, envPrefix, viperPrefix string) *Config {
	for _, key := range []string{"uri", "dbName", "user", "password", "tls", "connectTimeoutSec", "serverSelectionSec", "maxPoolSize", "minPoolSize"} {
		envKey := fmt.Sprintf("%s_%s", envPrefix, strings.ToUpper(key))
		viperKey := fmt.Sprintf("%s.%s", viperPrefix, key)
		_ = v.BindEnv(viperKey, envKey)
	}
	var db Config
	_ = v.UnmarshalKey(viperPrefix, &db)
	return &db
}

type Client struct {
	*mongo.Client
	DB *mongo.Database
}

func NewClient(cfg *Config, log *zap.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("mongo config is nil")
	}

	opts := options.Client().ApplyURI(cfg.URI)
	if cfg.User != "" {
		opts.SetAuth(options.Credential{
			Username: cfg.User,
			Password: cfg.Password,
		})
	}
	if cfg.TLS {
		opts.SetTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12})
	}
	if cfg.ConnectTimeoutSec > 0 {
		opts.SetConnectTimeout(time.Duration(cfg.ConnectTimeoutSec) * time.Second)
	}
	if cfg.ServerSelectionSec > 0 {
		opts.SetServerSelectionTimeout(time.Duration(cfg.ServerSelectionSec) * time.Second)
	}
	if cfg.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize > 0 {
		opts.SetMinPoolSize(cfg.MinPoolSize)
	}

	log.Info("Connecting to MongoDB...", zap.String("uri", cfg.URI), zap.String("db", cfg.DBName))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := cl.Ping(ctx, nil); err != nil {
		_ = cl.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	mc := &Client{Client: cl, DB: cl.Database(cfg.DBName)}

	log.Info("Connected to MongoDB")
	return mc, nil
}
