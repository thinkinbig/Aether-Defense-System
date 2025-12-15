package logic

import (
	"context"
	"testing"

	"github.com/aether-defense-system/service/user/rpc"
	"github.com/aether-defense-system/service/user/rpc/internal/config"
	"github.com/aether-defense-system/service/user/rpc/internal/svc"
)

func TestGetUserLogic_GetUser_ValidationAndSuccess(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{Config: cfg}
	logic := NewGetUserLogic(context.Background(), svcCtx)

	tests := []struct {
		req     *rpc.GetUserRequest
		name    string
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "invalid user id",
			req: &rpc.GetUserRequest{
				UserId: 0,
			},
			wantErr: true,
		},
		{
			name: "success",
			req: &rpc.GetUserRequest{
				UserId: 1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := logic.GetUser(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (resp=%+v)", resp)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if resp == nil {
				t.Fatalf("expected non-nil response")
			}
			if resp.UserId != tt.req.UserId {
				t.Errorf("expected UserId=%d, got %d", tt.req.UserId, resp.UserId)
			}
			if resp.Username == "" {
				t.Errorf("expected non-empty Username")
			}
			if resp.Mobile == "" {
				t.Errorf("expected non-empty Mobile")
			}
		})
	}
}
