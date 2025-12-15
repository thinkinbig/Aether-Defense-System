// Package logic contains business logic implementations for trade service.
package logic

import (
	"context"

	"github.com/aether-defense-system/service/trade/rpc"
	"github.com/aether-defense-system/service/trade/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlaceOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceOrderLogic {
	return &PlaceOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PlaceOrder places an order.
func (l *PlaceOrderLogic) PlaceOrder(in *rpc.PlaceOrderRequest) (*rpc.PlaceOrderResponse, error) {
	// todo: add your logic here and delete this line

	return &rpc.PlaceOrderResponse{}, nil
}
