package place_bid

import (
	"context"
	"errors"
	"kei-services/services/bid-command/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Mock implementations
type MockBidRepository struct {
	mock.Mock
}

func (m *MockBidRepository) Insert(ctx context.Context, b *domain.Bid) (string, int64, error) {
	args := m.Called(ctx, b)
	return args.String(0), args.Get(1).(int64), args.Error(2)
}

func (m *MockBidRepository) LatestForUpdate(ctx context.Context, auctionID string) (*domain.LatestBid, error) {
	args := m.Called(ctx, auctionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LatestBid), args.Error(1)
}

type MockAuctionMetadataStore struct {
	mock.Mock
}

func (m *MockAuctionMetadataStore) Get(ctx context.Context, auctionID string) (*domain.AuctionMetadata, error) {
	args := m.Called(ctx, auctionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuctionMetadata), args.Error(1)
}

type MockBidsPlacedPublisher struct {
	mock.Mock
}

func (m *MockBidsPlacedPublisher) Publish(ctx context.Context, evt domain.BidPlaced) error {
	args := m.Called(ctx, evt)
	return args.Error(0)
}

type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	if args.Error(0) != nil {
		return args.Error(0)
	}
	// Execute the function to test transaction logic
	return fn(ctx)
}

type MockClock struct {
	mock.Mock
}

func (m *MockClock) Now() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func TestService_Handle_Success(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	auction := &domain.AuctionMetadata{
		AuctionID:     "auction-1",
		Status:        domain.AuctionOpen,
		EndsAt:        fixedTime.Add(1 * time.Hour),
		StartingPrice: 100.0,
		CurrentPrice:  0,
		MinIncrement:  10.0,
		Version:       1,
	}

	cmd := Command{
		AuctionID: "auction-1",
		BidderID:  "bidder-1",
		Amount:    120.0,
	}

	// Setup mocks
	mockCache := new(MockAuctionMetadataStore)
	mockRepo := new(MockBidRepository)
	mockPub := new(MockBidsPlacedPublisher)
	mockTx := new(MockTxManager)
	mockClock := new(MockClock)

	mockCache.On("Get", ctx, "auction-1").Return(auction, nil)
	mockClock.On("Now").Return(fixedTime)
	mockTx.On("WithinTx", ctx, mock.Anything).Return(nil)
	mockRepo.On("LatestForUpdate", ctx, "auction-1").Return(nil, nil)
	mockRepo.On("Insert", ctx, mock.AnythingOfType("*domain.Bid")).Return("bid-123", int64(1), nil)
	mockPub.On("Publish", ctx, mock.AnythingOfType("domain.BidPlaced")).Return(nil)

	deps := Deps{
		BidRepo: mockRepo,
		Cache:   mockCache,
		Pub:     mockPub,
		Tx:      mockTx,
		Clock:   mockClock,
	}

	service := NewService(deps, zap.NewNop())

	// Execute
	result, err := service.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "bid-123", result.BidID)
	assert.Equal(t, "auction-1", result.AuctionID)
	assert.Equal(t, "bidder-1", result.BidderID)
	assert.Equal(t, 120.0, result.CurrentPrice)
	assert.Equal(t, 130.0, result.MinNextBid)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockClock.AssertExpectations(t)
}

func TestService_Handle_AuctionNotFound(t *testing.T) {
	ctx := context.Background()

	cmd := Command{
		AuctionID: "auction-1",
		BidderID:  "bidder-1",
		Amount:    120.0,
	}

	mockCache := new(MockAuctionMetadataStore)
	mockCache.On("Get", ctx, "auction-1").Return(nil, domain.ErrAuctionNotFound)

	deps := Deps{
		BidRepo: new(MockBidRepository),
		Cache:   mockCache,
		Pub:     new(MockBidsPlacedPublisher),
		Tx:      new(MockTxManager),
		Clock:   new(MockClock),
	}

	service := NewService(deps, zap.NewNop())

	// Execute
	result, err := service.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domain.ErrAuctionNotFound))

	mockCache.AssertExpectations(t)
}

func TestService_Handle_AuctionClosed(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	auction := &domain.AuctionMetadata{
		AuctionID:     "auction-1",
		Status:        domain.AuctionClose,
		EndsAt:        fixedTime.Add(-1 * time.Hour),
		StartingPrice: 100.0,
		CurrentPrice:  120.0,
		MinIncrement:  10.0,
		Version:       1,
	}

	cmd := Command{
		AuctionID: "auction-1",
		BidderID:  "bidder-1",
		Amount:    150.0,
	}

	mockCache := new(MockAuctionMetadataStore)
	mockCache.On("Get", ctx, "auction-1").Return(auction, nil)

	deps := Deps{
		BidRepo: new(MockBidRepository),
		Cache:   mockCache,
		Pub:     new(MockBidsPlacedPublisher),
		Tx:      new(MockTxManager),
		Clock:   new(MockClock),
	}

	service := NewService(deps, zap.NewNop())

	// Execute
	result, err := service.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domain.ErrAuctionClosed))

	mockCache.AssertExpectations(t)
}

func TestService_Handle_BelowMinIncrement(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	auction := &domain.AuctionMetadata{
		AuctionID:     "auction-1",
		Status:        domain.AuctionOpen,
		EndsAt:        fixedTime.Add(1 * time.Hour),
		StartingPrice: 100.0,
		CurrentPrice:  120.0,
		MinIncrement:  10.0,
		Version:       1,
	}

	cmd := Command{
		AuctionID: "auction-1",
		BidderID:  "bidder-1",
		Amount:    125.0, // Below minimum of 130.0
	}

	mockCache := new(MockAuctionMetadataStore)

	// The service does a fast pre-check with cache that will fail early
	mockCache.On("Get", ctx, "auction-1").Return(auction, nil)

	deps := Deps{
		BidRepo: new(MockBidRepository),
		Cache:   mockCache,
		Pub:     new(MockBidsPlacedPublisher),
		Tx:      new(MockTxManager),
		Clock:   new(MockClock),
	}

	service := NewService(deps, zap.NewNop())

	// Execute
	result, err := service.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, domain.ErrBelowMinIncrement))

	mockCache.AssertExpectations(t)
}

func TestService_Handle_PublishFails(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	auction := &domain.AuctionMetadata{
		AuctionID:     "auction-1",
		Status:        domain.AuctionOpen,
		EndsAt:        fixedTime.Add(1 * time.Hour),
		StartingPrice: 100.0,
		CurrentPrice:  0,
		MinIncrement:  10.0,
		Version:       1,
	}

	cmd := Command{
		AuctionID: "auction-1",
		BidderID:  "bidder-1",
		Amount:    120.0,
	}

	mockCache := new(MockAuctionMetadataStore)
	mockRepo := new(MockBidRepository)
	mockPub := new(MockBidsPlacedPublisher)
	mockTx := new(MockTxManager)
	mockClock := new(MockClock)

	mockCache.On("Get", ctx, "auction-1").Return(auction, nil)
	mockClock.On("Now").Return(fixedTime)
	mockTx.On("WithinTx", ctx, mock.Anything).Return(nil)
	mockRepo.On("LatestForUpdate", ctx, "auction-1").Return(nil, nil)
	mockRepo.On("Insert", ctx, mock.AnythingOfType("*domain.Bid")).Return("bid-123", int64(1), nil)
	mockPub.On("Publish", ctx, mock.AnythingOfType("domain.BidPlaced")).Return(errors.New("publish error"))

	deps := Deps{
		BidRepo: mockRepo,
		Cache:   mockCache,
		Pub:     mockPub,
		Tx:      mockTx,
		Clock:   mockClock,
	}

	service := NewService(deps, zap.NewNop())

	// Execute
	result, err := service.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "publish failed")

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockClock.AssertExpectations(t)
}
