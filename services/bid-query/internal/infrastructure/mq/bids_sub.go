package mq

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type BidDoc struct {
	AuctionID string    `bson:"auction_id"`
	BidID     string    `bson:"bid_id"`
	BidderID  string    `bson:"bidder_id"`
	Amount    float64   `bson:"amount"`
	At        time.Time `bson:"at"`
}

type BidPlaced struct {
	AuctionID string    `json:"auctionId"`
	BidID     string    `json:"bidId"`
	BidderID  string    `json:"bidderId"`
	Amount    float64   `json:"amount"`
	At        time.Time `json:"at"`
}

type BidPlacedProjector struct {
	Log             *zap.Logger
	BidPlacedReader *kafka.Reader // subscribed to "bids.placed"

	DB    *mongo.Database
	Batch int
}

func NewBidPlacedProjector(bidPlacedReader *kafka.Reader, db *mongo.Database, batch int, log *zap.Logger) *BidPlacedProjector {
	return &BidPlacedProjector{
		Log:             log,
		BidPlacedReader: bidPlacedReader,
		DB:              db,
		Batch:           batch,
	}
}

func (p *BidPlacedProjector) Run(ctx context.Context) error {
	if err := p.EnsureIndexes(ctx); err != nil {
		return err
	}

	bids := p.DB.Collection("bids_history")
	auct := p.DB.Collection("auctions_view")

	defer func() {
		_ = p.BidPlacedReader.Close()
	}()

	processed := 0
	for {
		m, err := p.BidPlacedReader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil // normal shutdown
			}
			return err
		}

		var evt BidPlaced
		if uerr := json.Unmarshal(m.Value, &evt); uerr != nil {
			p.Log.Warn("bids.placed: bad payload", zap.Error(uerr))
			_ = p.BidPlacedReader.CommitMessages(ctx, m) // skip poison
			continue
		}

		if ierr := p.insertBidDoc(ctx, bids, evt); ierr != nil {
			if !isDupKey(ierr) { // ignore if already exists
				p.Log.Error("bids_history insert", zap.Error(ierr), zap.String("auctionId", evt.AuctionID))
				continue // retry later
			}
		}

		if uerr := p.upsertAuctionView(ctx, auct, evt); uerr != nil {
			p.Log.Error("auctions_view upsert", zap.Error(uerr), zap.String("auctionId", evt.AuctionID))
			continue // retry later
		}

		// commit
		if err := p.BidPlacedReader.CommitMessages(ctx, m); err != nil {
			p.Log.Warn("kafka commit", zap.Error(err))
		}

		processed++
		// todo after moving to project service, batch process
	}
}

func (p *BidPlacedProjector) insertBidDoc(ctx context.Context, coll *mongo.Collection, evt BidPlaced) error {
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

func (p *BidPlacedProjector) upsertAuctionView(ctx context.Context, coll *mongo.Collection, evt BidPlaced) error {
	cctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := coll.UpdateOne(cctx, bson.M{"_id": evt.AuctionID},
		bson.M{"$set": bson.M{
			"auction_id":     evt.AuctionID,
			"current_price":  evt.Amount,
			"last_bidder_id": evt.BidderID,
			"updated_at":     time.Now().UTC(),
		}},
		options.Update().SetUpsert(true),
	)
	return err
}

func (p *BidPlacedProjector) EnsureIndexes(ctx context.Context) error {
	bids := p.DB.Collection("bids_history")

	// bids_history: unique bid_id + auction_id, at DESC for listing
	_, err := bids.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "bid_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_bid_id"),
		},
		{
			Keys:    bson.D{{Key: "auction_id", Value: 1}, {Key: "at", Value: -1}},
			Options: options.Index().SetName("auction_at_desc"),
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
