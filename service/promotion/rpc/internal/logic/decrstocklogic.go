package logic

import (
	"context"

	"github.com/aether-defense-system/service/promotion/rpc"
	"github.com/aether-defense-system/service/promotion/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DecrStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDecrStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DecrStockLogic {
	return &DecrStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Decrement Inventory Interface
func (l *DecrStockLogic) DecrStock(in *rpc.DecrStockRequest) (*rpc.DecrStockResponse, error) {
	// todo: add your logic here and delete this line

	return &rpc.DecrStockResponse{}, nil
}
