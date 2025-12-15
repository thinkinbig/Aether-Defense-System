package handler

import (
	"net/http"

	"github.com/aether-defense-system/cmd/api/user-api/internal/logic"
	"github.com/aether-defense-system/cmd/api/user-api/internal/svc"
	"github.com/aether-defense-system/cmd/api/user-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetUserHandler handles GET /v1/users/:userId requests.
func GetUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUserRequest
		if err := httpx.ParsePath(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetUserLogic(r.Context(), svcCtx)
		resp, err := l.GetUser(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
