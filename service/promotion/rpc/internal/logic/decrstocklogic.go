// Package logic contains business logic implementations for promotion service.
package logic

import (
	"context"
	"fmt"

	"github.com/aether-defense-system/service/promotion/rpc"
	"github.com/aether-defense-system/service/promotion/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// DecrStockLogic handles inventory deduction logic.
type DecrStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewDecrStockLogic creates a new DecrStockLogic instance.
func NewDecrStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DecrStockLogic {
	return &DecrStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DecrStock decrements inventory.
//
// Responsibilities:
//   - Validate the request (course ID and quantity)
//   - Delegate to inventory layer (via svcCtx) to perform atomic stock deduction
//   - Return a clear success/failure result
//
// Currently we implement validation and logging; inventory interaction will be
// wired via svcCtx in later iterations.
func (l *DecrStockLogic) DecrStock(req *rpc.DecrStockRequest) (*rpc.DecrStockResponse, error) {
	if req == nil {
		l.Errorf("received nil DecrStockRequest")
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.CourseId <= 0 {
		l.Errorf("invalid course_id: %d", req.CourseId)
		return nil, fmt.Errorf("invalid course_id: %d", req.CourseId)
	}

	if req.Num <= 0 {
		l.Errorf("invalid num: %d for course_id: %d", req.Num, req.CourseId)
		return nil, fmt.Errorf("num must be greater than 0")
	}

	l.Infof("decrementing stock: courseId=%d, num=%d", req.CourseId, req.Num)

	// TODO: call inventory repository / Redis Lua via svcCtx.

	return &rpc.DecrStockResponse{
		Success: true,
		Message: "Inventory deduction successful",
	}, nil
}
