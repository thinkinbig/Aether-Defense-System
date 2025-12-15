// Package handler wires HTTP routes for user-api.
package handler

import (
	"net/http"

	"github.com/aether-defense-system/cmd/api/user-api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers registers HTTP routes for user-api.
func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/health",
				Handler: HealthHandler,
			},
			{
				Method:  http.MethodGet,
				Path:    "/v1/users/:userId",
				Handler: GetUserHandler(svcCtx),
			},
		},
	)
}
