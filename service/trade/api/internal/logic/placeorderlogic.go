// Package logic contains business logic for trade-api handlers.
package logic

import (
	"context"
	"fmt"

	"github.com/aether-defense-system/common/snowflake"
	"github.com/aether-defense-system/service/trade/api/internal/svc"
	"github.com/aether-defense-system/service/trade/api/internal/types"
	"github.com/aether-defense-system/service/trade/rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

// PlaceOrderLogic handles order placement logic.
type PlaceOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPlaceOrderLogic creates a new PlaceOrderLogic instance.
func NewPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceOrderLogic {
	return &PlaceOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PlaceOrder places an order by calling Trade RPC.
func (l *PlaceOrderLogic) PlaceOrder(req *types.PlaceOrderReq, userID int64) (resp *types.PlaceOrderResp, err error) {
	// Generate order ID if not provided
	orderID := req.OrderID
	if orderID <= 0 {
		orderID, err = snowflake.Next()
		if err != nil {
			l.Errorf("failed to generate order ID: %v", err)
			return nil, fmt.Errorf("failed to generate order ID: %w", err)
		}
	}

	// Calculate real amount (simplified - in production, calculate from course prices and coupons)
	// For now, use a placeholder amount
	realAmount := int32(10000) // 100.00 in cents (placeholder)

	// Call Trade RPC to place order
	rpcReq := &rpc.PlaceOrderRequest{
		UserId:     userID,
		OrderId:    orderID,
		CourseIds:  req.CourseIDs,
		CouponIds:  req.CouponIDs,
		RealAmount: realAmount,
	}

	rpcResp, err := l.svcCtx.TradeRPC.PlaceOrder(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("failed to place order via RPC: %v, userID=%d, orderID=%d", err, userID, orderID)
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	return &types.PlaceOrderResp{
		OrderID:   rpcResp.OrderId,
		PayAmount: int(rpcResp.PayAmount),
		Status:    int(rpcResp.Status),
	}, nil
}
