package domain

import (
	"context"
	"time"
)

type LatestBid struct {
	ID     string
	Amount float64
	Seq    int64 // monotonic sequence
	At     time.Time
}

type IBidRepository interface {
	Insert(ctx context.Context, b *Bid) (id string, seq int64, err error)
	LatestForUpdate(ctx context.Context, auctionID string) (*LatestBid, error)
}

type IAuctionMetadataStore interface {
	Get(ctx context.Context, auctionID string) (*AuctionMetadata, error)
}

type IBidsPlacedPublisher interface {
	Publish(ctx context.Context, evt BidPlaced) error
}

type IClock interface {
	Now() time.Time
}
