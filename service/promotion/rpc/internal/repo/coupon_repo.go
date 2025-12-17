// Package repo provides data access layer for promotion service.
package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aether-defense-system/common/database"
)

// CouponRepo provides data access operations for coupon domain.
type CouponRepo struct {
	db *sql.DB
}

// NewCouponRepo creates a new CouponRepo instance.
func NewCouponRepo(db *sql.DB) *CouponRepo {
	return &CouponRepo{db: db}
}

// Create creates a new coupon record.
func (r *CouponRepo) Create(ctx context.Context, coupon *database.PromotionCouponRecord) error {
	query := `INSERT INTO promotion_coupon_record
	          (id, user_id, template_id, status, order_id)
	          VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		coupon.ID, coupon.UserID, coupon.TemplateID, coupon.Status, coupon.OrderID)
	if err != nil {
		return fmt.Errorf("failed to create coupon record: %w", err)
	}

	return nil
}

// GetByID retrieves a coupon record by ID.
func (r *CouponRepo) GetByID(ctx context.Context, couponID int64) (*database.PromotionCouponRecord, error) {
	query := `SELECT id, user_id, template_id, status, use_time, order_id, create_time, update_time
	          FROM promotion_coupon_record WHERE id = ?`

	var coupon database.PromotionCouponRecord
	err := r.db.QueryRowContext(ctx, query, couponID).
		Scan(&coupon.ID, &coupon.UserID, &coupon.TemplateID, &coupon.Status,
			&coupon.UseTime, &coupon.OrderID, &coupon.CreateTime, &coupon.UpdateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("coupon record not found: %d", couponID)
		}
		return nil, fmt.Errorf("failed to get coupon record: %w", err)
	}

	return &coupon, nil
}

// GetByUserIDAndTemplateID retrieves a coupon record by user ID and template ID.
func (r *CouponRepo) GetByUserIDAndTemplateID(
	ctx context.Context, userID, templateID int64,
) (*database.PromotionCouponRecord, error) {
	query := `SELECT id, user_id, template_id, status, use_time, order_id, create_time, update_time
	          FROM promotion_coupon_record WHERE user_id = ? AND template_id = ?`

	var coupon database.PromotionCouponRecord
	err := r.db.QueryRowContext(ctx, query, userID, templateID).
		Scan(&coupon.ID, &coupon.UserID, &coupon.TemplateID, &coupon.Status,
			&coupon.UseTime, &coupon.OrderID, &coupon.CreateTime, &coupon.UpdateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("coupon record not found: user_id=%d, template_id=%d", userID, templateID)
		}
		return nil, fmt.Errorf("failed to get coupon record: %w", err)
	}

	return &coupon, nil
}

// GetByUserID retrieves coupon records by user ID with status filter.
func (r *CouponRepo) GetByUserID(
	ctx context.Context, userID int64, status *int8, limit, offset int,
) ([]*database.PromotionCouponRecord, error) {
	query := `SELECT id, user_id, template_id, status, use_time, order_id, create_time, update_time
	          FROM promotion_coupon_record WHERE user_id = ?`
	args := []interface{}{userID}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query coupon records: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	var coupons []*database.PromotionCouponRecord
	for rows.Next() {
		var coupon database.PromotionCouponRecord
		scanErr := rows.Scan(&coupon.ID, &coupon.UserID, &coupon.TemplateID, &coupon.Status,
			&coupon.UseTime, &coupon.OrderID, &coupon.CreateTime, &coupon.UpdateTime)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan coupon record: %w", scanErr)
		}
		coupons = append(coupons, &coupon)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating coupon records: %w", err)
	}

	return coupons, nil
}

// UpdateStatus updates coupon status.
func (r *CouponRepo) UpdateStatus(
	ctx context.Context, couponID int64, oldStatus, newStatus int8, orderID *int64,
) error {
	query := `UPDATE promotion_coupon_record SET status = ?, use_time = NOW(), order_id = ?
	          WHERE id = ? AND status = ?`

	result, err := r.db.ExecContext(ctx, query, newStatus, orderID, couponID, oldStatus)
	if err != nil {
		return fmt.Errorf("failed to update coupon status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"coupon status update failed: coupon not found or status mismatch "+
				"(id=%d, expected_status=%d)",
			couponID, oldStatus)
	}

	return nil
}
