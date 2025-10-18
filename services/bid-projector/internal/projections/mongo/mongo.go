package mongo

import (
	"context"
	"errors"
	"kei-services/services/bid-projector/internal/events"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Projection struct {
	log *zap.Logger
	db  *mongo.Database
}

func NewProjection(db *mongo.Database, log *zap.Logger) *Projection {
	return &Projection{
		log: log,
		db:  db,
	}
}

func (p *Projection) OnBidsPlaced(ctx context.Context, evt events.BidPlaced) error {
	bidsColl := p.db.Collection("bids_history")

	if err := p.insertBidDoc(ctx, bidsColl, evt); err != nil {
		if isDupKey(err) {
			p.log.Info("duplicate bid, skipping",
				zap.String("bidID", evt.BidID),
				zap.String("auctionID", evt.AuctionID))
			return nil
		}
		return err
	}

	return nil
}

func (p *Projection) insertBidDoc(ctx context.Context, coll *mongo.Collection, evt events.BidPlaced) error {
	cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := coll.InsertOne(cctx, BidDoc{
		AuctionID: evt.AuctionID,
		BidID:     evt.BidID,
		BidderID:  evt.BidderID,
		Amount:    evt.Amount,
		At:        evt.At.UTC(),
	})
	return err
}

func (p *Projection) EnsureIndexes(ctx context.Context) error {
	bids := p.db.Collection("bids_history")

	// bids_history: unique bid_id + auction_id, at DESC for listing
	_, err := bids.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "bidId", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_bid_id"),
		},
		{
			Keys:    bson.D{{Key: "auctionId", Value: 1}, {Key: "at", Value: -1}},
			Options: options.Index().SetName("auction_at_desc"),
		},
		{
			Keys:    bson.D{{Key: "auctionId", Value: 1}, {Key: "at", Value: 1}, {Key: "bidId", Value: 1}},
			Options: options.Index().SetName("auction_at_asc_bid_asc"),
		},
	})
	return err
}

func isDupKey(err error) bool {
	var we *mongo.WriteException
	if errors.As(err, &we) {
		for _, e := range we.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}
	return false
}
