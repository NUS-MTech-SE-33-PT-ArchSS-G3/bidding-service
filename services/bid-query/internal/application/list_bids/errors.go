package list_bids

import "errors"

var (
	ErrInvalidCursor   = errors.New("invalid_cursor")
	ErrAuctionNotFound = errors.New("auction_not_found")
)
