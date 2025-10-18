package server

import (
	"context"
	"kei-services/services/bid-query/internal/cfg"
	httpPresentation "kei-services/services/bid-query/internal/presentation/http"
	"kei-services/services/bid-query/openapi"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func registerProtectedRoutes(r *gin.Engine, d *deps, _ *cfg.Config, log *zap.Logger) {
	protected := r.Group("")
	protected.Use()

	m := &MasterHandler{
		ListBidsHandler: *httpPresentation.NewHttpController(log, d.ListBidsService),
	}

	openapi.RegisterHandlers(protected, m)
}

type MasterHandler struct {
	ListBidsHandler httpPresentation.HttpController
}

func (m MasterHandler) GetApiV1BidsAuctionId(c *gin.Context, auctionId string, params openapi.GetApiV1BidsAuctionIdParams) {
	m.ListBidsHandler.GetApiV1BidsAuctionId(c, auctionId, params)
}

func registerHealthroutes(r *gin.Engine, db *mongo.Database, redis *redis.Client, _ *zap.Logger) {
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		type check struct {
			Name string `json:"name"`
			OK   bool   `json:"ok"`
			Err  string `json:"err,omitempty"`
		}

		var checks []check
		allOK := true

		// DB ping
		if err := db.Client().Ping(ctx, nil); err != nil {
			allOK = false
			checks = append(checks, check{Name: "db", OK: false, Err: err.Error()})
		} else {
			checks = append(checks, check{Name: "db", OK: true})
		}

		// Redis ping
		if err := redis.Ping(ctx).Err(); err != nil {
			allOK = false
			checks = append(checks, check{Name: "redis", OK: false, Err: err.Error()})
		} else {
			checks = append(checks, check{Name: "redis", OK: true})
		}

		status := http.StatusOK
		if !allOK {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{
			"status": map[bool]string{true: "ready", false: "degraded"}[allOK],
			"checks": checks,
			"uptime": time.Since(time.Unix(0, 0)).String(), // todo use start time
		})
	})
}
