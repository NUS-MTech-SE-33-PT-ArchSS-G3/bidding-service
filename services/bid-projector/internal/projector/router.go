package projector

import (
	"context"
	"errors"
	"kei-services/services/bid-projector/internal/events"

	"github.com/segmentio/kafka-go"
)

// Router maps topic + payload to handler
type Router struct {
	Codec    *events.Codec
	Handlers AuctionHandlers // todo after testing, accept array of handlers
}

func (r *Router) Route(msg kafka.Message) (evt any, handler func(context.Context, any) error, err error) {
	event, err := r.Codec.Decode(msg.Topic, msg.Value)
	if err != nil {
		return nil, nil, err
	}

	switch e := event.(type) {
	case events.BidPlaced:
		return e, func(ctx context.Context, v any) error {
			return r.Handlers.OnBidsPlaced(ctx, v.(events.BidPlaced))
		}, nil
	default:
		return nil, nil, errors.New("router: unsupported event type")
	}
}
