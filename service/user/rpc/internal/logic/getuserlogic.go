package logic

import (
	"context"

	"github.com/aether-defense-system/service/user/rpc"
	"github.com/aether-defense-system/service/user/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get User Information Interface
func (l *GetUserLogic) GetUser(in *rpc.GetUserRequest) (*rpc.GetUserResponse, error) {
	// todo: add your logic here and delete this line

	return &rpc.GetUserResponse{}, nil
}
