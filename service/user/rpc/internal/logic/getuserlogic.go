// Package logic contains business logic implementations for user service.
package logic

import (
	"context"

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
func (l *GetUserLogic) GetUser(_ *rpc.GetUserRequest) (*rpc.GetUserResponse, error) {
	// todo: add your logic here and delete this line

	return &rpc.GetUserResponse{}, nil
}
