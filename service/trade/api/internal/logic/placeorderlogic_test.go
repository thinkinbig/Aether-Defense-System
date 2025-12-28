package logic

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/aether-defense-system/service/trade/api/internal/svc"
	"github.com/aether-defense-system/service/trade/api/internal/types"
	"github.com/aether-defense-system/service/trade/rpc/tradeservice"
)

// mockTradeRPC mocks the TradeRPC service.
type mockTradeRPC struct {
	placeOrderFunc func(
		ctx context.Context,
		req *tradeservice.PlaceOrderRequest,
	) (*tradeservice.PlaceOrderResponse, error)
	cancelOrderFunc func(
		ctx context.Context,
		req *tradeservice.CancelOrderRequest,
	) (*tradeservice.CancelOrderResponse, error)
}

func (m *mockTradeRPC) PlaceOrder(
	ctx context.Context,
	req *tradeservice.PlaceOrderRequest,
	_ ...grpc.CallOption,
) (*tradeservice.PlaceOrderResponse, error) {
	if m.placeOrderFunc != nil {
		return m.placeOrderFunc(ctx, req)
	}
	return &tradeservice.PlaceOrderResponse{
		OrderId:   req.OrderId,
		PayAmount: req.RealAmount,
		Status:    1,
	}, nil
}

func (m *mockTradeRPC) CancelOrder(
	ctx context.Context,
	req *tradeservice.CancelOrderRequest,
	_ ...grpc.CallOption,
) (*tradeservice.CancelOrderResponse, error) {
	if m.cancelOrderFunc != nil {
		return m.cancelOrderFunc(ctx, req)
	}
	return &tradeservice.CancelOrderResponse{
		OrderId: req.OrderId,
		Status:  2,
		Success: true,
	}, nil
}

func TestPlaceOrderLogic_PlaceOrder_ValidationErrors(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	tests := []struct {
		req     *types.PlaceOrderReq
		name    string
		errMsg  string
		userID  int64
		wantErr bool
	}{
		{
			name: "empty course ids",
			req: &types.PlaceOrderReq{
				CourseIDs: []int64{},
			},
			userID:  1,
			wantErr: true,
			errMsg:  "course_ids cannot be empty",
		},
		{
			name: "too many courses",
			req: &types.PlaceOrderReq{
				CourseIDs: make([]int64, 101), // 101 courses
			},
			userID:  1,
			wantErr: true,
			errMsg:  "too many courses, maximum 100 allowed",
		},
		{
			name: "invalid course id - zero",
			req: &types.PlaceOrderReq{
				CourseIDs: []int64{0},
			},
			userID:  1,
			wantErr: true,
			errMsg:  "invalid course_id: 0",
		},
		{
			name: "invalid course id - negative",
			req: &types.PlaceOrderReq{
				CourseIDs: []int64{-1},
			},
			userID:  1,
			wantErr: true,
			errMsg:  "invalid course_id: -1",
		},
		{
			name: "invalid course id in middle",
			req: &types.PlaceOrderReq{
				CourseIDs: []int64{1, 0, 3},
			},
			userID:  1,
			wantErr: true,
			errMsg:  "invalid course_id: 0",
		},
		{
			name: "too many coupons",
			req: &types.PlaceOrderReq{
				CourseIDs: []int64{1},
				CouponIDs: make([]int64, 51), // 51 coupons
			},
			userID:  1,
			wantErr: true,
			errMsg:  "too many coupons, maximum 50 allowed",
		},
		{
			name: "invalid coupon id - zero",
			req: &types.PlaceOrderReq{
				CourseIDs: []int64{1},
				CouponIDs: []int64{0},
			},
			userID:  1,
			wantErr: true,
			errMsg:  "invalid coupon_id: 0",
		},
		{
			name: "invalid coupon id - negative",
			req: &types.PlaceOrderReq{
				CourseIDs: []int64{1},
				CouponIDs: []int64{-1},
			},
			userID:  1,
			wantErr: true,
			errMsg:  "invalid coupon_id: -1",
		},
		{
			name: "invalid order id - negative",
			req: &types.PlaceOrderReq{
				CourseIDs: []int64{1},
				OrderID:   -1,
			},
			userID:  1,
			wantErr: true,
			errMsg:  "invalid order_id: -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := logic.PlaceOrder(tt.req, tt.userID)

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

func TestPlaceOrderLogic_PlaceOrder_Success(t *testing.T) {
	mockRPC := &mockTradeRPC{
		placeOrderFunc: func(
			_ context.Context,
			req *tradeservice.PlaceOrderRequest,
		) (*tradeservice.PlaceOrderResponse, error) {
			return &tradeservice.PlaceOrderResponse{
				OrderId:   req.OrderId,
				PayAmount: req.RealAmount,
				Status:    1,
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{
		TradeRPC: mockRPC,
	}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	req := &types.PlaceOrderReq{
		CourseIDs: []int64{1, 2, 3},
		CouponIDs: []int64{10, 20},
		OrderID:   100,
	}

	resp, err := logic.PlaceOrder(req, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(100), resp.OrderID)
	assert.Equal(t, 10000, resp.PayAmount) // Placeholder amount
	assert.Equal(t, 1, resp.Status)
}

func TestPlaceOrderLogic_PlaceOrder_GenerateOrderID(t *testing.T) {
	mockRPC := &mockTradeRPC{
		placeOrderFunc: func(
			_ context.Context,
			req *tradeservice.PlaceOrderRequest,
		) (*tradeservice.PlaceOrderResponse, error) {
			// Verify that order ID was generated (should be > 0)
			assert.Greater(t, req.OrderId, int64(0))
			return &tradeservice.PlaceOrderResponse{
				OrderId:   req.OrderId,
				PayAmount: req.RealAmount,
				Status:    1,
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{
		TradeRPC: mockRPC,
	}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	req := &types.PlaceOrderReq{
		CourseIDs: []int64{1},
		OrderID:   0, // Will trigger order ID generation
	}

	resp, err := logic.PlaceOrder(req, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Greater(t, resp.OrderID, int64(0))
}

func TestPlaceOrderLogic_PlaceOrder_RPCError(t *testing.T) {
	mockRPC := &mockTradeRPC{
		placeOrderFunc: func(_ context.Context, _ *tradeservice.PlaceOrderRequest) (*tradeservice.PlaceOrderResponse, error) {
			return nil, fmt.Errorf("rpc error")
		},
	}
	svcCtx := &svc.ServiceContext{
		TradeRPC: mockRPC,
	}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	req := &types.PlaceOrderReq{
		CourseIDs: []int64{1},
	}

	resp, err := logic.PlaceOrder(req, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to place order")
	assert.Nil(t, resp)
}

func TestPlaceOrderLogic_NewPlaceOrderLogic(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	ctx := context.Background()

	logic := NewPlaceOrderLogic(ctx, svcCtx)
	assert.NotNil(t, logic)
	assert.Equal(t, ctx, logic.ctx)
	assert.Equal(t, svcCtx, logic.svcCtx)
}
