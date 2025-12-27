package logic

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/aether-defense-system/service/trade/rpc"
	"github.com/aether-defense-system/service/trade/rpc/internal/config"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"
	userservice "github.com/aether-defense-system/service/user/rpc/userservice"
)

// mockUserService mocks the UserService interface
type mockUserService struct {
	getUserFunc func(ctx context.Context, in *userservice.GetUserRequest, opts ...grpc.CallOption) (*userservice.GetUserResponse, error)
}

func (m *mockUserService) GetUser(ctx context.Context, in *userservice.GetUserRequest, opts ...grpc.CallOption) (*userservice.GetUserResponse, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, in, opts...)
	}
	return &userservice.GetUserResponse{
		UserId:   in.UserId,
		Username: "testuser",
		Mobile:   "13800138000",
	}, nil
}

func TestPlaceOrderLogic_PlaceOrder_ValidationErrors(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	tests := []struct {
		name    string
		req     *rpc.PlaceOrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "request cannot be nil",
		},
		{
			name: "invalid user id - zero",
			req: &rpc.PlaceOrderRequest{
				UserId:     0,
				OrderId:    1,
				CourseIds:  []int64{1},
				RealAmount: 100,
			},
			wantErr: true,
			errMsg:  "invalid user_id",
		},
		{
			name: "invalid user id - negative",
			req: &rpc.PlaceOrderRequest{
				UserId:     -1,
				OrderId:    1,
				CourseIds:  []int64{1},
				RealAmount: 100,
			},
			wantErr: true,
			errMsg:  "invalid user_id",
		},
		{
			name: "invalid order id - zero",
			req: &rpc.PlaceOrderRequest{
				UserId:     1,
				OrderId:    0,
				CourseIds:  []int64{1},
				RealAmount: 100,
			},
			wantErr: true,
			errMsg:  "invalid order_id",
		},
		{
			name: "invalid order id - negative",
			req: &rpc.PlaceOrderRequest{
				UserId:     1,
				OrderId:    -1,
				CourseIds:  []int64{1},
				RealAmount: 100,
			},
			wantErr: true,
			errMsg:  "invalid order_id",
		},
		{
			name: "empty course ids",
			req: &rpc.PlaceOrderRequest{
				UserId:     1,
				OrderId:    1,
				CourseIds:  []int64{},
				RealAmount: 100,
			},
			wantErr: true,
			errMsg:  "course_ids cannot be empty",
		},
		{
			name: "invalid real amount - zero",
			req: &rpc.PlaceOrderRequest{
				UserId:     1,
				OrderId:    1,
				CourseIds:  []int64{1},
				RealAmount: 0,
			},
			wantErr: true,
			errMsg:  "real_amount must be greater than 0",
		},
		{
			name: "invalid real amount - negative",
			req: &rpc.PlaceOrderRequest{
				UserId:     1,
				OrderId:    1,
				CourseIds:  []int64{1},
				RealAmount: -1,
			},
			wantErr: true,
			errMsg:  "real_amount must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := logic.PlaceOrder(tt.req)

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

func TestPlaceOrderLogic_PlaceOrder_UserRPCNotInitialized(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg, UserRPC: nil}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	req := &rpc.PlaceOrderRequest{
		UserId:     1,
		OrderId:    1,
		CourseIds:  []int64{1},
		RealAmount: 100,
	}

	resp, err := logic.PlaceOrder(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user validation service not available")
	assert.Nil(t, resp)
}

func TestPlaceOrderLogic_PlaceOrder_UserNotFound(t *testing.T) {
	cfg := &config.Config{}
	mockUserRPC := &mockUserService{
		getUserFunc: func(ctx context.Context, in *userservice.GetUserRequest, opts ...grpc.CallOption) (*userservice.GetUserResponse, error) {
			return nil, fmt.Errorf("user not found")
		},
	}
	svcCtx := &svc.ServiceContext{
		Config:  cfg,
		UserRPC: mockUserRPC,
	}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	req := &rpc.PlaceOrderRequest{
		UserId:     1,
		OrderId:    1,
		CourseIds:  []int64{1},
		RealAmount: 100,
	}

	resp, err := logic.PlaceOrder(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found or invalid")
	assert.Nil(t, resp)
}

func TestPlaceOrderLogic_PlaceOrder_UserValidationSuccess(t *testing.T) {
	cfg := &config.Config{}
	mockUserRPC := &mockUserService{
		getUserFunc: func(ctx context.Context, in *userservice.GetUserRequest, opts ...grpc.CallOption) (*userservice.GetUserResponse, error) {
			return &userservice.GetUserResponse{
				UserId:   in.UserId,
				Username: "testuser",
				Mobile:   "13800138000",
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{
		Config:  cfg,
		UserRPC: mockUserRPC,
		// RocketMQ is nil, so it will try to create one, which will fail without proper config
		// This tests the error path when RocketMQ initialization fails
		RocketMQ: nil,
	}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	req := &rpc.PlaceOrderRequest{
		UserId:     1,
		OrderId:    1,
		CourseIds:  []int64{1},
		RealAmount: 100,
	}

	// This will fail when trying to create RocketMQ producer (no real broker available)
	resp, err := logic.PlaceOrder(req)
	// We expect an error because we can't create a real RocketMQ producer in unit tests
	assert.Error(t, err)
	assert.Nil(t, resp)
	// Error should be related to RocketMQ initialization
	assert.Contains(t, err.Error(), "failed to initialize message queue", "Expected RocketMQ initialization error, got: %v", err)
}

func TestPlaceOrderLogic_PlaceOrder_UserRPCError(t *testing.T) {
	cfg := &config.Config{}
	mockUserRPC := &mockUserService{
		getUserFunc: func(ctx context.Context, in *userservice.GetUserRequest, opts ...grpc.CallOption) (*userservice.GetUserResponse, error) {
			return nil, fmt.Errorf("rpc connection error")
		},
	}
	svcCtx := &svc.ServiceContext{
		Config:  cfg,
		UserRPC: mockUserRPC,
	}
	logic := NewPlaceOrderLogic(context.Background(), svcCtx)

	req := &rpc.PlaceOrderRequest{
		UserId:     1,
		OrderId:    1,
		CourseIds:  []int64{1},
		RealAmount: 100,
	}

	resp, err := logic.PlaceOrder(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found or invalid")
	assert.Nil(t, resp)
}

func TestPlaceOrderLogic_NewPlaceOrderLogic(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	ctx := context.Background()

	logic := NewPlaceOrderLogic(ctx, svcCtx)
	assert.NotNil(t, logic)
	assert.Equal(t, ctx, logic.ctx)
	assert.Equal(t, svcCtx, logic.svcCtx)
}
