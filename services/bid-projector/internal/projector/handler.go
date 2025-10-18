package projector

import (
	"context"
	"kei-services/services/bid-projector/internal/events"
)

type AuctionHandlers interface {
	OnBidsPlaced(ctx context.Context, e events.BidPlaced) error
}
