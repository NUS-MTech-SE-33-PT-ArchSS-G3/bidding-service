package domain

import "time"

// BidPlaced is a domain event emitted after a bid is accepted
type BidPlaced struct {
	AuctionID string    `json:"auctionId"`
	BidID     string    `json:"bidId"`
	BidderID  string    `json:"bidderId"`
	Amount    float64   `json:"amount"`
	At        time.Time `json:"at"`
}

// AuctionOpened is a domain event emitted by the auction service when an auction is opened
type AuctionOpened struct {
	AuctionID     string    `json:"auctionId"`
	EndsAt        time.Time `json:"endsAt"`
	StartingPrice float64   `json:"startingPrice"`
	MinIncrement  float64   `json:"minIncrement"`
	Currency      string    `json:"currency,omitempty"`
	Version       int       `json:"version"`
}

// AuctionClosed is a domain event emitted by the auction service when an auction is closed
type AuctionClosed struct {
	AuctionID string    `json:"auctionId"`
	ClosedAt  time.Time `json:"closedAt"`
	Version   int       `json:"version"`
}
