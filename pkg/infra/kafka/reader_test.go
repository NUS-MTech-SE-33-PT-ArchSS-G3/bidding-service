package kafka

import (
	"testing"
	"time"

	segmentKafka "github.com/segmentio/kafka-go"
)

func TestNewReader_NilConfig(t *testing.T) {
	_, err := NewReader(nil)

	if err == nil {
		t.Error("expected error for nil config")
	}

	if err.Error() != "nil reader config" {
		t.Errorf("expected 'nil reader config' error, got: %v", err)
	}
}

func TestNewReader_MissingGroupID(t *testing.T) {
	cfg := &ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
	}

	_, err := NewReader(cfg)

	if err == nil {
		t.Error("expected error for missing GroupID")
	}

	if err.Error() != "GroupID is required for group readers" {
		t.Errorf("expected GroupID error, got: %v", err)
	}
}

func TestNewReader_NoTopics(t *testing.T) {
	cfg := &ReaderConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "test-group",
	}

	_, err := NewReader(cfg)

	if err == nil {
		t.Error("expected error when neither Topic nor GroupTopics is set")
	}

	if err.Error() != "must set either Topic or GroupTopics" {
		t.Errorf("expected topic error, got: %v", err)
	}
}

func TestNewReader_BothTopics(t *testing.T) {
	cfg := &ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		GroupID:     "test-group",
		Topic:       "test-topic",
		GroupTopics: []string{"topic1", "topic2"},
	}

	_, err := NewReader(cfg)

	if err == nil {
		t.Error("expected error when both Topic and GroupTopics are set")
	}

	if err.Error() != "cannot set both Topic and GroupTopics" {
		t.Errorf("expected both topics error, got: %v", err)
	}
}

func TestNewReader_SingleTopic(t *testing.T) {
	cfg := &ReaderConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "test-group",
		Topic:   "test-topic",
	}

	reader, err := NewReader(cfg)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if reader == nil {
		t.Fatal("expected non-nil reader")
	}

	// Clean up
	reader.Close()
}

func TestNewReader_MultipleTopics(t *testing.T) {
	cfg := &ReaderConfig{
		Brokers:     []string{"localhost:9092"},
		GroupID:     "test-group",
		GroupTopics: []string{"topic1", "topic2", "topic3"},
	}

	reader, err := NewReader(cfg)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if reader == nil {
		t.Fatal("expected non-nil reader")
	}

	// Clean up
	reader.Close()
}

func TestNewReader_WithTuning(t *testing.T) {
	cfg := &ReaderConfig{
		Brokers:           []string{"localhost:9092"},
		GroupID:           "test-group",
		Topic:             "test-topic",
		MinBytes:          100,
		MaxBytes:          1024 * 1024,
		MaxWait:           500 * time.Millisecond,
		CommitInterval:    2 * time.Second,
		SessionTimeout:    20 * time.Second,
		HeartbeatInterval: 5 * time.Second,
		RebalanceTimeout:  10 * time.Second,
		ReadLagInterval:   30 * time.Second,
		Offset:            OffsetFirst,
	}

	reader, err := NewReader(cfg)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if reader == nil {
		t.Fatal("expected non-nil reader")
	}

	// Clean up
	reader.Close()
}

func TestNewReader_WithOffsetLast(t *testing.T) {
	cfg := &ReaderConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "test-group",
		Topic:   "test-topic",
		Offset:  OffsetLast,
	}

	reader, err := NewReader(cfg)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if reader == nil {
		t.Fatal("expected non-nil reader")
	}

	// Clean up
	reader.Close()
}

func TestDefaultInt(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		defValue int
		expected int
	}{
		{
			name:     "zero value uses default",
			value:    0,
			defValue: 100,
			expected: 100,
		},
		{
			name:     "non-zero value is used",
			value:    50,
			defValue: 100,
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultInt(tt.value, tt.defValue)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDefaultDur(t *testing.T) {
	tests := []struct {
		name     string
		value    time.Duration
		defValue time.Duration
		expected time.Duration
	}{
		{
			name:     "zero duration uses default",
			value:    0,
			defValue: 5 * time.Second,
			expected: 5 * time.Second,
		},
		{
			name:     "non-zero duration is used",
			value:    10 * time.Second,
			defValue: 5 * time.Second,
			expected: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := defaultDur(tt.value, tt.defValue)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToStartOffset(t *testing.T) {
	tests := []struct {
		name     string
		mode     OffsetMode
		expected int64
	}{
		{
			name:     "OffsetFirst returns FirstOffset",
			mode:     OffsetFirst,
			expected: segmentKafka.FirstOffset,
		},
		{
			name:     "OffsetLast returns LastOffset",
			mode:     OffsetLast,
			expected: segmentKafka.LastOffset,
		},
		{
			name:     "empty mode defaults to LastOffset",
			mode:     "",
			expected: segmentKafka.LastOffset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toStartOffset(tt.mode)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestOffsetModeConstants(t *testing.T) {
	if OffsetLast != "last" {
		t.Errorf("expected OffsetLast to be 'last', got %s", OffsetLast)
	}

	if OffsetFirst != "first" {
		t.Errorf("expected OffsetFirst to be 'first', got %s", OffsetFirst)
	}
}
