package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateBid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		auction     *AuctionMetadata
		amount      float64
		lastBid     *LastAcceptedBid
		expectedErr error
	}{
		{
			name:        "nil auction returns error",
			auction:     nil,
			amount:      100.0,
			lastBid:     nil,
			expectedErr: ErrAuctionNotFound,
		},
		{
			name: "closed auction returns error",
			auction: &AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionClose,
				EndsAt:        now.Add(1 * time.Hour),
				StartingPrice: 100.0,
				MinIncrement:  10.0,
			},
			amount:      110.0,
			lastBid:     nil,
			expectedErr: ErrAuctionClosed,
		},
		{
			name: "zero amount returns error",
			auction: &AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionOpen,
				EndsAt:        now.Add(1 * time.Hour),
				StartingPrice: 100.0,
				MinIncrement:  10.0,
			},
			amount:      0,
			lastBid:     nil,
			expectedErr: ErrInvalidAmount,
		},
		{
			name: "negative amount returns error",
			auction: &AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionOpen,
				EndsAt:        now.Add(1 * time.Hour),
				StartingPrice: 100.0,
				MinIncrement:  10.0,
			},
			amount:      -50.0,
			lastBid:     nil,
			expectedErr: ErrInvalidAmount,
		},
		{
			name: "first bid below starting price returns error",
			auction: &AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionOpen,
				EndsAt:        now.Add(1 * time.Hour),
				StartingPrice: 100.0,
				MinIncrement:  10.0,
				CurrentPrice:  0,
			},
			amount:      90.0,
			lastBid:     nil,
			expectedErr: ErrBelowMinIncrement,
		},
		{
			name: "first bid at starting price is valid",
			auction: &AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionOpen,
				EndsAt:        now.Add(1 * time.Hour),
				StartingPrice: 100.0,
				MinIncrement:  10.0,
				CurrentPrice:  0,
			},
			amount:      100.0,
			lastBid:     nil,
			expectedErr: nil,
		},
		{
			name: "first bid above starting price is valid",
			auction: &AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionOpen,
				EndsAt:        now.Add(1 * time.Hour),
				StartingPrice: 100.0,
				MinIncrement:  10.0,
				CurrentPrice:  0,
			},
			amount:      150.0,
			lastBid:     nil,
			expectedErr: nil,
		},
		{
			name: "bid below last accepted bid plus increment returns error",
			auction: &AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionOpen,
				EndsAt:        now.Add(1 * time.Hour),
				StartingPrice: 100.0,
				MinIncrement:  10.0,
				CurrentPrice:  120.0,
			},
			amount: 125.0,
			lastBid: &LastAcceptedBid{
				Price:   120.0,
				Version: 1,
			},
			expectedErr: ErrBelowMinIncrement,
		},
		{
			name: "bid at minimum required amount is valid",
			auction: &AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionOpen,
				EndsAt:        now.Add(1 * time.Hour),
				StartingPrice: 100.0,
				MinIncrement:  10.0,
				CurrentPrice:  120.0,
			},
			amount: 130.0,
			lastBid: &LastAcceptedBid{
				Price:   120.0,
				Version: 1,
			},
			expectedErr: nil,
		},
		{
			name: "bid above minimum required amount is valid",
			auction: &AuctionMetadata{
				AuctionID:     "auction-1",
				Status:        AuctionOpen,
				EndsAt:        now.Add(1 * time.Hour),
				StartingPrice: 100.0,
				MinIncrement:  10.0,
				CurrentPrice:  120.0,
			},
			amount: 150.0,
			lastBid: &LastAcceptedBid{
				Price:   120.0,
				Version: 1,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBid(tt.auction, tt.amount, tt.lastBid)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMakeLastAcceptedBid(t *testing.T) {
	tests := []struct {
		name         string
		auction      *AuctionMetadata
		latestAmount *float64
		latestSeq    *int64
		expectedBid  LastAcceptedBid
	}{
		{
			name: "uses auction metadata when no latest bid",
			auction: &AuctionMetadata{
				CurrentPrice: 100.0,
				Version:      1,
			},
			latestAmount: nil,
			latestSeq:    nil,
			expectedBid: LastAcceptedBid{
				Price:   100.0,
				Version: 1,
			},
		},
		{
			name: "uses latest amount when higher than current price",
			auction: &AuctionMetadata{
				CurrentPrice: 100.0,
				Version:      1,
			},
			latestAmount: func() *float64 { v := 150.0; return &v }(),
			latestSeq:    func() *int64 { v := int64(2); return &v }(),
			expectedBid: LastAcceptedBid{
				Price:   150.0,
				Version: 2,
			},
		},
		{
			name: "keeps auction price when latest amount is lower",
			auction: &AuctionMetadata{
				CurrentPrice: 150.0,
				Version:      2,
			},
			latestAmount: func() *float64 { v := 100.0; return &v }(),
			latestSeq:    func() *int64 { v := int64(1); return &v }(),
			expectedBid: LastAcceptedBid{
				Price:   150.0,
				Version: 2,
			},
		},
		{
			name: "uses latest seq when higher than version",
			auction: &AuctionMetadata{
				CurrentPrice: 100.0,
				Version:      1,
			},
			latestAmount: func() *float64 { v := 100.0; return &v }(),
			latestSeq:    func() *int64 { v := int64(5); return &v }(),
			expectedBid: LastAcceptedBid{
				Price:   100.0,
				Version: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeLastAcceptedBid(tt.auction, tt.latestAmount, tt.latestSeq)

			assert.Equal(t, tt.expectedBid.Price, result.Price)
			assert.Equal(t, tt.expectedBid.Version, result.Version)
		})
	}
}

func TestMinNextPrice(t *testing.T) {
	tests := []struct {
		name     string
		bid      LastAcceptedBid
		auction  *AuctionMetadata
		expected float64
	}{
		{
			name: "uses starting price when no prior bid",
			bid: LastAcceptedBid{
				Price:   0,
				Version: 0,
			},
			auction: &AuctionMetadata{
				StartingPrice: 100.0,
				MinIncrement:  10.0,
			},
			expected: 100.0,
		},
		{
			name: "adds increment to last bid price",
			bid: LastAcceptedBid{
				Price:   120.0,
				Version: 1,
			},
			auction: &AuctionMetadata{
				StartingPrice: 100.0,
				MinIncrement:  10.0,
			},
			expected: 130.0,
		},
		{
			name: "handles decimal increments",
			bid: LastAcceptedBid{
				Price:   125.50,
				Version: 1,
			},
			auction: &AuctionMetadata{
				StartingPrice: 100.0,
				MinIncrement:  2.50,
			},
			expected: 128.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinNextPrice(tt.bid, tt.auction)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyAccepted(t *testing.T) {
	t.Run("updates current price and increments version", func(t *testing.T) {
		auction := AuctionMetadata{
			AuctionID:     "auction-1",
			Status:        AuctionOpen,
			StartingPrice: 100.0,
			CurrentPrice:  120.0,
			MinIncrement:  10.0,
			Version:       2,
		}

		bid := &Bid{
			ID:        "bid-1",
			AuctionID: "auction-1",
			BidderID:  "bidder-1",
			Amount:    150.0,
		}

		result := ApplyAccepted(auction, bid)

		assert.Equal(t, 150.0, result.CurrentPrice)
		assert.Equal(t, 3, result.Version)
		assert.Equal(t, auction.AuctionID, result.AuctionID)
		assert.Equal(t, auction.StartingPrice, result.StartingPrice)
		assert.Equal(t, auction.MinIncrement, result.MinIncrement)

		// Verify original is not modified
		assert.Equal(t, 120.0, auction.CurrentPrice)
		assert.Equal(t, 2, auction.Version)
	})
}

func TestSafeLess(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		eps      float64
		expected bool
	}{
		{
			name:     "a clearly less than b",
			a:        100.0,
			b:        110.0,
			eps:      1e-9,
			expected: true,
		},
		{
			name:     "a equal to b within epsilon",
			a:        100.0,
			b:        100.0000000001,
			eps:      1e-9,
			expected: false,
		},
		{
			name:     "a greater than b",
			a:        110.0,
			b:        100.0,
			eps:      1e-9,
			expected: false,
		},
		{
			name:     "handles epsilon boundary",
			a:        100.0,
			b:        100.01,
			eps:      0.001,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeLess(tt.a, tt.b, tt.eps)
			assert.Equal(t, tt.expected, result)
		})
	}
}
