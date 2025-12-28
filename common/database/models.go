// Package database provides data models for the Aether Defense System.
package database

import "time"

// TradeOrder represents the trade_order table.
//
//nolint:govet // Field order optimized for logical grouping
type TradeOrder struct {
	ID          int64      `db:"id"`
	UserID      int64      `db:"user_id"`
	Status      int8       `db:"status"`       // OrderStatus: 1=PendingPayment, 2=Closed, 3=Paid, 4=Finished, 5=Refunded
	TotalAmount int32      `db:"total_amount"` // Amount in cents
	PayAmount   int32      `db:"pay_amount"`   // Amount in cents
	PayChannel  *int8      `db:"pay_channel"`  // PayChannel: 1=Alipay, 2=WeChat
	OutTradeNo  *string    `db:"out_trade_no"`
	PayTime     *time.Time `db:"pay_time"`
	CreateTime  time.Time  `db:"create_time"`
	UpdateTime  time.Time  `db:"update_time"`
	Version     int32      `db:"version"` // Optimistic lock version
}

// TradeOrderItem represents the trade_order_item table.
//
//nolint:govet // Field order optimized for logical grouping
type TradeOrderItem struct {
	ID            int64     `db:"id"`
	OrderID       int64     `db:"order_id"`
	UserID        int64     `db:"user_id"`
	CourseID      int64     `db:"course_id"`
	CourseName    string    `db:"course_name"`
	Price         int32     `db:"price"`           // Amount in cents
	RealPayAmount int32     `db:"real_pay_amount"` // Amount in cents
	CreateTime    time.Time `db:"create_time"`
	UpdateTime    time.Time `db:"update_time"`
}

// PromotionCouponRecord represents the promotion_coupon_record table.
//
//nolint:govet // Field order optimized for logical grouping
type PromotionCouponRecord struct {
	ID         int64      `db:"id"`
	UserID     int64      `db:"user_id"`
	TemplateID int64      `db:"template_id"`
	Status     int8       `db:"status"` // CouponStatus: 1=Unused, 2=Used, 3=Expired
	UseTime    *time.Time `db:"use_time"`
	OrderID    *int64     `db:"order_id"`
	CreateTime time.Time  `db:"create_time"`
	UpdateTime time.Time  `db:"update_time"`
}

// User represents the user table.
//
//nolint:govet // Field order optimized for logical grouping
type User struct {
	ID         int64     `db:"id"`
	Username   string    `db:"username"`
	Mobile     string    `db:"mobile"`
	Email      *string   `db:"email"`
	Avatar     *string   `db:"avatar"`
	Status     int8      `db:"status"` // UserStatus: 1=Normal, 2=Banned
	CreateTime time.Time `db:"create_time"`
	UpdateTime time.Time `db:"update_time"`
}

// OrderStatus constants.
const (
	OrderStatusPendingPayment = 1 // Pending payment
	OrderStatusClosed         = 2 // Closed
	OrderStatusPaid           = 3 // Paid
	OrderStatusFinished       = 4 // Finished
	OrderStatusRefunded       = 5 // Refunded
)

// CouponStatus constants.
const (
	CouponStatusUnused  = 1 // Unused
	CouponStatusUsed    = 2 // Used
	CouponStatusExpired = 3 // Expired
)

// UserStatus constants.
const (
	UserStatusNormal = 1 // Normal
	UserStatusBanned = 2 // Banned
)

// PayChannel constants.
const (
	PayChannelAlipay = 1 // Alipay
	PayChannelWeChat = 2 // WeChat
)
