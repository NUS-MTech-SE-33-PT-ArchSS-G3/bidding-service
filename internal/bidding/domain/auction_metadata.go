package domain

import (
	"time"
)

type AuctionStatus int

const (
	AuctionOpen  AuctionStatus = iota + 1
	AuctionClose AuctionStatus = iota + 2
)

type AuctionMetadata struct {
	AuctionID     string
	Status        AuctionStatus
	EndsAt        time.Time
	StartingPrice float64 // min starting price
	CurrentPrice  float64 // last accepted price, 0 if none
	MinIncrement  float64 // required when >= CurrentPrice
	Version       int
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
