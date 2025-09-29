package redis

import (
	"context"
	"fmt"
	"kei-services/pkg/config"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func Client(cfg *config.Redis, log *zap.Logger) (*redis.Client, error) {
	log.Info("Connecting to Redis...")
	log.Debug("Connection parameters",
		zap.String("Host", cfg.Addr),
		zap.String("Port", cfg.Port),
		zap.Int("PoolSize", cfg.PoolSize))

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr + ":" + cfg.Port,
		Password: cfg.Password,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ping
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis cluster ping failed: %w", err)
	}

	log.Info("Connected to Redis cluster")
	return client, nil
}
