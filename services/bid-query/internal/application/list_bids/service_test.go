package list_bids

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock implementations
type MockBidReadRepository struct {
	mock.Mock
}

func (m *MockBidReadRepository) ListByAuction(ctx context.Context, auctionID string, after *Cursor, limit int, asc bool) (
	items []Item, hasMore bool, next *Cursor, err error) {
	args := m.Called(ctx, auctionID, after, limit, asc)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Get(2).(*Cursor), args.Error(3)
	}
	return args.Get(0).([]Item), args.Bool(1), args.Get(2).(*Cursor), args.Error(3)
}

func TestService_Handle_Success(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	items := []Item{
		{
			BidID:     "bid-1",
			AuctionID: "auction-1",
			BidderID:  "bidder-1",
			Amount:    120.0,
			At:        fixedTime,
		},
		{
			BidID:     "bid-2",
			AuctionID: "auction-1",
			BidderID:  "bidder-2",
			Amount:    130.0,
			At:        fixedTime.Add(1 * time.Minute),
		},
	}

	nextCursor := &Cursor{
		At: fixedTime.Add(1 * time.Minute),
		ID: "bid-2",
	}

	query := Query{
		AuctionID: "auction-1",
		Cursor:    "",
		Limit:     50,
		Direction: DirectionDesc,
	}

	mockRepo := new(MockBidReadRepository)
	mockRepo.On("ListByAuction", ctx, "auction-1", (*Cursor)(nil), 50, false).
		Return(items, true, nextCursor, nil)

	deps := Deps{
		BidReadRepo: mockRepo,
	}

	service := NewService(deps, zap.NewNop())

	// Execute
	result, err := service.Handle(ctx, query)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Items))
	assert.True(t, result.HasMore)
	assert.NotNil(t, result.NextCursor)

	mockRepo.AssertExpectations(t)
}

func TestService_Handle_WithCursor(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	cursor := &Cursor{
		At: fixedTime,
		ID: "bid-1",
	}
	encodedCursor, _ := encodeCursor(cursor)

	items := []Item{
		{
			BidID:     "bid-3",
			AuctionID: "auction-1",
			BidderID:  "bidder-3",
			Amount:    140.0,
			At:        fixedTime.Add(2 * time.Minute),
		},
	}

	query := Query{
		AuctionID: "auction-1",
		Cursor:    *encodedCursor,
		Limit:     50,
		Direction: DirectionDesc,
	}

	mockRepo := new(MockBidReadRepository)
	mockRepo.On("ListByAuction", ctx, "auction-1", cursor, 50, false).
		Return(items, false, (*Cursor)(nil), nil)

	deps := Deps{
		BidReadRepo: mockRepo,
	}

	service := NewService(deps, zap.NewNop())

	// Execute
	result, err := service.Handle(ctx, query)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(result.Items))
	assert.False(t, result.HasMore)
	assert.Nil(t, result.NextCursor)

	mockRepo.AssertExpectations(t)
}

func TestService_Handle_InvalidCursor(t *testing.T) {
	ctx := context.Background()

	query := Query{
		AuctionID: "auction-1",
		Cursor:    "invalid-cursor",
		Limit:     50,
		Direction: DirectionDesc,
	}

	deps := Deps{
		BidReadRepo: new(MockBidReadRepository),
	}

	service := NewService(deps, zap.NewNop())

	// Execute
	result, err := service.Handle(ctx, query)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrInvalidCursor))
}

func TestService_Handle_LimitSanitization(t *testing.T) {
	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int
	}{
		{
			name:          "zero limit defaults to 50",
			inputLimit:    0,
			expectedLimit: 50,
		},
		{
			name:          "negative limit defaults to 50",
			inputLimit:    -10,
			expectedLimit: 50,
		},
		{
			name:          "limit above 200 capped at 200",
			inputLimit:    500,
			expectedLimit: 200,
		},
		{
			name:          "valid limit used as-is",
			inputLimit:    75,
			expectedLimit: 75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			query := Query{
				AuctionID: "auction-1",
				Cursor:    "",
				Limit:     tt.inputLimit,
				Direction: DirectionDesc,
			}

			mockRepo := new(MockBidReadRepository)
			mockRepo.On("ListByAuction", ctx, "auction-1", (*Cursor)(nil), tt.expectedLimit, false).
				Return([]Item{}, false, (*Cursor)(nil), nil)

			deps := Deps{
				BidReadRepo: mockRepo,
			}

			service := NewService(deps, zap.NewNop())

			// Execute
			_, err := service.Handle(ctx, query)

			// Assert
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Handle_DirectionHandling(t *testing.T) {
	tests := []struct {
		name        string
		direction   Direction
		expectedAsc bool
	}{
		{
			name:        "DirectionAsc sets asc to true",
			direction:   DirectionAsc,
			expectedAsc: true,
		},
		{
			name:        "DirectionDesc sets asc to false",
			direction:   DirectionDesc,
			expectedAsc: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			query := Query{
				AuctionID: "auction-1",
				Cursor:    "",
				Limit:     50,
				Direction: tt.direction,
			}

			mockRepo := new(MockBidReadRepository)
			mockRepo.On("ListByAuction", ctx, "auction-1", (*Cursor)(nil), 50, tt.expectedAsc).
				Return([]Item{}, false, (*Cursor)(nil), nil)

			deps := Deps{
				BidReadRepo: mockRepo,
			}

			service := NewService(deps, zap.NewNop())

			// Execute
			_, err := service.Handle(ctx, query)

			// Assert
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Handle_AuctionNotFound(t *testing.T) {
	ctx := context.Background()

	query := Query{
		AuctionID: "auction-1",
		Cursor:    "",
		Limit:     50,
		Direction: DirectionDesc,
	}

	mockRepo := new(MockBidReadRepository)
	mockRepo.On("ListByAuction", ctx, "auction-1", (*Cursor)(nil), 50, false).
		Return(nil, false, (*Cursor)(nil), ErrAuctionNotFound)

	deps := Deps{
		BidReadRepo: mockRepo,
	}

	service := NewService(deps, zap.NewNop())

	// Execute
	result, err := service.Handle(ctx, query)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}
