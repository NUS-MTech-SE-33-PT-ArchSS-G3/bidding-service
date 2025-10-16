package kafka

import (
	"crypto/tls"
	"fmt"
	"time"

	segmentKafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type WriterConfig struct {
	Brokers     []string
	Topic       string // leave empty for multi-topic writer
	ClientID    string
	Acks        segmentKafka.RequiredAcks // eg kafka.RequireAll
	Compression segmentKafka.Compression  // eg kafka.Snappy
	SASLPlain   *struct{ Username, Password string }
	TLS         *tls.Config
	Balancer    segmentKafka.Balancer // eg &kafka.Hash{}, &kafka.LeastBytes{}
}

func NewWriter(cfg *WriterConfig, log *zap.Logger) *segmentKafka.Writer {
	w := &segmentKafka.Writer{
		Addr:         segmentKafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     balancerOrDefault(cfg.Balancer), // hash by Key
		RequiredAcks: firstAcks(cfg.Acks, segmentKafka.RequireAll),
		Compression:  firstCompression(cfg.Compression, segmentKafka.Snappy),
		BatchTimeout: 20 * time.Millisecond,
		BatchBytes:   64 << 10, // 64KB
		Async:        false,    // block and surface errors
	}
	w.AllowAutoTopicCreation = true

	// todo: add metrics hook

	w.Logger = segmentKafka.LoggerFunc(func(msg string, args ...interface{}) {
		log.Debug("kafka writer", zap.String("msg", fmt.Sprintf(msg, args...)))
	})
	w.ErrorLogger = segmentKafka.LoggerFunc(func(msg string, args ...interface{}) {
		log.Error("kafka writer error", zap.String("msg", fmt.Sprintf(msg, args...)))
	})

	// todo WriterTransport tls
	//w.Addr = kafka.TCP(cfg.Brokers...)

	return w
}

func balancerOrDefault(b segmentKafka.Balancer) segmentKafka.Balancer {
	if b != nil {
		return b
	}
	return &segmentKafka.Hash{}
}

func firstAcks(v, def segmentKafka.RequiredAcks) segmentKafka.RequiredAcks {
	if v == 0 {
		return def
	}
	return v
}

func firstCompression(v, def segmentKafka.Compression) segmentKafka.Compression {
	if v == 0 {
		return def
	}
	return v
}
