package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBid(t *testing.T) {
	tests := []struct {
		name      string
		auctionID string
		bidderID  string
		amount    float64
		at        time.Time
	}{
		{
			name:      "creates bid with valid data",
			auctionID: "auction-1",
			bidderID:  "bidder-1",
			amount:    100.50,
			at:        time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:      "creates bid with different timezone converts to UTC",
			auctionID: "auction-2",
			bidderID:  "bidder-2",
			amount:    200.00,
			at:        time.Date(2024, 1, 1, 12, 0, 0, 0, time.FixedZone("EST", -5*3600)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bid := NewBid(tt.auctionID, tt.bidderID, tt.amount, tt.at)

			assert.Equal(t, tt.auctionID, bid.AuctionID)
			assert.Equal(t, tt.bidderID, bid.BidderID)
			assert.Equal(t, tt.amount, bid.Amount)
			assert.Equal(t, time.UTC, bid.At.Location())
			assert.Empty(t, bid.ID)
		})
	}
}

func TestBid_WithID(t *testing.T) {
	t.Run("sets ID on bid copy", func(t *testing.T) {
		original := NewBid("auction-1", "bidder-1", 100.00, time.Now())
		newID := "bid-123"

		result := original.WithID(newID)

		assert.Equal(t, newID, result.ID)
		assert.Empty(t, original.ID, "original bid should not be modified")
		assert.Equal(t, original.AuctionID, result.AuctionID)
		assert.Equal(t, original.BidderID, result.BidderID)
		assert.Equal(t, original.Amount, result.Amount)
	})
}
