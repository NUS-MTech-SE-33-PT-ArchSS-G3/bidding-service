package events

import (
	"encoding/json"
	"fmt"
	"time"
)

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

type Codec struct{}

func (c *Codec) Decode(topic string, payload []byte) (any, error) {
	switch topic {
	case "auction.opened":
		var e AuctionOpened
		if err := json.Unmarshal(payload, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "auction.closed":
		var e AuctionClosed
		if err := json.Unmarshal(payload, &e); err != nil {
			return nil, err
		}
		return e, nil
	default:
		return nil, fmt.Errorf("unknown topic %s", topic)
	}
}
