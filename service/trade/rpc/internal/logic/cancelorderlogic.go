// Package logic contains business logic implementations for trade service.
package logic

import (
	"context"
	"fmt"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/service/trade/rpc"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// CancelOrderLogic handles order cancellation logic.
type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewCancelOrderLogic creates a new CancelOrderLogic instance.
func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CancelOrder cancels an order.
//
// Responsibilities:
//   - Validate the cancel request
//   - Load and verify order ownership and status
//   - Transition order status with optimistic locking
//   - Coordinate side effects (inventory rollback if needed)
func (l *CancelOrderLogic) CancelOrder(req *rpc.CancelOrderRequest) (*rpc.CancelOrderResponse, error) {
	if req == nil {
		l.Errorf("received nil CancelOrderRequest")
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

	l.Infof("canceling order: userId=%d, orderId=%d", req.UserId, req.OrderId)

	// Load order from database
	if l.svcCtx.OrderRepo == nil {
		l.Errorf("order repository not initialized")
		return nil, fmt.Errorf("order repository not available")
	}

	order, err := l.svcCtx.OrderRepo.GetByID(l.ctx, req.OrderId)
	if err != nil {
		l.Errorf("failed to load order: %v, orderId=%d", err, req.OrderId)
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Verify ownership
	if order.UserID != req.UserId {
		l.Errorf("order ownership mismatch: orderUserId=%d, requestUserId=%d, orderId=%d",
			order.UserID, req.UserId, req.OrderId)
		return nil, fmt.Errorf("order does not belong to user")
	}

	// Verify order can be canceled (only pending payment orders can be canceled)
	if order.Status != database.OrderStatusPendingPayment {
		l.Errorf("order cannot be canceled: orderId=%d, currentStatus=%d", req.OrderId, order.Status)
		return nil, fmt.Errorf("order cannot be canceled: current status is %d", order.Status)
	}

	// Update order status with optimistic locking
	err = l.svcCtx.OrderRepo.UpdateStatus(
		l.ctx,
		req.OrderId,
		database.OrderStatusPendingPayment,
		database.OrderStatusClosed,
		order.Version,
	)
	if err != nil {
		l.Errorf("failed to update order status: %v, orderId=%d", err, req.OrderId)
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	l.Infof("order canceled successfully: orderId=%d, userId=%d", req.OrderId, req.UserId)

	// Note: Inventory rollback would typically be handled by a consumer
	// listening to order cancellation events. For now, we just update the order status.
	// If inventory was already deducted (via RocketMQ), a separate rollback message
	// would need to be sent to Promotion RPC.

	return &rpc.CancelOrderResponse{
		OrderId: req.OrderId,
		Status:  database.OrderStatusClosed,
		Success: true,
	}, nil
}
