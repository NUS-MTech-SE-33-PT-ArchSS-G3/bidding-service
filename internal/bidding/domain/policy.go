package domain

import (
	"fmt"
)

// ValidateBid enforces the bid amount against the auction metadata
func ValidateBid(auction *AuctionMetadata, amount float64) error {
	if auction == nil {
		return ErrAuctionMetaNotFound
	}
	if !auction.IsOpen() {
		return ErrAuctionClosed
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}

	min := auction.MinNextBid()

	// allow tiny float slack
	if SafeLess(amount, min, 1e-9) {
		return fmt.Errorf("%w: next valid bid must be >= %.2f", ErrBelowMinIncrement, min)
	}

	return nil
}

// ApplyAccepted updates a copy of AuctionMetadata after accepting a bid
func ApplyAccepted(meta AuctionMetadata, bid *Bid) AuctionMetadata {
	meta.CurrentPrice = bid.Amount
	meta.Version += 1
	return meta
}
