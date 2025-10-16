package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type BidDoc struct {
	AuctionID string    `bson:"auctionId"`
	BidID     string    `bson:"bidId"`
	BidderID  string    `bson:"bidderId"`
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
	DB              *mongo.Database
	Batch           int
}

func NewBidPlacedProjector(bidPlacedReader *kafka.Reader, db *mongo.Database, batch int, log *zap.Logger) *BidPlacedProjector {
	log.Info("NewBidPlacedProjector",
		zap.Int("batch", batch),
		zap.String("kafka_topic", bidPlacedReader.Config().Topic),
		zap.String("mongo_db", db.Name()))

	return &BidPlacedProjector{
		Log:             log,
		BidPlacedReader: bidPlacedReader,
		DB:              db,
		Batch:           batch,
	}
}

func (p *BidPlacedProjector) Run(ctx context.Context) error {
	p.Log.Info("projector starting; ensuring indexes")

	if err := p.EnsureIndexes(ctx); err != nil {
		return err
	}

	bids := p.DB.Collection("bids_history")
	auct := p.DB.Collection("auctions_view")
	p.Log.Info("indexes ready; entering consume loop")

	go func() {
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s := p.BidPlacedReader.Stats()
				p.Log.Info("kafka reader stats",
					zap.Int64("lag", s.Lag),
					zap.Int64("offset", s.Offset),
					zap.Int64("dials", s.Dials),
					zap.Int64("fetches", s.Fetches),
					zap.Int64("messages", s.Messages),
					zap.Int64("timeouts", s.Timeouts),
					zap.Int64("errors", s.Errors),
					zap.Any("partitions", s.Partition),
				)
			}
		}
	}()

	if err := p.waitForAssignment(ctx, 600*time.Second); err != nil {
		p.Log.Warn("kafka: no activity yet; continuing to read", zap.Error(err))
	} else {
		s := p.BidPlacedReader.Stats()
		p.Log.Info("kafka: reader active", zap.Int64("fetches", s.Fetches), zap.Int64("messages", s.Messages))
	}

	processed := 0
	for {
		m, err := p.BidPlacedReader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				p.Log.Info("context canceled")
				return nil // normal shutdown
			}
			p.Log.Error("FetchMessage failed", zap.Error(err))
			return err
		}

		p.Log.Info("bids.placed: received",
			zap.String("key", string(m.Key)),
			zap.Int("size", len(m.Value)),
			zap.Int("partition", m.Partition),
			zap.Int64("offset", m.Offset))

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
		// todo after moving to projector service, batch process
	}
}

func (p *BidPlacedProjector) waitForAssignment(ctx context.Context, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	for {
		s := p.BidPlacedReader.Stats()
		if s.Fetches > 0 || s.Messages > 0 {
			// no partition assigned yet
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("no group assignment within deadline")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
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

func EnsureTopic(ctx context.Context, brokers []string, topic string, numPartitions, replicationFactor int) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers provided")
	}
	if topic == "" {
		return fmt.Errorf("empty topic")
	}

	d := &kafka.Dialer{Timeout: 5 * time.Second, DualStack: true}

	conn, err := d.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("dial broker: %w", err)
	}
	defer conn.Close()

	// check if already exists
	parts, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("read partitions: %w", err)
	}
	for _, p := range parts {
		if p.Topic == topic {
			return nil // already exists
		}
	}

	// create topic if doesnt exist
	cfg := []kafka.TopicConfig{{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: replicationFactor,
	}}
	if err = conn.CreateTopics(cfg...); err != nil {
		return fmt.Errorf("create topics: %w", err)
	}
	return nil
}
