package list_bids

import (
	"context"
)

type IBidReadRepository interface {
	ListByAuction(ctx context.Context, auctionID string, after *Cursor, limit int, asc bool) (
		items []Item, hasMore bool, next *Cursor, err error)
}
