// Package handler provides HTTP handlers for the trade API service.
package handler

import (
	"net/http"

	"github.com/aether-defense-system/service/trade/api/internal/logic"
	"github.com/aether-defense-system/service/trade/api/internal/svc"
	"github.com/aether-defense-system/service/trade/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// PlaceOrderHandler handles POST /v1/trade/order/place requests.
func PlaceOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PlaceOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// Extract user_id from JWT token (set by go-zero JWT middleware)
		// The JWT middleware validates the token and sets userId in context
		userIDVal := r.Context().Value("userId")
		if userIDVal == nil {
			logx.WithContext(r.Context()).Errorf("missing user_id in JWT token")
			http.Error(w, "unauthorized: missing user_id", http.StatusUnauthorized)
			return
		}

		userID, ok := userIDVal.(int64)
		if !ok {
			logx.WithContext(r.Context()).Errorf("invalid user_id type in JWT token: %T", userIDVal)
			http.Error(w, "unauthorized: invalid user_id", http.StatusUnauthorized)
			return
		}

		if userID <= 0 {
			logx.WithContext(r.Context()).Errorf("invalid user_id value: %d", userID)
			http.Error(w, "unauthorized: invalid user_id", http.StatusUnauthorized)
			return
		}

		l := logic.NewPlaceOrderLogic(r.Context(), svcCtx)
		resp, err := l.PlaceOrder(&req, userID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
