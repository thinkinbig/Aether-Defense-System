// Package logic contains business logic implementations for trade service.
package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/aether-defense-system/common/database"
	"github.com/aether-defense-system/common/mq"
	"github.com/aether-defense-system/service/trade/rpc"
	tradesvc "github.com/aether-defense-system/service/trade/rpc/internal/svc"
	userservice "github.com/aether-defense-system/service/user/rpc/userservice"
)

// PlaceOrderLogic handles order placement logic.
type PlaceOrderLogic struct {
	ctx    context.Context
	svcCtx *tradesvc.ServiceContext
	logx.Logger
}

// NewPlaceOrderLogic creates a new PlaceOrderLogic instance.
func NewPlaceOrderLogic(ctx context.Context, svcCtx *tradesvc.ServiceContext) *PlaceOrderLogic {
	return &PlaceOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// OrderMessage represents the message sent to RocketMQ for inventory deduction.
type OrderMessage struct {
	CourseIDs  []int64 `json:"courseIds"`
	OrderID    int64   `json:"orderId"`
	UserID     int64   `json:"userId"`
	RealAmount int32   `json:"realAmount"`
}

// PlaceOrder places an order.
//
// This method implements the complete order placement flow:
//   - Validates user exists
//   - Creates order in database
//   - Sends RocketMQ transactional message for inventory deduction
//   - Returns order response
func (l *PlaceOrderLogic) PlaceOrder(req *rpc.PlaceOrderRequest) (*rpc.PlaceOrderResponse, error) {
	// Parameter validation (business rules)
	if req == nil {
		l.Errorf("received nil PlaceOrderRequest")
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.UserId <= 0 {
		l.Errorf("invalid user_id: %d", req.UserId)
		return nil, fmt.Errorf("invalid user_id: %d", req.UserId)
	}

	if req.OrderId <= 0 {
		l.Errorf("invalid order_id: %d", req.OrderId)
		return nil, fmt.Errorf("invalid order_id: %d", req.OrderId)
	}

	if len(req.CourseIds) == 0 {
		l.Errorf("empty course_ids for user_id: %d", req.UserId)
		return nil, fmt.Errorf("course_ids cannot be empty")
	}

	if req.RealAmount <= 0 {
		l.Errorf("invalid real_amount: %d for order_id: %d", req.RealAmount, req.OrderId)
		return nil, fmt.Errorf("real_amount must be greater than 0")
	}

	// Validate user exists (mandatory)
	if l.svcCtx.UserRPC == nil {
		l.Errorf("user RPC client not initialized")
		return nil, fmt.Errorf("user validation service not available")
	}

	_, err := l.svcCtx.UserRPC.GetUser(l.ctx, &userservice.GetUserRequest{UserId: req.UserId})
	if err != nil {
		l.Errorf("user validation failed: %v, user_id=%d", err, req.UserId)
		return nil, fmt.Errorf("user not found or invalid: %w", err)
	}
	l.Infof("user validated: userId=%d", req.UserId)

	// Create RocketMQ transaction producer if not already created
	// Note: The producer is reused across transactions, so the local executor
	// must reconstruct order data from the message
	if l.svcCtx.RocketMQ == nil {
		// Local transaction executor: creates order in database
		localExecutor := func(ctx context.Context, msg *primitive.MessageExt) (mq.LocalTransactionState, error) {
			// Parse message to get order info
			var orderMsg OrderMessage
			if parseErr := json.Unmarshal(msg.Body, &orderMsg); parseErr != nil {
				l.Errorf("failed to parse order message: %v", parseErr)
				return mq.RollbackMessageState, parseErr
			}

			// Reconstruct order and items from message
			// In production, you might want to include more details in the message
			// or fetch course prices from a service
			order := &database.TradeOrder{
				ID:          orderMsg.OrderID,
				UserID:      orderMsg.UserID,
				Status:      database.OrderStatusPendingPayment,
				TotalAmount: orderMsg.RealAmount,
				PayAmount:   orderMsg.RealAmount,
				CreateTime:  time.Now(),
				UpdateTime:  time.Now(),
				Version:     1,
			}

			courseCount := len(orderMsg.CourseIDs)
			if courseCount == 0 {
				l.Errorf("empty course list in order message: orderId=%d", orderMsg.OrderID)
				return mq.RollbackMessageState, fmt.Errorf("course list cannot be empty")
			}
			if courseCount > 2147483647 { // Max int32 value
				l.Errorf("course count exceeds int32 limit: orderId=%d, count=%d", orderMsg.OrderID, courseCount)
				return mq.RollbackMessageState, fmt.Errorf("course count too large")
			}
			courseCount32 := int32(courseCount)
			orderItems := make([]*database.TradeOrderItem, 0, courseCount)
			for i, courseID := range orderMsg.CourseIDs {
				pricePerCourse := orderMsg.RealAmount / courseCount32
				if i == courseCount-1 {
					pricePerCourse = orderMsg.RealAmount - (pricePerCourse * (courseCount32 - 1))
				}
				item := &database.TradeOrderItem{
					ID:            orderMsg.OrderID + int64(i+1),
					OrderID:       orderMsg.OrderID,
					UserID:        orderMsg.UserID,
					CourseID:      courseID,
					CourseName:    fmt.Sprintf("Course %d", courseID),
					Price:         pricePerCourse,
					RealPayAmount: pricePerCourse,
					CreateTime:    time.Now(),
					UpdateTime:    time.Now(),
				}
				orderItems = append(orderItems, item)
			}

			// Create order in database
			if l.svcCtx.OrderRepo == nil {
				l.Errorf("order repository not initialized")
				return mq.RollbackMessageState, fmt.Errorf("order repository not available")
			}

			if createErr := l.svcCtx.OrderRepo.CreateOrder(ctx, order, orderItems); createErr != nil {
				l.Errorf("failed to create order in local transaction: %v, orderId=%d", createErr, orderMsg.OrderID)
				return mq.RollbackMessageState, createErr
			}

			l.Infof("order created successfully in local transaction: orderId=%d", orderMsg.OrderID)
			return mq.CommitMessageState, nil
		}

		// Check-back executor: verifies if order exists in database
		checkBack := func(ctx context.Context, msg *primitive.MessageExt) (mq.LocalTransactionState, error) {
			var orderMsg OrderMessage
			if parseErr := json.Unmarshal(msg.Body, &orderMsg); parseErr != nil {
				l.Errorf("failed to parse order message in check-back: %v", parseErr)
				return mq.RollbackMessageState, parseErr
			}

			if l.svcCtx.OrderRepo == nil {
				return mq.RollbackMessageState, fmt.Errorf("order repository not available")
			}

			// Check if order exists
			order, getErr := l.svcCtx.OrderRepo.GetByID(ctx, orderMsg.OrderID)
			if getErr != nil {
				l.Infof("order not found in check-back: orderId=%d, error=%v", orderMsg.OrderID, getErr)
				return mq.RollbackMessageState, nil // Order doesn't exist, rollback
			}

			// Order exists, commit
			l.Infof("order found in check-back: orderId=%d, status=%d", orderMsg.OrderID, order.Status)
			return mq.CommitMessageState, nil
		}

		producer, producerErr := mq.NewTransactionProducer(&l.svcCtx.Config.RocketMQ, localExecutor, checkBack)
		if producerErr != nil {
			l.Errorf("failed to create RocketMQ transaction producer: %v", producerErr)
			return nil, fmt.Errorf("failed to initialize message queue: %w", producerErr)
		}
		l.svcCtx.RocketMQ = producer
	}

	// Prepare message for RocketMQ
	orderMsg := OrderMessage{
		OrderID:    req.OrderId,
		UserID:     req.UserId,
		CourseIDs:  req.CourseIds,
		RealAmount: req.RealAmount,
	}
	msgBody, err := json.Marshal(orderMsg)
	if err != nil {
		l.Errorf("failed to marshal order message: %v", err)
		return nil, fmt.Errorf("failed to prepare message: %w", err)
	}

	// Create RocketMQ message
	msg := primitive.NewMessage(l.svcCtx.Config.RocketMQ.Topic, msgBody)
	msg.WithKeys([]string{fmt.Sprintf("order_%d", req.OrderId)})
	msg.WithTag("ORDER_PLACED")

	// Send transactional message
	// This will trigger the local transaction executor which creates the order
	result, err := l.svcCtx.RocketMQ.SendMessageInTransaction(l.ctx, msg)
	if err != nil {
		l.Errorf("failed to send transactional message: %v, orderId=%d", err, req.OrderId)
		return nil, fmt.Errorf("failed to send order message: %w", err)
	}

	l.Infof("transactional message sent: orderId=%d, msgId=%s, status=%s",
		req.OrderId, result.MsgID, result.Status)

	// Note: The order is created in the local transaction executor.
	// If the message is committed, the order exists. If rolled back, it doesn't.
	// We return success here, but the actual order creation happens asynchronously.
	// In a production system, you might want to wait for the transaction to complete
	// or use a different pattern.

	return &rpc.PlaceOrderResponse{
		OrderId:   req.OrderId,
		PayAmount: req.RealAmount,
		Status:    database.OrderStatusPendingPayment,
	}, nil
}
