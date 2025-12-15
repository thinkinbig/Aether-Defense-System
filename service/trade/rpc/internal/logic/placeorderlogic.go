// Package logic contains business logic implementations for trade service.
package logic

import (
	"context"
	"fmt"

	"github.com/aether-defense-system/service/trade/rpc"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// PlaceOrderLogic handles order placement logic.
type PlaceOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewPlaceOrderLogic creates a new PlaceOrderLogic instance.
func NewPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceOrderLogic {
	return &PlaceOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PlaceOrder places an order.
//
// This method is the primary entrypoint for trade order creation business logic.
// It is responsible for:
//   - Validating the request
//   - Performing domain-level operations (persisting order, MQ, etc.)
//   - Returning a meaningful response to the caller
//
// For now, we implement full input validation and a realistic skeleton that can
// be extended with repositories, RPC clients, and MQ producers via svcCtx.
func (l *PlaceOrderLogic) PlaceOrder(req *rpc.PlaceOrderRequest) (*rpc.PlaceOrderResponse, error) {
	// Parameter validation (business rules)
	if req == nil {
		l.Errorf("received nil PlaceOrderRequest")
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.UserId <= 0 {
		l.Errorf("invalid user_id: %d", req.UserId)
		return nil, fmt.Errorf("invalid user_id: %d", req.UserId)
	}

	if req.OrderId <= 0 {
		l.Errorf("invalid order_id: %d", req.OrderId)
		return nil, fmt.Errorf("invalid order_id: %d", req.OrderId)
	}

	if len(req.CourseIds) == 0 {
		l.Errorf("empty course_ids for user_id: %d", req.UserId)
		return nil, fmt.Errorf("course_ids cannot be empty")
	}

	if req.RealAmount <= 0 {
		l.Errorf("invalid real_amount: %d for order_id: %d", req.RealAmount, req.OrderId)
		return nil, fmt.Errorf("real_amount must be greater than 0")
	}

	// TODO: Inject and use domain dependencies via svcCtx, for example:
	//   - OrderRepository to persist the order
	//   - Promotion/User RPC clients to validate coupons and user state
	//   - MQ producer to send transactional messages

	l.Infof("placing order: userId=%d, orderId=%d, courseIds=%v, couponIds=%v, realAmount=%d cents",
		req.UserId, req.OrderId, req.CourseIds, req.CouponIds, req.RealAmount)

	// Skeleton response: pending payment.
	// As domain logic evolves, this should reflect real order status and amounts.
	return &rpc.PlaceOrderResponse{
		OrderId:   req.OrderId,
		PayAmount: req.RealAmount,
		Status:    1, // Pending Payment
	}, nil
}
