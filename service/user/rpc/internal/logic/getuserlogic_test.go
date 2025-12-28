package logic

import (
	"context"
	"fmt"
	"testing"

	"github.com/aether-defense-system/common/database"
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
			wantErr: true, // still error in this table-driven test (repo not set here)
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

type fakeUserRepo struct {
	user *database.User
	err  error
}

func (f *fakeUserRepo) GetByID(_ context.Context, _ int64) (*database.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return f.user, nil
}

func TestGetUserLogic_GetUser_Success_WithRepo(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{
		Config: cfg,
		UserRepo: &fakeUserRepo{
			user: &database.User{ID: 1, Username: "u1", Mobile: "13800000000"},
		},
	}
	logic := NewGetUserLogic(context.Background(), svcCtx)

	resp, err := logic.GetUser(&rpc.GetUserRequest{UserId: 1})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	if resp.UserId != 1 || resp.Username != "u1" || resp.Mobile != "13800000000" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGetUserLogic_GetUser_RepoError(t *testing.T) {
	cfg := &config.Config{}
	svcCtx := &svc.ServiceContext{
		Config:   cfg,
		UserRepo: &fakeUserRepo{err: fmt.Errorf("db down")},
	}
	logic := NewGetUserLogic(context.Background(), svcCtx)

	_, err := logic.GetUser(&rpc.GetUserRequest{UserId: 1})
	if err == nil {
		t.Fatalf("expected error")
	}
}
