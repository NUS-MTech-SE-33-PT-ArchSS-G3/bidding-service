package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCodec_Decode_BidPlaced(t *testing.T) {
	codec := &Codec{}
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("valid bid placed event", func(t *testing.T) {
		evt := BidPlaced{
			AuctionID: "auction-1",
			BidID:     "bid-1",
			BidderID:  "bidder-1",
			Amount:    120.5,
			At:        fixedTime,
		}

		payload, err := json.Marshal(evt)
		assert.NoError(t, err)

		decoded, err := codec.Decode("bids.placed", payload)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)

		decodedEvt, ok := decoded.(BidPlaced)
		assert.True(t, ok)
		assert.Equal(t, evt.AuctionID, decodedEvt.AuctionID)
		assert.Equal(t, evt.BidID, decodedEvt.BidID)
		assert.Equal(t, evt.BidderID, decodedEvt.BidderID)
		assert.Equal(t, evt.Amount, decodedEvt.Amount)
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		payload := []byte(`{"invalid json`)

		_, err := codec.Decode("bids.placed", payload)
		assert.Error(t, err)
	})

	t.Run("unknown topic returns error", func(t *testing.T) {
		payload := []byte(`{}`)

		_, err := codec.Decode("unknown.topic", payload)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown topic")
	})
}
