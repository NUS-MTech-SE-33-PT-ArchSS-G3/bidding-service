package domain

import (
	"fmt"
)

// LastAcceptedBid is the authoritative state we validate against inside the tx.
type LastAcceptedBid struct {
	Price   float64
	Version int
}

// ValidateBid enforces the bid amount against the auction metadata
// pass nil for LastAcceptedBid if no prior bid exists or for quick checks
func ValidateBid(auction *AuctionMetadata, amount float64, b *LastAcceptedBid) error {
	if auction == nil {
		return ErrAuctionNotFound
	}
	if !auction.IsOpen() {
		return ErrAuctionClosed
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}

	var min float64
	if b == nil {
		// no prior bid, so use auction metadata only
		min = auction.MinNextBid()
	} else {
		min = MinNextPrice(*b, auction)
	}
	// allow tiny float slack
	if SafeLess(amount, min, 1e-9) {
		return fmt.Errorf("%w: next valid bid must be >= %.2f", ErrBelowMinIncrement, min)
	}

	return nil
}

// MakeLastAcceptedBid merges cached auction metadata with DB's latest bid amount
func MakeLastAcceptedBid(auction *AuctionMetadata, latestAmount *float64, latestSeq *int64) LastAcceptedBid {
	currPrice := auction.CurrentPrice
	ver := auction.Version

	if latestAmount != nil && *latestAmount > currPrice {
		currPrice = *latestAmount
	}
	if latestSeq != nil && int(*latestSeq) > ver {
		ver = int(*latestSeq)
	}
	return LastAcceptedBid{Price: currPrice, Version: ver}
}

// MinNextPrice computes the min acceptable amount given a last accepted bid + auction min increment policy
func MinNextPrice(bid LastAcceptedBid, auction *AuctionMetadata) float64 {
	if bid.Price <= 0 {
		return auction.StartingPrice
	}
	return bid.Price + auction.MinIncrement
}

// ApplyAccepted updates a copy of AuctionMetadata after accepting a bid
func ApplyAccepted(auction AuctionMetadata, bid *Bid) AuctionMetadata {
	auction.CurrentPrice = bid.Amount
	auction.Version += 1
	return auction
}
