package http

import (
	"errors"
	"kei-services/pkg/middleware"
	"kei-services/services/bid-query/internal/application/list_bids"
	"kei-services/services/bid-query/openapi"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *HttpController) GetApiV1BidsAuctionId(c *gin.Context, auctionId string,
	params openapi.GetApiV1BidsAuctionIdParams) {
	log := middleware.LoggerFrom(c.Request.Context(), h.log)

	log.Info("list bids: request received", zap.String("auctionId", auctionId), zap.Any("params", params))

	// defaults & bounds
	limit := 50
	if params.Limit != nil {
		limit = *params.Limit
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 200 {
		limit = 200
	}

	dirStr := ""
	if params.Direction != nil {
		dirStr = string(*params.Direction)
	}
	direction := list_bids.DirectionDesc
	if strings.EqualFold(dirStr, "asc") {
		direction = list_bids.DirectionAsc
	}

	cursor := strDeref(params.Cursor)

	log.Debug("list bids: resolved query arguments",
		zap.String("auctionId", auctionId),
		zap.Int("limit", limit),
		zap.String("direction", direction.String()),
		zap.String("cursor", cursor),
	)

	q := list_bids.Query{
		AuctionID: auctionId,
		Cursor:    cursor,
		Limit:     limit,
		Direction: direction,
	}

	res, err := h.svc.Handle(c.Request.Context(), q)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// map to openapi
	items := make([]openapi.Bid, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, openapi.Bid{
			BidId:     it.BidID,
			AuctionId: it.AuctionID,
			BidderId:  it.BidderID,
			Amount:    it.Amount,
			At:        it.At,
		})
	}

	var first, last string
	if len(res.Items) > 0 {
		first = res.Items[0].BidID
		last = res.Items[len(res.Items)-1].BidID
	}
	log.Info("list bids: response",
		zap.String("auctionId", auctionId),
		zap.Int("count", len(items)),
		zap.Bool("hasMore", res.HasMore),
		zap.String("nextCursor", strDeref(res.NextCursor)),
		zap.String("firstBidId", first),
		zap.String("lastBidId", last),
	)

	body := openapi.ListBidsResponse{
		Items:      items,
		HasMore:    &res.HasMore,
		NextCursor: nil,
	}
	if res.NextCursor != nil {
		body.NextCursor = res.NextCursor
		c.Header("X-Next-Cursor", *res.NextCursor)
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, body)
}

func (h *HttpController) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, list_bids.ErrInvalidCursor):
		writeProblem(c, http.StatusBadRequest,
			"https://example.com/problems/invalid-cursor",
			"Invalid cursor",
			"The supplied cursor could not be parsed",
		)
	case errors.Is(err, list_bids.ErrAuctionNotFound):
		writeProblem(c, http.StatusNotFound,
			"https://example.com/problems/auction-not-found",
			"Auction not found",
			"No bids exist or the auction is unknown",
		)
	default:
		h.log.Error("list bids failed", zap.Error(err))
		writeProblem(c, http.StatusInternalServerError,
			"https://example.com/problems/internal",
			"Internal Server Error",
			"An unexpected error occurred",
		)
	}
}
