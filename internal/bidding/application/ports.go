package application

import (
	"context"
	"kei-services/internal/bidding/domain"
	"time"
)

type LatestBid struct {
	ID     string
	Amount float64
	Seq    int64 // monotonic sequence
	At     time.Time
}

// todo: move interfaces to domain layer
type BidRepository interface {
	Insert(ctx context.Context, b *domain.Bid) (id string, seq int64, err error)
	LatestForUpdate(ctx context.Context, auctionID string) (*LatestBid, error)
}

type AuctionMetadataStore interface {
	Get(ctx context.Context, auctionID string) (*domain.AuctionMetadata, error)
}

type BidsPlacedPublisher interface {
	Publish(ctx context.Context, evt domain.BidPlaced) error
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Clock interface {
	Now() time.Time
}
