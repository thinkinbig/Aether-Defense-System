// Package logic contains business logic for the user HTTP API.
package logic

import (
	"context"

	"github.com/aether-defense-system/cmd/api/user-api/internal/svc"
	"github.com/aether-defense-system/cmd/api/user-api/internal/types"
	"github.com/aether-defense-system/service/user/rpc/userservice"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetUserLogic contains business logic for the user HTTP API.
type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetUserLogic creates a new GetUserLogic.
func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetUser fetches user information by ID via the user RPC service.
func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (*types.GetUserResponse, error) {
	if req.UserID <= 0 {
		l.Errorf("invalid user_id: %d", req.UserID)
		return nil, types.ErrInvalidUserID
	}

	rpcResp, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userservice.GetUserRequest{
		UserId: req.UserID,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetUserResponse{
		UserID:   rpcResp.UserId,
		Username: rpcResp.Username,
		Mobile:   rpcResp.Mobile,
	}, nil
}
