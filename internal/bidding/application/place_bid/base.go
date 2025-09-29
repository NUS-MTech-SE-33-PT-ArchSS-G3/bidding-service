package place_bid

import (
	"context"
	"time"
)

type IService interface {
	Handle(ctx context.Context, cmd Command) (*Result, error)
}

type Command struct {
	AuctionID      string
	BidderID       string
	Amount         float64
	IdempotencyKey string
}

type Result struct {
	BidID        string
	AuctionID    string
	BidderID     string
	CurrentPrice float64
	MinNextBid   float64
	LeaderBidID  *string
	Version      int
	At           time.Time
}
