// Package types defines request/response types and errors for user-api.
package types

import "errors"

// GetUserRequest represents the HTTP request to fetch a user by ID.
type GetUserRequest struct {
	UserID int64 `path:"userId"`
}

// GetUserResponse represents the HTTP response for a user.
type GetUserResponse struct {
	Username string `json:"username"`
	Mobile   string `json:"mobile"`
	UserID   int64  `json:"userId"`
}

// Domain-level errors returned by the HTTP layer.
var (
	ErrInvalidUserID = errors.New("invalid user_id: must be greater than 0")
)
