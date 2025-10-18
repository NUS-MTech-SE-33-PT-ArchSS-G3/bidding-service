package mongo

import "time"

type BidDoc struct {
	AuctionID string    `bson:"auctionId"`
	BidID     string    `bson:"bidId"`
	BidderID  string    `bson:"bidderId"`
	Amount    float64   `bson:"amount"`
	At        time.Time `bson:"at"`
}
