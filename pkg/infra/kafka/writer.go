package kafka

// https://github.com/segmentio/kafka-go
import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type WriterConfig struct {
	Brokers     []string
	Topic       string
	ClientID    string
	Acks        kafka.RequiredAcks // eg kafka.RequireAll
	Compression kafka.Compression  // eg kafka.Snappy
	SASLPlain   *struct{ Username, Password string }
	TLS         *tls.Config
	Balancer    kafka.Balancer // eg &kafka.Hash{}, &kafka.LeastBytes{}
}

func NewWriter(cfg WriterConfig, log *zap.Logger) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     balancerOrDefault(cfg.Balancer), // hash by Key
		RequiredAcks: firstAcks(cfg.Acks, kafka.RequireAll),
		Compression:  firstCompression(cfg.Compression, kafka.Snappy),
		BatchTimeout: 20 * time.Millisecond,
		BatchBytes:   64 << 10, // 64KB
		Async:        true,
	}
	w.AllowAutoTopicCreation = true

	// todo: add metrics hook

	w.Logger = kafka.LoggerFunc(func(msg string, args ...interface{}) {
		log.Debug("kafka writer", zap.String("msg", fmt.Sprintf(msg, args...)))
	})
	w.ErrorLogger = kafka.LoggerFunc(func(msg string, args ...interface{}) {
		log.Error("kafka writer error", zap.String("msg", fmt.Sprintf(msg, args...)))
	})

	// todo WriterTransport tls
	//w.Addr = kafka.TCP(cfg.Brokers...)

	return w
}

func balancerOrDefault(b kafka.Balancer) kafka.Balancer {
	if b != nil {
		return b
	}
	return &kafka.Hash{}
}

func firstAcks(v, def kafka.RequiredAcks) kafka.RequiredAcks {
	if v == 0 {
		return def
	}
	return v
}

func firstCompression(v, def kafka.Compression) kafka.Compression {
	if v == 0 {
		return def
	}
	return v
}
