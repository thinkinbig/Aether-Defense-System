package rpc

import (
	"context"
	"fmt"
)

// PromotionService marketing service struct
type PromotionService struct{}

// DecrStock decrement inventory
func (s *PromotionService) DecrStock(ctx context.Context, req *DecrStockRequest) (*DecrStockResponse, error) {
	// TODO: Implement inventory deduction logic
	// 1. Call Redis Lua script to perform atomic inventory deduction
	// 2. Return execution result
	
	fmt.Printf("Deducting inventory for course ID=%d, quantity=%d\n", req.CourseId, req.Num)
	
	return &DecrStockResponse{
		Success: true,
		Message: "Inventory deduction successful",
	}, nil
}