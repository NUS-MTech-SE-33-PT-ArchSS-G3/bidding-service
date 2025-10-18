package read_model

import "time"

type Bids struct {
	AuctionID string    `bson:"auctionId"`
	BidID     string    `bson:"bidId"`
	BidderID  string    `bson:"bidderId"`
	Amount    float64   `bson:"amount"`
	At        time.Time `bson:"at"`
}
