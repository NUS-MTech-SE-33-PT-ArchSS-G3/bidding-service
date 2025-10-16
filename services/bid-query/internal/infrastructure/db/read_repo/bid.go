package read_repo

import (
	"context"
	"errors"
	"kei-services/pkg/middleware"
	"kei-services/services/bid-query/internal/application/list_bids"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var _ list_bids.IBidReadRepository = (*MongoBidReadRepo)(nil)

type MongoBidReadRepo struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func NewMongoBidReadRepo(db *mongo.Database, collection string, log *zap.Logger) *MongoBidReadRepo {
	return &MongoBidReadRepo{
		coll: db.Collection(collection),
		log:  log,
	}
}

type bidDoc struct {
	AuctionID string    `bson:"auctionId"`
	BidID     string    `bson:"bidId"`
	BidderID  string    `bson:"bidderId"`
	Amount    float64   `bson:"amount"`
	At        time.Time `bson:"at"`
}

// EnsureIndexes creates indexes. Called only once on service startup
func (r *MongoBidReadRepo) EnsureIndexes(ctx context.Context) error {
	if r.coll == nil {
		return errors.New("nil collection")
	}

	models := []mongo.IndexModel{
		// for pagination & listing auctionId + at desc + bidId desc
		{
			Keys:    bson.D{{Key: "auctionId", Value: 1}, {Key: "at", Value: -1}, {Key: "bidId", Value: -1}},
			Options: options.Index().SetName("auction_at_desc_bid_desc"),
		},
		// reverse order for ASC
		{
			Keys:    bson.D{{Key: "auctionId", Value: 1}, {Key: "at", Value: 1}, {Key: "bidId", Value: 1}},
			Options: options.Index().SetName("auction_at_asc_bid_asc"),
		},
	}
	_, err := r.coll.Indexes().CreateMany(ctx, models)
	return err
}

func (r *MongoBidReadRepo) ListByAuction(ctx context.Context, auctionID string, after *list_bids.Cursor, limit int,
	asc bool) (items []list_bids.Item, hasMore bool, next *list_bids.Cursor, err error) {
	log := middleware.LoggerFrom(ctx, r.log).With(zap.String("auctionId", auctionID))

	f := bson.M{"auctionId": auctionID}

	// pagination
	if after != nil {
		atCmp := "$lt"
		idCmp := "$lt"
		if asc {
			atCmp = "$gt"
			idCmp = "$gt"
		}
		f["$or"] = bson.A{
			bson.M{"at": bson.M{atCmp: after.At}},
			bson.M{"at": after.At, "bidId": bson.M{idCmp: after.ID}},
		}
	}

	// sort
	sort := bson.D{{Key: "at", Value: -1}, {Key: "bidId", Value: -1}}
	if asc {
		sort = bson.D{{Key: "at", Value: 1}, {Key: "bidId", Value: 1}}
	}

	findOpts := options.Find().
		SetSort(sort).
		SetLimit(int64(limit + 1)) // fetch one extra to detect hasMore

	log.Debug("listing bids", zap.Any("filter", f), zap.Any("findOpts", findOpts))
	cur, err := r.coll.Find(ctx, f, findOpts)
	if err != nil {
		log.Warn("failed to list items", zap.Error(err))
		return nil, false, nil, err
	}
	defer cur.Close(ctx)

	var docs []bidDoc
	if err := cur.All(ctx, &docs); err != nil {
		log.Warn("failed to decode items", zap.Error(err))
		return nil, false, nil, err
	}

	hasMore = len(docs) > limit
	if hasMore {
		docs = docs[:limit]
	}

	items = make([]list_bids.Item, 0, len(docs))
	for _, d := range docs {
		items = append(items, list_bids.Item{
			BidID:     d.BidID,
			AuctionID: d.AuctionID,
			BidderID:  d.BidderID,
			Amount:    d.Amount,
			At:        toTime(d.At),
		})
	}

	if len(docs) > 0 {
		last := docs[len(docs)-1]
		next = &list_bids.Cursor{At: last.At, ID: last.BidID}
	}

	return items, hasMore, next, nil
}

// convert driver time to time.Time
func toTime(v any) (t time.Time) {
	switch x := v.(type) {
	case time.Time:
		return x
	case *time.Time:
		if x != nil {
			return *x
		}
	}
	return time.Time{}
}
