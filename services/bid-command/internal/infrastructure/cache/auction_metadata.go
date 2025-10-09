package cache

import (
	"context"
	"encoding/json"
	"errors"
	"kei-services/services/bid-command/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var _ domain.IAuctionMetadataStore = (*AuctionMetadataCache)(nil)

type AuctionMetadataCache struct {
	CacheKey string
	R        *redis.Client
	Log      *zap.Logger
}

func NewAuctionMetadataCache(r *redis.Client, log *zap.Logger) *AuctionMetadataCache {
	return &AuctionMetadataCache{
		CacheKey: "auction:",
		R:        r,
		Log:      log,
	}
}

// Get returns auction metadata from cache
func (c AuctionMetadataCache) Get(ctx context.Context, auctionId string) (*domain.AuctionMetadata, error) {
	raw, err := c.R.Get(ctx, c.CacheKey+auctionId).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("auction_metadata_not_found")
	}
	if err != nil {
		return nil, err
	}

	var meta domain.AuctionMetadata
	if err = json.Unmarshal(raw, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// Set stores auction metadata in cache with a TTL.
// TTL should be set according to auction duration + buffer
func (c AuctionMetadataCache) Set(ctx context.Context, id string, auction domain.AuctionMetadata, ttl time.Duration) error {
	b, err := json.Marshal(auction)
	if err != nil {
		return err
	}

	return c.R.Set(ctx, c.CacheKey+id, b, ttl).Err()
}

func (c AuctionMetadataCache) Delete(ctx context.Context, id string) error {
	return c.R.Del(ctx, c.CacheKey+id).Err()
}
