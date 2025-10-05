package domain

import "errors"

var (
	ErrInvalidBidderID   = errors.New("invalid_bidder_id")
	ErrAuctionClosed     = errors.New("auction_closed")
	ErrAuctionNotFound   = errors.New("auction_not_found")
	ErrBelowMinIncrement = errors.New("below_min_increment")
	ErrVersionConflict   = errors.New("version_conflict")
	ErrInvalidAmount     = errors.New("invalid_amount")
)
