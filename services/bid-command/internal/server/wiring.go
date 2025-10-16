package server

import (
	"kei-services/services/bid-command/internal/application/place_bid"
	"kei-services/services/bid-command/internal/cfg"
	"kei-services/services/bid-command/internal/infrastructure/cache"
	"kei-services/services/bid-command/internal/infrastructure/db/repo"
	"kei-services/services/bid-command/internal/infrastructure/db/tx"
	"kei-services/services/bid-command/internal/infrastructure/mq"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type deps struct {
	PlaceBidService place_bid.IService
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func initDependencies(db *gorm.DB, redis *redis.Client, w *kafka.Writer, cfg *cfg.Config, log *zap.Logger) *deps {
	sqlDb, err := db.DB()
	if err != nil {
		log.Fatal("failed to get sql db from gorm", zap.Error(err))
	}

	placeBidService := place_bid.NewService(place_bid.Deps{
		BidRepo: repo.NewBidRepo(sqlDb, log),
		Cache:   cache.NewAuctionMetadataCache(redis, log),
		Pub:     mq.NewBidsPublisher(w, log),
		Tx:      tx.NewTxManager(sqlDb),
		Clock:   systemClock{},
	}, log)

	return &deps{
		PlaceBidService: placeBidService,
	}
}
