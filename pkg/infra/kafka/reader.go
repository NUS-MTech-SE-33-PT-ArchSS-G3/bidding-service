package kafka

import (
	"crypto/tls"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// OffsetMode controls where a new group starts reading.
type OffsetMode string

const (
	OffsetLast  OffsetMode = "last"  // default: start at the end
	OffsetFirst OffsetMode = "first" // start from earliest
)

type ReaderConfig struct {
	Brokers []string
	Topic   string
	GroupID string

	// Optional tuning. Sensible defaults applied if zero.
	MinBytes          int           // default: 1
	MaxBytes          int           // default: 10 << 20
	MaxWait           time.Duration // default: 250ms
	CommitInterval    time.Duration // default: 1s (0 means sync each message)
	SessionTimeout    time.Duration // default: 10s
	HeartbeatInterval time.Duration // default: 3s
	RebalanceTimeout  time.Duration // default: 0 (use kafka-go default)
	ReadLagInterval   time.Duration // default: -1 (disabled)

	// Where to start for a *new* consumer group (existing offsets are respected)
	Offset OffsetMode // default: OffsetLast

	// Security (optional)
	TLS       *tls.Config
	SASLPlain *struct {
		Username string
		Password string
	}
}

func NewReader(cfg ReaderConfig) *kafka.Reader {
	rc := kafka.ReaderConfig{
		Brokers:           cfg.Brokers,
		GroupID:           cfg.GroupID,
		Topic:             cfg.Topic,
		MinBytes:          defaultInt(cfg.MinBytes, 1),
		MaxBytes:          defaultInt(cfg.MaxBytes, 10<<20),
		MaxWait:           defaultDur(cfg.MaxWait, 250*time.Millisecond),
		CommitInterval:    defaultDur(cfg.CommitInterval, time.Second),
		SessionTimeout:    defaultDur(cfg.SessionTimeout, 10*time.Second),
		HeartbeatInterval: defaultDur(cfg.HeartbeatInterval, 3*time.Second),
		ReadLagInterval:   defaultDur(cfg.ReadLagInterval, -1), // disabled
		StartOffset:       toStartOffset(cfg.Offset),
	}

	// Build a Dialer if TLS/SASL is configured.
	if cfg.TLS != nil || cfg.SASLPlain != nil {
		dialer := &kafka.Dialer{
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

	return kafka.NewReader(rc)
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
		return kafka.FirstOffset
	case OffsetLast:
		fallthrough
	default:
		return kafka.LastOffset
	}
}
