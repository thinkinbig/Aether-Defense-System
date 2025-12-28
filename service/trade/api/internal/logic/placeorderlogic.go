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
	// Input validation: CourseIDs
	if len(req.CourseIDs) == 0 {
		l.Errorf("empty course_ids for user_id: %d", userID)
		return nil, fmt.Errorf("course_ids cannot be empty")
	}
	if len(req.CourseIDs) > 100 {
		l.Errorf("too many courses: %d (max 100) for user_id: %d", len(req.CourseIDs), userID)
		return nil, fmt.Errorf("too many courses, maximum 100 allowed")
	}
	for i, courseID := range req.CourseIDs {
		if courseID <= 0 {
			l.Errorf("invalid course_id at index %d: %d for user_id: %d", i, courseID, userID)
			return nil, fmt.Errorf("invalid course_id: %d", courseID)
		}
	}

	// Input validation: CouponIDs (if provided)
	if len(req.CouponIDs) > 50 {
		l.Errorf("too many coupons: %d (max 50) for user_id: %d", len(req.CouponIDs), userID)
		return nil, fmt.Errorf("too many coupons, maximum 50 allowed")
	}
	for i, couponID := range req.CouponIDs {
		if couponID <= 0 {
			l.Errorf("invalid coupon_id at index %d: %d for user_id: %d", i, couponID, userID)
			return nil, fmt.Errorf("invalid coupon_id: %d", couponID)
		}
	}

	// Input validation: OrderID (if provided)
	if req.OrderID < 0 {
		l.Errorf("invalid order_id: %d for user_id: %d", req.OrderID, userID)
		return nil, fmt.Errorf("invalid order_id: %d", req.OrderID)
	}

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
