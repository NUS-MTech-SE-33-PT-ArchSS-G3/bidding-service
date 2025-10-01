package server

import (
	"kei-services/internal/bidding/cfg"
	httpPresentation "kei-services/internal/bidding/presentation/http"
	"kei-services/openapi"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func registerProtectedRoutes(r *gin.Engine, d *deps, cfg *cfg.Config, log *zap.Logger) {
	protected := r.Group("")
	protected.Use()

	m := &MasterHandler{
		PlaceBidHandler: *httpPresentation.NewPlaceBidController(log, d.PlaceBidService),
	}

	openapi.RegisterHandlers(protected, m)
}

type MasterHandler struct {
	PlaceBidHandler httpPresentation.PlaceBidController
}

func (m MasterHandler) PostAuctionsAuctionIdBids(c *gin.Context, auctionId string, params openapi.PostAuctionsAuctionIdBidsParams) {
	m.PlaceBidHandler.PostAuctionsAuctionIdBids(c, auctionId, params)
}
