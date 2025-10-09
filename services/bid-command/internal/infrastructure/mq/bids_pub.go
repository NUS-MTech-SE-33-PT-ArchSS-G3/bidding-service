package mq

import (
	"context"
	"encoding/json"
	"kei-services/services/bid-command/internal/domain"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var _ domain.IBidsPlacedPublisher = (*BidsPublisher)(nil)

type BidsPublisher struct {
	Writer *kafka.Writer
	Topic  string
	Log    *zap.Logger
}

func NewBidsPublisher(w *kafka.Writer, topic string, log *zap.Logger) BidsPublisher {
	return BidsPublisher{
		Writer: w,
		Topic:  topic,
		Log:    log,
	}
}

func (p BidsPublisher) Publish(ctx context.Context, evt domain.BidPlaced) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(evt.AuctionID),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "content-type", Value: []byte("application/json")},
			{Key: "schema", Value: []byte("bids.placed")},
			{Key: "schema-version", Value: []byte("1")},
		},
		Topic: p.Topic,
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return p.Writer.WriteMessages(ctx, msg)
}
