// Package rpc implements the user RPC service.
package rpc

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

// UserService user service struct.
type UserService struct{}

// GetUser get user information.
func (s *UserService) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	// TODO: Implement user information retrieval logic
	// 1. Query database to get user information
	// 2. Return user information

	logx.WithContext(ctx).Infof("Getting information for user ID=%d", req.UserId)

	return &GetUserResponse{
		UserId:   req.UserId,
		Username: "testuser",
		Mobile:   "13800138000",
	}, nil
}
