package list_bids

import (
	"context"
	"time"
)

type IService interface {
	Handle(ctx context.Context, q Query) (*Result, error)
}

// Direction (server-side sort by sequence)
type Direction int

const (
	DirectionDesc Direction = iota
	DirectionAsc
)

type Query struct {
	AuctionID string
	Cursor    string // empty = first page
	Limit     int
	Direction Direction
}

type Item struct {
	BidID     string
	AuctionID string
	BidderID  string
	Amount    float64
	At        time.Time
}

type Result struct {
	Items      []Item
	NextCursor *string // nil when no more
	HasMore    bool
}
