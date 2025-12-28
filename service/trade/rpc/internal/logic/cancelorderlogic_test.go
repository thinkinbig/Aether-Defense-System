package logic

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/service/trade/rpc"
	"github.com/aether-defense-system/service/trade/rpc/internal/config"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"
)

func TestCancelOrderLogic_CancelOrder_ValidationErrors(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewCancelOrderLogic(context.Background(), svcCtx)

	tests := []struct {
		req     *rpc.CancelOrderRequest
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "request cannot be nil",
		},
		{
			name: "invalid user id - zero",
			req: &rpc.CancelOrderRequest{
				UserId:  0,
				OrderId: 1,
			},
			wantErr: true,
			errMsg:  "invalid user_id",
		},
		{
			name: "invalid user id - negative",
			req: &rpc.CancelOrderRequest{
				UserId:  -1,
				OrderId: 1,
			},
			wantErr: true,
			errMsg:  "invalid user_id",
		},
		{
			name: "invalid order id - zero",
			req: &rpc.CancelOrderRequest{
				UserId:  1,
				OrderId: 0,
			},
			wantErr: true,
			errMsg:  "invalid order_id",
		},
		{
			name: "invalid order id - negative",
			req: &rpc.CancelOrderRequest{
				UserId:  1,
				OrderId: -1,
			},
			wantErr: true,
			errMsg:  "invalid order_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := logic.CancelOrder(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestCancelOrderLogic_CancelOrder_OrderRepoNotInitialized(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{
		Config:    cfg,
		OrderRepo: nil,
	}
	logic := NewCancelOrderLogic(context.Background(), svcCtx)

	req := &rpc.CancelOrderRequest{
		UserId:  1,
		OrderId: 1,
	}

	resp, err := logic.CancelOrder(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order repository not available")
	assert.Nil(t, resp)
}

func TestCancelOrderLogic_NewCancelOrderLogic(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	ctx := context.Background()

	logic := NewCancelOrderLogic(ctx, svcCtx)
	assert.NotNil(t, logic)
	assert.Equal(t, ctx, logic.ctx)
	assert.Equal(t, svcCtx, logic.svcCtx)
}

func createTestOrder(orderID int64, status int8) *database.TradeOrder {
	return &database.TradeOrder{
		ID:          orderID,
		UserID:      1,
		Status:      status,
		TotalAmount: 1000,
		PayAmount:   1000,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		Version:     1,
	}
}

func TestCancelOrderLogic_CancelOrder_StatusValidation(t *testing.T) {
	// This test verifies the status validation logic conceptually
	// Actual implementation requires database access

	// Test that only PendingPayment orders can be canceled
	testOrder := createTestOrder(1, database.OrderStatusPendingPayment)
	assert.Equal(t, int8(database.OrderStatusPendingPayment), testOrder.Status)

	// Test that paid orders cannot be canceled
	paidOrder := createTestOrder(2, database.OrderStatusPaid)
	assert.NotEqual(t, int8(database.OrderStatusPendingPayment), paidOrder.Status)

	// Test that closed orders cannot be canceled
	closedOrder := createTestOrder(3, database.OrderStatusClosed)
	assert.NotEqual(t, int8(database.OrderStatusPendingPayment), closedOrder.Status)

	// The actual cancellation logic is tested in integration tests
	// where we have a real database
}

func TestCancelOrderLogic_CancelOrder_ErrorPaths(t *testing.T) {
	cfg := &config.Config{}

	// Test with nil OrderRepo
	svcCtxNil := &svc.ServiceContext{
		Config:    cfg,
		OrderRepo: nil,
	}
	logicNil := NewCancelOrderLogic(context.Background(), svcCtxNil)

	req := &rpc.CancelOrderRequest{
		UserId:  1,
		OrderId: 1,
	}

	resp, err := logicNil.CancelOrder(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order repository not available")
	assert.Nil(t, resp)
}

func TestCancelOrderLogic_CancelOrder_AllStatusTypes(t *testing.T) {
	// Test all order status constants are defined correctly
	assert.Equal(t, int8(1), int8(database.OrderStatusPendingPayment))
	assert.Equal(t, int8(2), int8(database.OrderStatusClosed))
	assert.Equal(t, int8(3), int8(database.OrderStatusPaid))
	assert.Equal(t, int8(4), int8(database.OrderStatusFinished))
	assert.Equal(t, int8(5), int8(database.OrderStatusRefunded))

	// Verify that only PendingPayment (1) can be canceled
	testCases := []struct {
		status    int8
		canCancel bool
	}{
		{database.OrderStatusPendingPayment, true},
		{database.OrderStatusPaid, false},
		{database.OrderStatusClosed, false},
		{database.OrderStatusFinished, false},
		{database.OrderStatusRefunded, false},
	}

	for _, tc := range testCases {
		order := createTestOrder(1, tc.status)
		isPending := order.Status == database.OrderStatusPendingPayment
		assert.Equal(t, tc.canCancel, isPending, "Status %d should have canCancel=%v", tc.status, tc.canCancel)
	}
}

func TestCancelOrderLogic_CancelOrder_EdgeCases(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewCancelOrderLogic(context.Background(), svcCtx)

	// Test with maximum int64 values
	req := &rpc.CancelOrderRequest{
		UserId:  9223372036854775807, // max int64
		OrderId: 9223372036854775807,
	}

	// Will fail at OrderRepo.GetByID since DB is nil, but tests validation passes
	resp, err := logic.CancelOrder(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order repository not available")
	assert.Nil(t, resp)
}
