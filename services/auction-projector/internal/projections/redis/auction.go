package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goRedis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type AuctionStatus string

const (
	AuctionOpen  AuctionStatus = "OPEN"
	AuctionClose AuctionStatus = "CLOSED"
)

type AuctionMetadata struct {
	AuctionID     string        `json:"auctionID"`
	Status        AuctionStatus `json:"status"`
	EndsAt        time.Time     `json:"endsAt"`
	StartingPrice float64       `json:"startingPrice"`
	CurrentPrice  float64       `json:"currentPrice"`
	MinIncrement  float64       `json:"minIncrement"`
	Version       int           `json:"version"`
}

type AuctionMetadataProjection struct {
	keyPrefix string
	redis     *goRedis.Client
	log       *zap.Logger
}

func NewAuctionMetadataProjection(r *goRedis.Client, log *zap.Logger) *AuctionMetadataProjection {
	return &AuctionMetadataProjection{
		keyPrefix: "auction:",
		redis:     r,
		log:       log,
	}
}

// key returns the Redis key for a given auction ID
func (p *AuctionMetadataProjection) key(id string) string { return p.keyPrefix + id }

// Get returns auction metadata from cache
func (p *AuctionMetadataProjection) Get(ctx context.Context, auctionID string) (*AuctionMetadata, error) {
	raw, err := p.redis.Get(ctx, p.key(auctionID)).Bytes()
	if errors.Is(err, goRedis.Nil) {
		p.log.Debug("auction metadata not found in cache", zap.String("auctionID", auctionID))
		return nil, errors.New("auction_metadata_not_found")
	}
	if err != nil {
		p.log.Warn("failed to get auction metadata from cache", zap.String("auctionID", auctionID), zap.Error(err))
		return nil, err
	}

	var auction AuctionMetadata
	if err = json.Unmarshal(raw, &auction); err != nil {
		p.log.Error("failed to unmarshal auction metadata", zap.String("auctionID", auctionID), zap.Error(err))
		return nil, err
	}
	return &auction, nil
}

//func (p *AuctionMetadataProjection) Set(ctx context.Context, auctionID string, auction AuctionMetadata, ttl time.Duration) error {
//	raw, err := json.Marshal(auction)
//	if err != nil {
//		p.log.Error("failed to marshal auction metadata", zap.String("auctionID", auctionID), zap.Error(err))
//		return err
//	}
//
//	p.log.Debug("setting auction metadata in cache", zap.String("auctionID", auctionID), zap.Int("version", auction.Version))
//	return p.redis.Set(ctx, p.key(auctionID), raw, ttl).Err()
//}

// SetIfNewer ensures monotonic version updates
func (p *AuctionMetadataProjection) SetIfNewer(ctx context.Context, auctionID string, auction AuctionMetadata, ttl time.Duration) error {
	raw, err := json.Marshal(auction)
	if err != nil {
		p.log.Error("failed to marshal auction metadata", zap.String("auctionID", auctionID), zap.Error(err))
		return err
	}

	p.log.Debug("setting auction metadata in cache if newer", zap.String("auctionID", auctionID), zap.Int("version", auction.Version))
	ttlSec := int(ttl / time.Second)
	_, err = setIfNewerLua.Run(ctx, p.redis, []string{p.key(auctionID)}, auction.Version, string(raw), ttlSec).Result()
	return err
}

// setIfNewerLua sets the key only if the new version is greater than the existing version
var setIfNewerLua = goRedis.NewScript(`
local key = KEYS[1]
local version = tonumber(ARGV[1])
local value = ARGV[2]
local ttlsec = tonumber(ARGV[3])

local cur = redis.call('GET', key)
if cur then
  local ok, obj = pcall(cjson.decode, cur)
  if ok and obj['version'] and tonumber(obj['version']) > version then
    return 0
  end
end
redis.call('SET', key, value)
if ttlsec > 0 then
  redis.call('EXPIRE', key, ttlsec)
end
return 1
`)
