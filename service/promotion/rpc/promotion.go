// Package rpc implements the promotion RPC service.
package rpc

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// PromotionService marketing service struct.
// This is a placeholder implementation. The actual implementation with ServiceContext
// is in internal/server/promotionserviceserver.go, which is used when properly initialized.
type PromotionService struct {
	UnimplementedPromotionServiceServer
}

// DecrStock decrement inventory.
// This is a placeholder. Use internal/server.PromotionServiceServer for actual implementation.
func (s *PromotionService) DecrStock(ctx context.Context, _ *DecrStockRequest) (*DecrStockResponse, error) {
	logx.WithContext(ctx).Errorf("PromotionService.DecrStock: service not properly initialized")
	return &DecrStockResponse{
		Success: false,
		Message: "Service not properly initialized. Use server.PromotionServiceServer instead.",
	}, nil
}
