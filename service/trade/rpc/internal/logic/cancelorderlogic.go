// Package logic contains business logic implementations for trade service.
package logic

import (
	"context"

	"github.com/aether-defense-system/service/trade/rpc"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CancelOrder cancels an order.
func (l *CancelOrderLogic) CancelOrder(in *rpc.CancelOrderRequest) (*rpc.CancelOrderResponse, error) {
	// todo: add your logic here and delete this line

	return &rpc.CancelOrderResponse{}, nil
}
