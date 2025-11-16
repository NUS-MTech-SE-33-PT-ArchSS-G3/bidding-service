package kafka

import (
	"testing"

	segmentKafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func TestNewWriter_Basic(t *testing.T) {
	cfg := &WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "test-topic",
		ClientID: "test-client",
	}

	logger := zap.NewNop()
	writer := NewWriter(cfg, logger)

	if writer == nil {
		t.Fatal("expected non-nil writer")
	}

	// Clean up
	writer.Close()
}

func TestNewWriter_MultiTopic(t *testing.T) {
	cfg := &WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "", // Empty for multi-topic writer
		ClientID: "test-client",
	}

	logger := zap.NewNop()
	writer := NewWriter(cfg, logger)

	if writer == nil {
		t.Fatal("expected non-nil writer")
	}

	if writer.Topic != "" {
		t.Error("expected empty topic for multi-topic writer")
	}

	// Clean up
	writer.Close()
}

func TestNewWriter_WithAcks(t *testing.T) {
	cfg := &WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "test-topic",
		ClientID: "test-client",
		Acks:     segmentKafka.RequireOne,
	}

	logger := zap.NewNop()
	writer := NewWriter(cfg, logger)

	if writer == nil {
		t.Fatal("expected non-nil writer")
	}

	if writer.RequiredAcks != segmentKafka.RequireOne {
		t.Errorf("expected RequiredAcks to be RequireOne, got %v", writer.RequiredAcks)
	}

	// Clean up
	writer.Close()
}

func TestNewWriter_WithCompression(t *testing.T) {
	cfg := &WriterConfig{
		Brokers:     []string{"localhost:9092"},
		Topic:       "test-topic",
		ClientID:    "test-client",
		Compression: segmentKafka.Gzip,
	}

	logger := zap.NewNop()
	writer := NewWriter(cfg, logger)

	if writer == nil {
		t.Fatal("expected non-nil writer")
	}

	if writer.Compression != segmentKafka.Gzip {
		t.Errorf("expected Compression to be Gzip, got %v", writer.Compression)
	}

	// Clean up
	writer.Close()
}

func TestNewWriter_WithBalancer(t *testing.T) {
	cfg := &WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "test-topic",
		ClientID: "test-client",
		Balancer: &segmentKafka.LeastBytes{},
	}

	logger := zap.NewNop()
	writer := NewWriter(cfg, logger)

	if writer == nil {
		t.Fatal("expected non-nil writer")
	}

	// Clean up
	writer.Close()
}

func TestNewWriter_DefaultValues(t *testing.T) {
	cfg := &WriterConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "test-topic",
		ClientID: "test-client",
	}

	logger := zap.NewNop()
	writer := NewWriter(cfg, logger)

	if writer == nil {
		t.Fatal("expected non-nil writer")
	}

	// Check default values
	if writer.RequiredAcks != segmentKafka.RequireAll {
		t.Errorf("expected default RequiredAcks to be RequireAll, got %v", writer.RequiredAcks)
	}

	if writer.Compression != segmentKafka.Snappy {
		t.Errorf("expected default Compression to be Snappy, got %v", writer.Compression)
	}

	if writer.Async != false {
		t.Error("expected Async to be false")
	}

	if writer.AllowAutoTopicCreation != true {
		t.Error("expected AllowAutoTopicCreation to be true")
	}

	// Clean up
	writer.Close()
}

func TestBalancerOrDefault_Nil(t *testing.T) {
	result := balancerOrDefault(nil)

	if result == nil {
		t.Error("expected non-nil balancer")
	}

	// Should be Hash balancer by default
	if _, ok := result.(*segmentKafka.Hash); !ok {
		t.Error("expected default balancer to be Hash")
	}
}

func TestBalancerOrDefault_NotNil(t *testing.T) {
	customBalancer := &segmentKafka.LeastBytes{}
	result := balancerOrDefault(customBalancer)

	if result != customBalancer {
		t.Error("expected custom balancer to be returned")
	}
}

func TestFirstAcks_Zero(t *testing.T) {
	result := firstAcks(0, segmentKafka.RequireAll)

	if result != segmentKafka.RequireAll {
		t.Errorf("expected default RequireAll, got %v", result)
	}
}

func TestFirstAcks_NonZero(t *testing.T) {
	result := firstAcks(segmentKafka.RequireOne, segmentKafka.RequireAll)

	if result != segmentKafka.RequireOne {
		t.Errorf("expected RequireOne, got %v", result)
	}
}

func TestFirstCompression_Zero(t *testing.T) {
	result := firstCompression(0, segmentKafka.Snappy)

	if result != segmentKafka.Snappy {
		t.Errorf("expected default Snappy, got %v", result)
	}
}

func TestFirstCompression_NonZero(t *testing.T) {
	result := firstCompression(segmentKafka.Gzip, segmentKafka.Snappy)

	if result != segmentKafka.Gzip {
		t.Errorf("expected Gzip, got %v", result)
	}
}
