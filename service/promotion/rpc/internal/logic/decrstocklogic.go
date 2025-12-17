// Package logic contains business logic implementations for promotion service.
package logic

import (
	"context"
	"fmt"

	"github.com/aether-defense-system/service/promotion/rpc"
	"github.com/aether-defense-system/service/promotion/rpc/svc"

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

	// Check if Redis client is available
	if l.svcCtx.Redis == nil {
		l.Errorf("Redis client not initialized")
		return nil, fmt.Errorf("redis client not available")
	}

	// Generate inventory key for the course
	inventoryKey := fmt.Sprintf("inventory:course:%d", req.CourseId)

	// Check current stock before deduction (for debugging)
	currentStock, getErr := l.svcCtx.Redis.Get(l.ctx, inventoryKey)
	if getErr != nil {
		l.Infof("before deduction: courseId=%d, currentStock=<err:%v>, num=%d", req.CourseId, getErr, req.Num)
	} else {
		l.Infof("before deduction: courseId=%d, currentStock=%s, num=%d", req.CourseId, currentStock, req.Num)
	}

	// Perform atomic inventory deduction using Redis Lua script
	err := l.svcCtx.Redis.DecrStock(l.ctx, inventoryKey, int64(req.Num))
	if err != nil {
		l.Errorf("failed to decrement stock: %v, courseId=%d, num=%d", err, req.CourseId, req.Num)
		return &rpc.DecrStockResponse{
			Success: false,
			Message: fmt.Sprintf("Inventory deduction failed: %v", err),
		}, nil
	}

	// Check stock after deduction (for debugging)
	afterStock, getAfterErr := l.svcCtx.Redis.Get(l.ctx, inventoryKey)
	if getAfterErr != nil {
		l.Infof("after deduction: courseId=%d, afterStock=<err:%v>, num=%d", req.CourseId, getAfterErr, req.Num)
	} else {
		l.Infof("after deduction: courseId=%d, afterStock=%s, num=%d", req.CourseId, afterStock, req.Num)
	}

	l.Infof("successfully decremented stock: courseId=%d, num=%d", req.CourseId, req.Num)

	return &rpc.DecrStockResponse{
		Success: true,
		Message: "Inventory deduction successful",
	}, nil
}
