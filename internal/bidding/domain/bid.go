package domain

import "time"

// Bid is a domain entity created when a bid is placed
type Bid struct {
	ID        string // todo: id currently assigned by repo, empty when bid placed maybe pass from app layer?
	AuctionID string
	BidderID  string
	Amount    float64
	At        time.Time
}

func NewBid(auctionID string, bidderID string, amount float64, at time.Time) *Bid {
	return &Bid{
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    amount,
		At:        at.UTC(),
	}
}

// WithID returns a copy with ID set
func (b *Bid) WithID(id string) *Bid {
	nb := *b
	nb.ID = id
	return &nb
}
