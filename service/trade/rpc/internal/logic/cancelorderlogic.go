// Package logic contains business logic implementations for trade service.
package logic

import (
	"context"
	"fmt"

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
//   - Transition order status and coordinate side effects (inventory, coupons, MQ)
//
// Currently this implements full validation and a realistic status transition
// skeleton; persistence and side effects will be added via injected dependencies.
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

	// TODO: Inject and use domain dependencies via svcCtx:
	//   - OrderRepository to load and update the order
	//   - Promotion/Inventory services to rollback side effects
	//   - MQ producer for cancellation notifications
	//
	// For now we assume:
	//   - Order exists
	//   - Status transition from PendingPayment(1) to Closed(2) succeeds.

	return &rpc.CancelOrderResponse{
		OrderId: req.OrderId,
		Status:  2, // Closed
		Success: true,
	}, nil
}
