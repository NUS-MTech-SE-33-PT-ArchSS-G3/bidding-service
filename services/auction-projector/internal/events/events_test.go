package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCodec_Decode_AuctionOpened(t *testing.T) {
	codec := &Codec{}
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("valid auction opened event", func(t *testing.T) {
		evt := AuctionOpened{
			AuctionID:     "auction-1",
			EndsAt:        fixedTime,
			StartingPrice: 100.0,
			MinIncrement:  10.0,
			Version:       1,
		}

		payload, err := json.Marshal(evt)
		assert.NoError(t, err)

		decoded, err := codec.Decode("auction.opened", payload)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)

		decodedEvt, ok := decoded.(AuctionOpened)
		assert.True(t, ok)
		assert.Equal(t, evt.AuctionID, decodedEvt.AuctionID)
		assert.Equal(t, evt.StartingPrice, decodedEvt.StartingPrice)
		assert.Equal(t, evt.MinIncrement, decodedEvt.MinIncrement)
		assert.Equal(t, evt.Version, decodedEvt.Version)
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		payload := []byte(`{"invalid json`)

		_, err := codec.Decode("auction.opened", payload)
		assert.Error(t, err)
	})
}

func TestCodec_Decode_AuctionClosed(t *testing.T) {
	codec := &Codec{}
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("valid auction closed event", func(t *testing.T) {
		evt := AuctionClosed{
			AuctionID: "auction-1",
			ClosedAt:  fixedTime,
			Version:   2,
		}

		payload, err := json.Marshal(evt)
		assert.NoError(t, err)

		decoded, err := codec.Decode("auction.closed", payload)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)

		decodedEvt, ok := decoded.(AuctionClosed)
		assert.True(t, ok)
		assert.Equal(t, evt.AuctionID, decodedEvt.AuctionID)
		assert.Equal(t, evt.Version, decodedEvt.Version)
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		payload := []byte(`{"invalid json`)

		_, err := codec.Decode("auction.closed", payload)
		assert.Error(t, err)
	})
}

func TestCodec_Decode_UnknownTopic(t *testing.T) {
	codec := &Codec{}

	t.Run("unknown topic returns error", func(t *testing.T) {
		payload := []byte(`{}`)

		_, err := codec.Decode("unknown.topic", payload)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown topic")
	})
}
