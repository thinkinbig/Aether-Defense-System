package rpc

import (
	"context"
	"fmt"
)

// TradeService trading service struct
type TradeService struct{}

// PlaceOrder place order
func (s *TradeService) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
	// TODO: Implement order placement logic
	// 1. Send RocketMQ Half Message
	// 2. Execute local transaction
	// 3. Commit or Rollback message based on execution result
	
	fmt.Printf("User ID=%d placing order, Order ID=%d, Course IDs=%v, Coupon IDs=%v, Actual Amount=%d cents\n", 
		req.UserId, req.OrderId, req.CourseIds, req.CouponIds, req.RealAmount)
	
	return &PlaceOrderResponse{
		OrderId:   req.OrderId,
		PayAmount: int32(req.RealAmount),
		Status:    1, // Pending Payment
	}, nil
}