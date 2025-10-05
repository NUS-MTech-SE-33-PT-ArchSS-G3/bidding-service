package domain

import (
	"time"
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
	StartingPrice float64       `json:"startingPrice"` // min starting price
	CurrentPrice  float64       `json:"currentPrice"`  // last accepted price, 0 if none
	MinIncrement  float64       `json:"minIncrement"`  // required when >= CurrentPrice
	Version       int           `json:"version"`
}

func (m AuctionMetadata) IsOpen() bool {
	return m.Status == AuctionOpen
}

// MinNextBid returns the min acceptable next bid
func (m AuctionMetadata) MinNextBid() float64 {
	if m.CurrentPrice <= 0 {
		// no bids yet, so starting price is min
		return m.StartingPrice
	}

	return m.CurrentPrice + m.MinIncrement
}
