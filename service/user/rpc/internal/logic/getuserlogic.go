// Package logic contains business logic implementations for user service.
package logic

import (
	"context"
	"fmt"

	"github.com/aether-defense-system/service/user/rpc"
	"github.com/aether-defense-system/service/user/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetUserLogic handles user information retrieval logic.
type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetUserLogic creates a new GetUserLogic instance.
func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUser gets user information.
//
// Responsibilities:
//   - Validate the incoming request (user ID)
//   - Load user information from the user domain (via svcCtx)
//   - Map domain user to RPC response
//
// Currently we implement validation and return a deterministic stub user;
// repository and cache integration will be added via svcCtx.
func (l *GetUserLogic) GetUser(req *rpc.GetUserRequest) (*rpc.GetUserResponse, error) {
	if req == nil {
		l.Errorf("received nil GetUserRequest")
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.UserId <= 0 {
		l.Errorf("invalid user_id: %d", req.UserId)
		return nil, fmt.Errorf("invalid user_id: %d", req.UserId)
	}

	l.Infof("getting user information for userId=%d", req.UserId)

	// TODO: fetch user from repository/cache via svcCtx.
	// For now we return a predictable stub that can be used in callers and tests.
	return &rpc.GetUserResponse{
		UserId:   req.UserId,
		Username: "testuser",
		Mobile:   "13800138000",
	}, nil
}
