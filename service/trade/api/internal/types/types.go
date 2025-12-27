// Package types defines request/response types for trade-api.
package types

// PlaceOrderReq represents the HTTP request to place an order.
type PlaceOrderReq struct {
	CourseIDs []int64 `json:"courseIds"`           // Purchased course list
	CouponIDs []int64 `json:"couponIds,omitempty"` // Selected coupon IDs, optional
	OrderID   int64   `json:"orderId"`             // Pre-generated order ID
}

// PlaceOrderResp represents the HTTP response for placing an order.
type PlaceOrderResp struct {
	OrderID   int64 `json:"orderId"`   // Returned order number
	PayAmount int   `json:"payAmount"` // Actual payment amount (in cents)
	Status    int   `json:"status"`    // Order status (1: Pending Payment)
}
