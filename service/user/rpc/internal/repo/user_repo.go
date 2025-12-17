// Package repo provides data access layer for user service.
package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aether-defense-system/common/database"
)

// UserRepo provides data access operations for user domain.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a new UserRepo instance.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// GetByID retrieves a user by ID.
func (r *UserRepo) GetByID(ctx context.Context, userID int64) (*database.User, error) {
	query := `SELECT id, username, mobile, email, avatar, status, create_time, update_time
	          FROM user WHERE id = ? AND status = ?`

	var user database.User
	err := r.db.QueryRowContext(ctx, query, userID, database.UserStatusNormal).
		Scan(&user.ID, &user.Username, &user.Mobile, &user.Email, &user.Avatar,
			&user.Status, &user.CreateTime, &user.UpdateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %d", userID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByMobile retrieves a user by mobile number.
func (r *UserRepo) GetByMobile(ctx context.Context, mobile string) (*database.User, error) {
	query := `SELECT id, username, mobile, email, avatar, status, create_time, update_time
	          FROM user WHERE mobile = ? AND status = ?`

	var user database.User
	err := r.db.QueryRowContext(ctx, query, mobile, database.UserStatusNormal).
		Scan(&user.ID, &user.Username, &user.Mobile, &user.Email, &user.Avatar,
			&user.Status, &user.CreateTime, &user.UpdateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: mobile=%s", mobile)
		}
		return nil, fmt.Errorf("failed to get user by mobile: %w", err)
	}

	return &user, nil
}

// Create creates a new user.
func (r *UserRepo) Create(ctx context.Context, user *database.User) error {
	query := `INSERT INTO user (id, username, mobile, email, avatar, status)
	          VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Mobile, user.Email, user.Avatar, user.Status)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Update updates user information.
func (r *UserRepo) Update(ctx context.Context, user *database.User) error {
	query := `UPDATE user SET username = ?, mobile = ?, email = ?, avatar = ?, status = ?
	          WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		user.Username, user.Mobile, user.Email, user.Avatar, user.Status, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %d", user.ID)
	}

	return nil
}
