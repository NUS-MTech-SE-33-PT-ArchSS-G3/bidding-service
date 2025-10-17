package domain

import "errors"

var (
	ErrAuctionClosed     = errors.New("auction_closed")
	ErrAuctionNotFound   = errors.New("auction_not_found")
	ErrBelowMinIncrement = errors.New("below_min_increment")
	ErrInvalidAmount     = errors.New("invalid_amount")
)
