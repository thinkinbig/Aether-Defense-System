// Package mq provides message queue client wrappers for RocketMQ.
package mq

import (
	"context"
	"fmt"
	"strings"
	"time"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/zeromicro/go-zero/core/logx"
)

// TransactionProducer wraps RocketMQ transaction producer with check-back support.
type TransactionProducer struct {
	producer rocketmq.TransactionProducer
	config   *Config
}

// LocalTransactionState represents the result of local transaction execution.
type LocalTransactionState int

const (
	// CommitMessageState indicates the local transaction succeeded and the message should be committed.
	CommitMessageState LocalTransactionState = iota
	// RollbackMessageState indicates the local transaction failed and the message should be rolled back.
	RollbackMessageState
	// UnknownState indicates the local transaction state is unknown and needs check-back.
	UnknownState
)

// LocalTransactionExecutor executes a local transaction and returns the state.
// This function is called by RocketMQ after sending a half message.
type LocalTransactionExecutor func(ctx context.Context, msg *primitive.MessageExt) (LocalTransactionState, error)

// CheckBackExecutor checks the status of a local transaction.
// This function is called by RocketMQ when the transaction state is unknown.
// It should be stateless and efficient (query database, not perform heavy operations).
type CheckBackExecutor func(ctx context.Context, msg *primitive.MessageExt) (LocalTransactionState, error)

// transactionListener implements primitive.TransactionListener interface.
type transactionListener struct {
	executor  LocalTransactionExecutor
	checkBack CheckBackExecutor
}

// ExecuteLocalTransaction executes the local transaction.
// Note: RocketMQ API doesn't provide context, so we use context.Background().
func (tl *transactionListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	ctx := context.Background()
	newMsg := primitive.NewMessage(msg.Topic, msg.Body)
	msgExt := &primitive.MessageExt{
		Message: *newMsg,
	}

	state, err := tl.executor(ctx, msgExt)
	if err != nil {
		logx.Errorf("local transaction execution failed: %v", err)
		return primitive.RollbackMessageState
	}

	switch state {
	case CommitMessageState:
		return primitive.CommitMessageState
	case RollbackMessageState:
		return primitive.RollbackMessageState
	case UnknownState:
		return primitive.UnknowState
	default:
		logx.Errorf("unknown transaction state: %d", state)
		return primitive.RollbackMessageState
	}
}

// CheckLocalTransaction checks the status of a local transaction.
// Note: RocketMQ API doesn't provide context, so we use context.Background().
func (tl *transactionListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	ctx := context.Background()
	state, err := tl.checkBack(ctx, msg)
	if err != nil {
		logx.Errorf("check-back execution failed: %v", err)
		return primitive.RollbackMessageState
	}

	switch state {
	case CommitMessageState:
		return primitive.CommitMessageState
	case RollbackMessageState:
		return primitive.RollbackMessageState
	case UnknownState:
		return primitive.UnknowState
	default:
		logx.Errorf("unknown transaction state: %d", state)
		return primitive.RollbackMessageState
	}
}

// NewTransactionProducer creates a new RocketMQ transaction producer.
func NewTransactionProducer(
	cfg *Config,
	executor LocalTransactionExecutor,
	checkBack CheckBackExecutor,
) (*TransactionProducer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("mq config cannot be nil")
	}
	if cfg.NameServer == "" {
		return nil, fmt.Errorf("nameServer is required")
	}
	if cfg.Group == "" {
		return nil, fmt.Errorf("group is required")
	}
	if executor == nil {
		return nil, fmt.Errorf("local transaction executor cannot be nil")
	}
	if checkBack == nil {
		return nil, fmt.Errorf("check-back executor cannot be nil")
	}

	// Create TransactionListener that implements both ExecuteLocalTransaction and CheckLocalTransaction
	listener := &transactionListener{
		executor:  executor,
		checkBack: checkBack,
	}

	// Parse NameServer addresses (support comma or semicolon separated)
	nameServers := parseNameServers(cfg.NameServer)
	if len(nameServers) == 0 {
		return nil, fmt.Errorf("invalid nameServer configuration: %s", cfg.NameServer)
	}

	opts := []producer.Option{
		producer.WithNameServer(nameServers),
		producer.WithGroupName(cfg.Group),
		producer.WithRetry(cfg.getRetryTimes()),
		producer.WithSendMsgTimeout(time.Duration(cfg.getSendTimeout()) * time.Millisecond),
	}

	p, err := rocketmq.NewTransactionProducer(
		listener,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction producer: %w", err)
	}

	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("failed to start transaction producer: %w", err)
	}

	return &TransactionProducer{
		producer: p,
		config:   cfg,
	}, nil
}

// SendMessageInTransaction sends a transactional message.
func (tp *TransactionProducer) SendMessageInTransaction(
	ctx context.Context,
	msg *primitive.Message,
) (*primitive.SendResult, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	// Set topic if not set
	if msg.Topic == "" {
		msg.Topic = tp.config.Topic
	}

	result, err := tp.producer.SendMessageInTransaction(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send transactional message: %w", err)
	}

	// Convert TransactionSendResult to SendResult
	if result == nil {
		return nil, fmt.Errorf("transaction send result is nil")
	}

	sendResult := &primitive.SendResult{
		Status:        result.Status,
		MessageQueue:  result.MessageQueue,
		MsgID:         result.MsgID,
		OffsetMsgID:   result.OffsetMsgID,
		QueueOffset:   result.QueueOffset,
		TransactionID: result.TransactionID,
	}

	return sendResult, nil
}

// Shutdown gracefully shuts down the producer.
func (tp *TransactionProducer) Shutdown() error {
	if tp.producer != nil {
		return tp.producer.Shutdown()
	}
	return nil
}

// getRetryTimes returns the retry times, defaulting to 2.
func (c *Config) getRetryTimes() int {
	if c.RetryTimes <= 0 {
		return 2
	}
	return c.RetryTimes
}

// getSendTimeout returns the send timeout in milliseconds, defaulting to 3000.
func (c *Config) getSendTimeout() int {
	if c.SendTimeout <= 0 {
		return 3000
	}
	return c.SendTimeout
}

// parseNameServers parses NameServer string into a slice of addresses.
// Supports comma or semicolon separated addresses.
func parseNameServers(nameServer string) []string {
	if nameServer == "" {
		return nil
	}

	// Try semicolon first (RocketMQ standard), then comma
	separator := ";"
	if !strings.Contains(nameServer, ";") && strings.Contains(nameServer, ",") {
		separator = ","
	}

	addresses := strings.Split(nameServer, separator)
	result := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			result = append(result, addr)
		}
	}
	return result
}
