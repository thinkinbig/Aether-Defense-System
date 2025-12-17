// Package repo provides data access layer for trade service.
package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aether-defense-system/common/database"
)

// OrderRepo provides data access operations for order domain.
type OrderRepo struct {
	db *sql.DB
}

// NewOrderRepo creates a new OrderRepo instance.
func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

// CreateOrder creates a new order with items in a transaction.
func (r *OrderRepo) CreateOrder(
	ctx context.Context,
	order *database.TradeOrder,
	items []*database.TradeOrderItem,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error but don't override original error
				_ = rollbackErr
			}
		}
	}()

	// Insert order
	orderQuery := `INSERT INTO trade_order
	               (id, user_id, status, total_amount, pay_amount, pay_channel, version)
	               VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, orderQuery,
		order.ID, order.UserID, order.Status, order.TotalAmount, order.PayAmount,
		order.PayChannel, order.Version)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Insert order items
	itemQuery := `INSERT INTO trade_order_item
	              (id, order_id, user_id, course_id, course_name, price, real_pay_amount)
	              VALUES (?, ?, ?, ?, ?, ?, ?)`

	for _, item := range items {
		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID, item.OrderID, item.UserID, item.CourseID, item.CourseName,
			item.Price, item.RealPayAmount)
		if err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves an order by ID.
func (r *OrderRepo) GetByID(ctx context.Context, orderID int64) (*database.TradeOrder, error) {
	query := `SELECT id, user_id, status, total_amount, pay_amount, pay_channel,
	                 out_trade_no, pay_time, create_time, update_time, version
	          FROM trade_order WHERE id = ?`

	var order database.TradeOrder
	err := r.db.QueryRowContext(ctx, query, orderID).
		Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount, &order.PayAmount,
			&order.PayChannel, &order.OutTradeNo, &order.PayTime,
			&order.CreateTime, &order.UpdateTime, &order.Version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %d", orderID)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// GetByUserID retrieves orders by user ID with pagination.
func (r *OrderRepo) GetByUserID(
	ctx context.Context, userID int64, status *int8, limit, offset int,
) ([]*database.TradeOrder, error) {
	query := `SELECT id, user_id, status, total_amount, pay_amount, pay_channel,
	                 out_trade_no, pay_time, create_time, update_time, version
	          FROM trade_order WHERE user_id = ?`
	args := []interface{}{userID}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	var orders []*database.TradeOrder
	for rows.Next() {
		var order database.TradeOrder
		scanErr := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
			&order.PayAmount, &order.PayChannel, &order.OutTradeNo,
			&order.PayTime, &order.CreateTime, &order.UpdateTime, &order.Version)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan order: %w", scanErr)
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

// GetItemsByOrderID retrieves order items by order ID.
func (r *OrderRepo) GetItemsByOrderID(ctx context.Context, orderID int64) ([]*database.TradeOrderItem, error) {
	query := `SELECT id, order_id, user_id, course_id, course_name, price, real_pay_amount,
	                 create_time, update_time
	          FROM trade_order_item WHERE order_id = ?`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	var items []*database.TradeOrderItem
	for rows.Next() {
		var item database.TradeOrderItem
		scanErr := rows.Scan(&item.ID, &item.OrderID, &item.UserID, &item.CourseID,
			&item.CourseName, &item.Price, &item.RealPayAmount,
			&item.CreateTime, &item.UpdateTime)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", scanErr)
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return items, nil
}

// UpdateStatus updates order status with optimistic lock.
func (r *OrderRepo) UpdateStatus(
	ctx context.Context, orderID int64, oldStatus, newStatus int8, version int32,
) error {
	query := `UPDATE trade_order SET status = ?, version = version + 1
	          WHERE id = ? AND status = ? AND version = ?`

	result, err := r.db.ExecContext(ctx, query, newStatus, orderID, oldStatus, version)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"order status update failed: order not found or version mismatch "+
				"(id=%d, expected_status=%d, expected_version=%d)",
			orderID, oldStatus, version)
	}

	return nil
}

// UpdatePayInfo updates payment information.
func (r *OrderRepo) UpdatePayInfo(
	ctx context.Context, orderID int64, payChannel int8, outTradeNo string, payTime interface{},
) error {
	query := `UPDATE trade_order SET pay_channel = ?, out_trade_no = ?, pay_time = ?, status = ?
	          WHERE id = ? AND status = ?`

	result, err := r.db.ExecContext(ctx, query, payChannel, outTradeNo, payTime,
		database.OrderStatusPaid, orderID, database.OrderStatusPendingPayment)
	if err != nil {
		return fmt.Errorf("failed to update pay info: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pay info update failed: order not found or status mismatch (id=%d)", orderID)
	}

	return nil
}
