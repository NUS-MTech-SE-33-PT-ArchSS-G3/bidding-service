package list_bids

import (
	"context"
	"fmt"
	"kei-services/pkg/middleware"

	"go.uber.org/zap"
)

type Service struct {
	bidReadRepo IBidReadRepository
	log         *zap.Logger
}

var _ IService = (*Service)(nil)

type Deps struct {
	BidReadRepo IBidReadRepository
}

func NewService(d Deps, log *zap.Logger) *Service {
	return &Service{
		bidReadRepo: d.BidReadRepo,
		log:         log,
	}
}

func (s *Service) Handle(ctx context.Context, q Query) (*Result, error) {
	log := middleware.LoggerFrom(ctx, s.log).With(
		zap.String("auctionId", q.AuctionID),
		zap.Int("limit", q.Limit),
		zap.String("direction", q.Direction.String()),
	)

	// sanitize
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	asc := q.Direction == DirectionAsc

	var after *Cursor
	if q.Cursor != "" {
		c, err := decodeCursor(q.Cursor)
		if err != nil {
			return nil, ErrInvalidCursor
		}
		after = c
	}

	log.Debug("list bids: resolved query arguments",
		zap.Int("limit", limit),
		zap.Bool("asc", asc),
		zap.Any("after", after),
	)

	log.Debug("fetching bids from repo")
	items, hasMore, next, err := s.bidReadRepo.ListByAuction(ctx, q.AuctionID, after, limit, asc)
	if err != nil {
		log.Warn("list bids failed", zap.Error(err))
		// let infra map "not found" to ErrAuctionNotFound
		return nil, fmt.Errorf("list bids: %w", err)
	}

	nextStr, err := encodeCursor(next)
	if err != nil {
		// encoding failure shouldn't 500 the whole call, log and fall back to end of list
		log.Warn("encode next cursor failed", zap.Error(err))
		nextStr = nil
		hasMore = false
	}

	nc := ""
	if nextStr != nil {
		nc = *nextStr
	}
	log.Debug("list bids: returning result",
		zap.Int("items", len(items)),
		zap.Bool("hasMore", hasMore),
		zap.String("nextCursor", nc))

	return &Result{
		Items:      items,
		HasMore:    hasMore,
		NextCursor: nextStr,
	}, nil
}
