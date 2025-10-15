package server

import (
	"kei-services/services/bid-query/internal/application/list_bids"
	"kei-services/services/bid-query/internal/cfg"
	"kei-services/services/bid-query/internal/infrastructure/db/read_repo"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type deps struct {
	ListBidsService list_bids.IService
}

func initDependencies(db *mongo.Database, _ *redis.Client, _ *cfg.Config, log *zap.Logger) *deps {

	listBidService := list_bids.NewService(list_bids.Deps{
		BidReadRepo: read_repo.NewMongoBidReadRepo(db, "bids_history", log)},
		log,
	)

	return &deps{
		ListBidsService: listBidService,
	}
}
