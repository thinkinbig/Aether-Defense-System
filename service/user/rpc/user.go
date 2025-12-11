package rpc

import (
	"context"
	"fmt"
)

// UserService user service struct
type UserService struct{}

// GetUser get user information
func (s *UserService) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	// TODO: Implement user information retrieval logic
	// 1. Query database to get user information
	// 2. Return user information
	
	fmt.Printf("Getting information for user ID=%d\n", req.UserId)
	
	return &GetUserResponse{
		UserId:   req.UserId,
		Username: "testuser",
		Mobile:   "13800138000",
	}, nil
}