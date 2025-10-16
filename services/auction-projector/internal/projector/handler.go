package projector

import (
	"context"
	"kei-services/services/auction-projector/internal/events"
)

type AuctionHandlers interface {
	OnAuctionOpened(ctx context.Context, e events.AuctionOpened) error
	OnAuctionClosed(ctx context.Context, e events.AuctionClosed) error
}
