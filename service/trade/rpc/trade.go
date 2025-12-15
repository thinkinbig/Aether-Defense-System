package rpc

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

// TradeService trading service struct.
type TradeService struct{}

// PlaceOrder place order.
func (s *TradeService) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
	// Parameter validation
	if req.UserId <= 0 {
		logx.WithContext(ctx).Errorf("Invalid user_id: %d", req.UserId)
		return nil, fmt.Errorf("invalid user_id: %d", req.UserId)
	}
	if req.OrderId <= 0 {
		logx.WithContext(ctx).Errorf("Invalid order_id: %d", req.OrderId)
		return nil, fmt.Errorf("invalid order_id: %d", req.OrderId)
	}
	if len(req.CourseIds) == 0 {
		logx.WithContext(ctx).Errorf("Empty course_ids for user_id: %d", req.UserId)
		return nil, fmt.Errorf("course_ids cannot be empty")
	}
	if req.RealAmount <= 0 {
		logx.WithContext(ctx).Errorf("Invalid real_amount: %d for order_id: %d", req.RealAmount, req.OrderId)
		return nil, fmt.Errorf("real_amount must be greater than 0")
	}

	// TODO: Implement order placement logic
	// 1. Send RocketMQ Half Message
	// 2. Execute local transaction
	// 3. Commit or Rollback message based on execution result

	logx.WithContext(ctx).Infof("Placing order: userId=%d, orderId=%d, courseIds=%v, couponIds=%v, realAmount=%d cents",
		req.UserId, req.OrderId, req.CourseIds, req.CouponIds, req.RealAmount)

	return &PlaceOrderResponse{
		OrderId:   req.OrderId,
		PayAmount: int32(req.RealAmount),
		Status:    1, // Pending Payment
	}, nil
}

// CancelOrder cancel order.
func (s *TradeService) CancelOrder(ctx context.Context, req *CancelOrderRequest) (*CancelOrderResponse, error) {
	// Parameter validation
	if req.UserId <= 0 {
		logx.WithContext(ctx).Errorf("Invalid user_id: %d", req.UserId)
		return nil, fmt.Errorf("invalid user_id: %d", req.UserId)
	}
	if req.OrderId <= 0 {
		logx.WithContext(ctx).Errorf("Invalid order_id: %d", req.OrderId)
		return nil, fmt.Errorf("invalid order_id: %d", req.OrderId)
	}

	logx.WithContext(ctx).Infof("Canceling order: userId=%d, orderId=%d", req.UserId, req.OrderId)

	// TODO: Implement order cancellation logic
	// 1. Query order by orderId and userId (ensure user owns the order)
	// 2. Check order status - only PendingPayment (1) can be canceled
	// 3. Update order status to Closed (2) with optimistic lock (version field)
	// 4. If order contains coupons, restore coupon usage status
	// 5. If order contains inventory deduction, restore inventory
	// 6. Send cancellation notification via RocketMQ

	// For now, return success response
	// In production, this should:
	// - Query database to verify order exists and belongs to user
	// - Check if status is PendingPayment (1)
	// - Update status to Closed (2) using optimistic lock
	// - Handle related resources (coupons, inventory, etc.)

	return &CancelOrderResponse{
		OrderId: req.OrderId,
		Status:  2, // Closed
		Success: true,
	}, nil
}
