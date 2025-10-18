package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// BidPlaced is a domain event emitted by the bid command service when a bid is accepted
type BidPlaced struct {
	AuctionID string    `json:"auctionId"`
	BidID     string    `json:"bidId"`
	BidderID  string    `json:"bidderId"`
	Amount    float64   `json:"amount"`
	At        time.Time `json:"at"`
}

type Codec struct{}

func (c *Codec) Decode(topic string, payload []byte) (any, error) {
	switch topic {
	case "bids.placed":
		var e BidPlaced
		if err := json.Unmarshal(payload, &e); err != nil {
			return nil, err
		}
		return e, nil
	default:
		return nil, fmt.Errorf("unknown topic %s", topic)
	}
}
