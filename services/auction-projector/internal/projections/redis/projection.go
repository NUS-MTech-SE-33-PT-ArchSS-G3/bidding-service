package redis

import (
	"context"
	"kei-services/services/auction-projector/internal/events"
	"time"

	"go.uber.org/zap"
)

type Projection struct {
	cache     *AuctionMetadataProjection
	log       *zap.Logger
	ttlBuffer time.Duration // extra time after EndsAt to keep key
}

func NewProjection(cache *AuctionMetadataProjection, log *zap.Logger, ttlBuffer time.Duration) *Projection {
	return &Projection{
		cache:     cache,
		log:       log,
		ttlBuffer: ttlBuffer,
	}
}

func (p *Projection) OnAuctionOpened(ctx context.Context, e events.AuctionOpened) error {
	meta := AuctionMetadata{
		AuctionID:     e.AuctionID,
		Status:        AuctionOpen,
		EndsAt:        e.EndsAt.UTC(),
		StartingPrice: e.StartingPrice,
		CurrentPrice:  0,
		MinIncrement:  e.MinIncrement,
		Version:       e.Version,
	}
	ttl := ttlFromEnd(e.EndsAt, p.ttlBuffer)
	return p.cache.SetIfNewer(ctx, e.AuctionID, meta, ttl)
}

func (p *Projection) OnAuctionClosed(ctx context.Context, e events.AuctionClosed) error {
	cur, err := p.cache.Get(ctx, e.AuctionID)
	if err != nil && err.Error() != "auction_metadata_not_found" {
		p.log.Warn("failed to get current auction metadata", zap.String("auctionID", e.AuctionID), zap.Error(err))
		// set closed state even if get fails
	}

	// populate fields from curr state and set status to closed
	meta := AuctionMetadata{
		AuctionID: e.AuctionID,
		Status:    AuctionClose,
		EndsAt: func() time.Time {
			if cur != nil {
				return cur.EndsAt
			}
			return time.Time{}
		}(),
		StartingPrice: func() float64 {
			if cur != nil {
				return cur.StartingPrice
			}
			return 0
		}(),
		CurrentPrice: func() float64 {
			if cur != nil {
				return cur.CurrentPrice
			}
			return 0
		}(),
		MinIncrement: func() float64 {
			if cur != nil {
				return cur.MinIncrement
			}
			return 0
		}(),
		Version: e.Version,
	}

	// keep closed auctions for 1 more hr
	return p.cache.SetIfNewer(ctx, e.AuctionID, meta, 1*time.Hour)
}
