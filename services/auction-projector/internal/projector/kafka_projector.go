package projector

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Projector struct {
	reader *kafka.Reader
	router *Router
	log    *zap.Logger
}

func New(reader *kafka.Reader, router *Router, log *zap.Logger) *Projector {
	return &Projector{
		reader: reader,
		router: router,
		log:    log,
	}
}

func (p *Projector) Run(ctx context.Context) error {
	p.log.Info("projector starting")

	go func() {
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s := p.reader.Stats()
				p.log.Info("kafka reader stats",
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
		return fmt.Errorf("no partition assignment, %w", err)
	} else {
		p.log.Info("partitions assigned")
	}

	for {
		msg, err := p.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				p.log.Info("context canceled")
				return nil // normal shutdown
			}
			p.log.Error("ReadMessage", zap.Error(err))
			return err
		}

		evt, handler, err := p.router.Route(msg)
		if err != nil {
			p.log.Warn("router route", zap.Error(err), zap.String("topic", msg.Topic))
			continue
		}

		if err = handler(ctx, evt); err != nil {
			p.log.Error("handler", zap.Error(err), zap.String("topic", msg.Topic))
			// continue
		}
	}
}

func (p *Projector) waitForAssignment(ctx context.Context, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	for {
		s := p.reader.Stats()
		if s.Fetches > 0 || s.Messages > 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("still waiting for group assignment")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func EnsureTopics(ctx context.Context, brokers []string, topics []string, numPartitions, replicationFactor int) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers provided")
	}
	d := &kafka.Dialer{Timeout: 5 * time.Second, DualStack: true}
	conn, err := d.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("dial broker: %w", err)
	}
	defer conn.Close()

	existing := map[string]bool{}
	// check if already exists
	parts, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("read partitions: %w", err)
	}
	for _, p := range parts {
		existing[p.Topic] = true
	}

	// create topics if doesnt exist
	var toCreate []kafka.TopicConfig
	for _, t := range topics {
		if t == "" {
			continue
		}
		if existing[t] {
			continue
		}
		toCreate = append(toCreate, kafka.TopicConfig{
			Topic:             t,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		})
	}

	if len(toCreate) == 0 {
		return nil
	}
	if err = conn.CreateTopics(toCreate...); err != nil {
		return fmt.Errorf("create topics: %w", err)
	}
	return nil
}
