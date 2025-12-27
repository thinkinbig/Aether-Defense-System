package logic

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/aether-defense-system/service/user/api/internal/svc"
	"github.com/aether-defense-system/service/user/api/internal/types"
	userservice "github.com/aether-defense-system/service/user/rpc/userservice"
)

// mockUserRPC mocks the UserRPC service
type mockUserRPC struct {
	getUserFunc func(ctx context.Context, in *userservice.GetUserRequest, opts ...grpc.CallOption) (*userservice.GetUserResponse, error)
}

func (m *mockUserRPC) GetUser(ctx context.Context, in *userservice.GetUserRequest, opts ...grpc.CallOption) (*userservice.GetUserResponse, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, in, opts...)
	}
	return &userservice.GetUserResponse{
		UserId:   in.UserId,
		Username: "testuser",
		Mobile:   "13800138000",
	}, nil
}

func TestGetUserLogic_GetUser_ValidationErrors(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	logic := NewGetUserLogic(context.Background(), svcCtx)

	tests := []struct {
		name    string
		req     *types.GetUserRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid user id - zero",
			req: &types.GetUserRequest{
				UserID: 0,
			},
			wantErr: true,
			errMsg:  "invalid user_id",
		},
		{
			name: "invalid user id - negative",
			req: &types.GetUserRequest{
				UserID: -1,
			},
			wantErr: true,
			errMsg:  "invalid user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := logic.GetUser(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, types.ErrInvalidUserID, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestGetUserLogic_GetUser_Success(t *testing.T) {
	mockRPC := &mockUserRPC{
		getUserFunc: func(ctx context.Context, in *userservice.GetUserRequest, opts ...grpc.CallOption) (*userservice.GetUserResponse, error) {
			return &userservice.GetUserResponse{
				UserId:   in.UserId,
				Username: "testuser",
				Mobile:   "13800138000",
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{
		UserRPC: mockRPC,
	}
	logic := NewGetUserLogic(context.Background(), svcCtx)

	req := &types.GetUserRequest{
		UserID: 1,
	}

	resp, err := logic.GetUser(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.UserID)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, "13800138000", resp.Mobile)
}

func TestGetUserLogic_GetUser_RPCError(t *testing.T) {
	mockRPC := &mockUserRPC{
		getUserFunc: func(ctx context.Context, in *userservice.GetUserRequest, opts ...grpc.CallOption) (*userservice.GetUserResponse, error) {
			return nil, fmt.Errorf("user not found")
		},
	}
	svcCtx := &svc.ServiceContext{
		UserRPC: mockRPC,
	}
	logic := NewGetUserLogic(context.Background(), svcCtx)

	req := &types.GetUserRequest{
		UserID: 1,
	}

	resp, err := logic.GetUser(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	assert.Nil(t, resp)
}

func TestGetUserLogic_NewGetUserLogic(t *testing.T) {
	svcCtx := &svc.ServiceContext{}
	ctx := context.Background()

	logic := NewGetUserLogic(ctx, svcCtx)
	assert.NotNil(t, logic)
	assert.Equal(t, ctx, logic.ctx)
	assert.Equal(t, svcCtx, logic.svcCtx)
}
