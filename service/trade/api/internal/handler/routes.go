package handler

import (
	"github.com/aether-defense-system/service/trade/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers registers all HTTP handlers.
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  "POST",
				Path:    "/v1/trade/order/place",
				Handler: PlaceOrderHandler(serverCtx),
			},
		},
	)
}
