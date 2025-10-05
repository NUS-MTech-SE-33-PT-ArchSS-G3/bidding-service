package place_bid

import (
	"context"
	"fmt"
	"kei-services/internal/bid-command/application"
	"kei-services/internal/bid-command/domain"

	"go.uber.org/zap"
)

type Service struct {
	bidRepo domain.IBidRepository
	cache   domain.IAuctionMetadataStore
	pub     domain.IBidsPlacedPublisher
	tx      application.ITxManager
	clock   domain.IClock
	log     *zap.Logger
}

var _ IService = (*Service)(nil)

type Deps struct {
	BidRepo domain.IBidRepository
	Cache   domain.IAuctionMetadataStore
	Pub     domain.IBidsPlacedPublisher
	Tx      application.ITxManager
	Clock   domain.IClock
}

func NewService(d Deps, log *zap.Logger) *Service {
	return &Service{
		bidRepo: d.BidRepo,
		cache:   d.Cache,
		pub:     d.Pub,
		tx:      d.Tx,
		clock:   d.Clock,
		log:     log,
	}
}

// Handle processes the PlaceBid command
func (s *Service) Handle(ctx context.Context, cmd Command) (*Result, error) {
	s.log.Info("placing bid", zap.String("auction_id", cmd.AuctionID), zap.String("bidder_id", cmd.BidderID), zap.Float64("amount", cmd.Amount))
	// fast pre-check using cache
	auction, err := s.cache.Get(ctx, cmd.AuctionID)
	if err != nil {
		s.log.Warn("cache miss or error", zap.String("auction_id", cmd.AuctionID), zap.Error(err))
		return nil, fmt.Errorf("load auction meta: %w", err)
	}
	if err = domain.ValidateBid(auction, cmd.Amount, nil); err != nil {
		s.log.Warn("validate bid failed", zap.String("auction_id", cmd.AuctionID), zap.Error(err))
		return nil, err
	}

	bid := domain.NewBid(cmd.AuctionID, cmd.BidderID, cmd.Amount, s.clock.Now().UTC())

	// authoritative check against latest bid inside DB transaction
	var out *Result
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		// get latest bid with row lock to serialize concurrent bids
		latest, err := s.bidRepo.LatestForUpdate(ctx, cmd.AuctionID)
		if err != nil {
			s.log.Warn("get latest for update", zap.String("auction_id", cmd.AuctionID), zap.Error(err))
			return err
		}

		// build last accepted bid from cache + latest
		var la *float64
		var ls *int64
		if latest != nil {
			la = &latest.Amount
			ls = &latest.Seq
		}
		b := domain.MakeLastAcceptedBid(auction, la, ls)

		// validate again using last accepted bid as baseline
		if err = domain.ValidateBid(auction, cmd.Amount, &b); err != nil {
			return err
		}

		// persist the bid
		id, seq, err := s.bidRepo.Insert(ctx, bid)
		if err != nil {
			s.log.Warn("insert bid failed", zap.String("auction_id", cmd.AuctionID), zap.Error(err))
			return err
		}
		if id != "" {
			bid = bid.WithID(id)
		}

		// Compute the snapshot after acceptance
		after := domain.ApplyAccepted(domain.AuctionMetadata{
			AuctionID:     auction.AuctionID,
			Status:        auction.Status,
			EndsAt:        auction.EndsAt,
			StartingPrice: auction.StartingPrice,
			CurrentPrice:  b.Price,
			MinIncrement:  auction.MinIncrement,
			Version:       b.Version,
		}, bid)

		version := after.Version
		if seq > 0 {
			version = int(seq)
		}

		out = &Result{
			BidID:        bid.ID,
			AuctionID:    bid.AuctionID,
			BidderID:     bid.BidderID,
			CurrentPrice: bid.Amount,
			MinNextBid:   bid.Amount + auction.MinIncrement,
			Version:      version,
			At:           bid.At,
		}
		return nil
	})
	if err != nil {
		s.log.Warn("place bid tx failed", zap.String("auction_id", cmd.AuctionID), zap.Error(err))
		return nil, err
	}

	// publish after commit
	// todo: maybe use outbox pattern if have time
	_ = s.pub.Publish(ctx, domain.BidPlaced{
		AuctionID: out.AuctionID,
		BidID:     out.BidID,
		BidderID:  out.BidderID,
		Amount:    out.CurrentPrice,
		At:        out.At,
	})

	return out, nil
}
