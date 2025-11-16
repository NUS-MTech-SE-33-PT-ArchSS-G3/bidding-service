package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuctionMetadata_IsOpen(t *testing.T) {
	tests := []struct {
		name     string
		status   AuctionStatus
		expected bool
	}{
		{
			name:     "open auction returns true",
			status:   AuctionOpen,
			expected: true,
		},
		{
			name:     "closed auction returns false",
			status:   AuctionClose,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auction := AuctionMetadata{
				AuctionID: "auction-1",
				Status:    tt.status,
			}

			result := auction.IsOpen()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuctionMetadata_MinNextBid(t *testing.T) {
	tests := []struct {
		name          string
		currentPrice  float64
		startingPrice float64
		minIncrement  float64
		expected      float64
	}{
		{
			name:          "no bids yet returns starting price",
			currentPrice:  0,
			startingPrice: 100.0,
			minIncrement:  10.0,
			expected:      100.0,
		},
		{
			name:          "with existing bid returns current price plus increment",
			currentPrice:  150.0,
			startingPrice: 100.0,
			minIncrement:  10.0,
			expected:      160.0,
		},
		{
			name:          "handles decimal prices",
			currentPrice:  125.50,
			startingPrice: 100.0,
			minIncrement:  5.25,
			expected:      130.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auction := AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionOpen,
				EndsAt:        time.Now().Add(1 * time.Hour),
				StartingPrice: tt.startingPrice,
				CurrentPrice:  tt.currentPrice,
				MinIncrement:  tt.minIncrement,
			}

			result := auction.MinNextBid()
			assert.Equal(t, tt.expected, result)
		})
	}
}
