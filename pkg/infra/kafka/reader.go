package kafka

import (
	"crypto/tls"
	"errors"
	"time"

	segmentKafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// OffsetMode controls where a new group starts reading.
type OffsetMode string

const (
	OffsetLast  OffsetMode = "last"  // default start at end
	OffsetFirst OffsetMode = "first" // start from earliest
)

type ReaderConfig struct {
	Brokers     []string
	Topic       string
	GroupTopics []string

	GroupID string

	// optional tuning
	MinBytes          int           // default: 1
	MaxBytes          int           // default: 10 << 20
	MaxWait           time.Duration // default: 250ms
	CommitInterval    time.Duration // default: 1s (0 means sync each message)
	SessionTimeout    time.Duration // default: 10s
	HeartbeatInterval time.Duration // default: 3s
	RebalanceTimeout  time.Duration // default: 0 (use kafka-go default)
	ReadLagInterval   time.Duration // default: -1 (disabled)

	Offset OffsetMode // default OffsetLast

	TLS       *tls.Config
	SASLPlain *struct {
		Username string
		Password string
	}
}

func NewReader(cfg *ReaderConfig) (*segmentKafka.Reader, error) {
	if cfg == nil {
		return nil, errors.New("nil reader config")
	}
	if cfg.GroupID == "" {
		return nil, errors.New("GroupID is required for group readers")
	}

	// validate single-topic vs multi-topic
	single := cfg.Topic != ""
	multi := len(cfg.GroupTopics) > 0
	switch {
	case !single && !multi:
		return nil, errors.New("must set either Topic or GroupTopics")
	case single && multi:
		return nil, errors.New("cannot set both Topic and GroupTopics")
	}

	rc := segmentKafka.ReaderConfig{
		Brokers:           cfg.Brokers,
		GroupID:           cfg.GroupID,
		MinBytes:          defaultInt(cfg.MinBytes, 1),
		MaxBytes:          defaultInt(cfg.MaxBytes, 10<<20),
		MaxWait:           defaultDur(cfg.MaxWait, 250*time.Millisecond),
		CommitInterval:    defaultDur(cfg.CommitInterval, time.Second),
		SessionTimeout:    defaultDur(cfg.SessionTimeout, 10*time.Second),
		HeartbeatInterval: defaultDur(cfg.HeartbeatInterval, 3*time.Second),
		ReadLagInterval:   defaultDur(cfg.ReadLagInterval, -1), // disabled
		StartOffset:       toStartOffset(cfg.Offset),
	}

	if single {
		rc.Topic = cfg.Topic
	} else {
		rc.GroupTopics = append([]string(nil), cfg.GroupTopics...)
	}

	if cfg.TLS != nil || cfg.SASLPlain != nil {
		dialer := &segmentKafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS:       cfg.TLS,
		}
		if cfg.SASLPlain != nil {
			dialer.SASLMechanism = plain.Mechanism{
				Username: cfg.SASLPlain.Username,
				Password: cfg.SASLPlain.Password,
			}
		}
		rc.Dialer = dialer
	}

	r := segmentKafka.NewReader(rc)

	// todo: add metrics hook
	// todo: figure out how to log from reader

	return r, nil
}

func defaultInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}

func defaultDur(v, def time.Duration) time.Duration {
	if v == 0 {
		return def
	}
	return v
}

func toStartOffset(m OffsetMode) int64 {
	switch m {
	case OffsetFirst:
		return segmentKafka.FirstOffset
	case OffsetLast:
		fallthrough
	default:
		return segmentKafka.LastOffset
	}
}
