package http

import (
	"errors"
	"fmt"
	"kei-services/services/bid-command/internal/application/place_bid"
	"kei-services/services/bid-command/internal/domain"
	"kei-services/services/bid-command/openapi"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PlaceBidController struct {
	log *zap.Logger
	svc place_bid.IService
}

func NewPlaceBidController(log *zap.Logger, svc place_bid.IService) *PlaceBidController {
	return &PlaceBidController{log: log, svc: svc}
}

var _ openapi.ServerInterface = (*PlaceBidController)(nil)

// todo: add logging
func (h *PlaceBidController) PostAuctionsAuctionIdBids(c *gin.Context, auctionId string) {
	var req openapi.PlaceBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeProblem(c, http.StatusBadRequest,
			"https://example.com/problems/invalid-request",
			"Invalid request body",
			fmt.Sprintf("JSON decode/validation error: %v", err),
		)
		return
	}

	// todo: use bidning tags for validation
	if req.BidderId == "" {
		writeProblem(c, http.StatusBadRequest,
			"https://example.com/problems/invalid-request",
			"Invalid request body",
			"bidderId is required",
		)
		return
	}
	if req.Amount <= 0 {
		writeProblem(c, http.StatusBadRequest,
			"https://example.com/problems/invalid-request",
			"Invalid request body",
			"amount must be > 0",
		)
		return
	}

	// call application layer
	res, err := h.svc.Handle(c.Request.Context(), place_bid.Command{
		AuctionID: auctionId,
		BidderID:  req.BidderId,
		Amount:    req.Amount,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	// map to oapi schema
	out := openapi.PlaceBidResponse{
		BidId:        res.BidID,
		AuctionId:    res.AuctionID,
		BidderId:     res.BidderID,
		Accepted:     true,
		CurrentPrice: res.CurrentPrice,
		MinNextBid:   res.MinNextBid,
		At:           res.At,
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusCreated, out)
}

// todo: improve error handling and logging
func (h *PlaceBidController) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, place_bid.ErrUnauthorized):
		writeProblem(c, http.StatusUnauthorized,
			"https://example.com/problems/unauthorized",
			"Unauthorized",
			"Missing or invalid credentials",
		)

	// Domain/business
	case errors.Is(err, domain.ErrAuctionClosed):
		writeProblem(c, http.StatusUnprocessableEntity,
			"https://example.com/problems/auction-closed",
			"Auction closed",
			"No further bids are accepted for this auction",
		)
	case errors.Is(err, domain.ErrBelowMinIncrement):
		writeProblem(c, http.StatusUnprocessableEntity,
			"https://example.com/problems/below-min-increment",
			"Bid rejected: below minimum increment",
			err.Error(), // e.g., "next valid bid must be >= 102.5"
		)
	case errors.Is(err, domain.ErrAuctionNotFound):
		writeProblem(c, http.StatusUnprocessableEntity,
			"https://example.com/problems/auction-not-found",
			"Auction not found",
			"Cannot place bid because auction metadata is unavailable",
		)

	// conflicts
	case errors.Is(err, place_bid.ErrVersionConflict):
		writeProblem(c, http.StatusConflict,
			"https://example.com/problems/version-conflict",
			"Conflict",
			"Concurrent update detected; fetch latest price and retry",
		)

	// fallback
	default:
		h.log.Error("unhandled error in PostAuctionsAuctionIdBids", zap.Error(err))
		writeProblem(c, http.StatusInternalServerError,
			"https://example.com/problems/internal",
			"Internal Server Error",
			"An unexpected error occurred",
		)
	}
}
