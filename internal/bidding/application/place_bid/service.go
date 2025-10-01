package place_bid

import (
	"context"
	"fmt"
	"kei-services/internal/bidding/application"
	"kei-services/internal/bidding/domain"

	"go.uber.org/zap"
)

type Service struct {
	BidRepo application.IBidRepository
	Cache   application.IAuctionMetadataStore
	Pub     application.IBidsPlacedPublisher
	Tx      application.ITxManager
	Clock   application.IClock
	Log     *zap.Logger
}

var _ IService = (*Service)(nil)

type Deps struct {
	BidRepo application.IBidRepository
	Cache   application.IAuctionMetadataStore
	Pub     application.IBidsPlacedPublisher
	Tx      application.ITxManager
	Clock   application.IClock
}

func NewService(d Deps, log *zap.Logger) *Service {
	return &Service{
		BidRepo: d.BidRepo,
		Cache:   d.Cache,
		Pub:     d.Pub,
		Tx:      d.Tx,
		Clock:   d.Clock,
		Log:     log,
	}
}

// Handle processes the PlaceBid command
func (s *Service) Handle(ctx context.Context, cmd Command) (*Result, error) {
	// fast pre-check using cache
	auction, err := s.Cache.Get(ctx, cmd.AuctionID)
	if err != nil {
		return nil, fmt.Errorf("load auction meta: %w", err)
	}
	if err = domain.ValidateBid(auction, cmd.Amount, nil); err != nil {
		return nil, err
	}

	bid := domain.NewBid(cmd.AuctionID, cmd.BidderID, cmd.Amount, s.Clock.Now().UTC())

	// authoritative check against latest bid inside DB transaction
	var out *Result
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		// get latest bid with row lock to serialize concurrent bids
		latest, err := s.BidRepo.LatestForUpdate(ctx, cmd.AuctionID)
		if err != nil {
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
		id, seq, err := s.BidRepo.Insert(ctx, bid)
		if err != nil {
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
		return nil, err
	}

	// publish after commit
	// todo: maybe use outbox pattern if have time
	_ = s.Pub.Publish(ctx, domain.BidPlaced{
		AuctionID: out.AuctionID,
		BidID:     out.BidID,
		BidderID:  out.BidderID,
		Amount:    out.CurrentPrice,
		At:        out.At,
	})

	return out, nil

}
