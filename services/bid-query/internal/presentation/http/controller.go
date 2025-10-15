package http

import (
	"kei-services/services/bid-query/internal/application/list_bids"
	"kei-services/services/bid-query/openapi"

	"go.uber.org/zap"
)

type HttpController struct {
	log *zap.Logger
	svc list_bids.IService
}

func NewHttpController(log *zap.Logger, svc list_bids.IService) *HttpController {
	return &HttpController{log: log, svc: svc}
}

var _ openapi.ServerInterface = (*HttpController)(nil)
