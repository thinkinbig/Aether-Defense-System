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

		// Extract user_id from JWT token
		// In go-zero, JWT claims are typically stored in the request context
		// For now, we'll extract it from the context or header
		// This is a placeholder - actual JWT extraction should be done via middleware
		userID := int64(0)
		if userIDVal := r.Context().Value("userId"); userIDVal != nil {
			if id, ok := userIDVal.(int64); ok {
				userID = id
			}
		}

		if userID <= 0 {
			logx.WithContext(r.Context()).Errorf("invalid or missing user_id in request")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
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
