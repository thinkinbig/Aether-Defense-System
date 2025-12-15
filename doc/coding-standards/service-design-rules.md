# 微服务设计规则

本文档定义了 Aether-Defense-System 项目中微服务的设计规则和最佳实践。

## 目录

- [服务边界](#服务边界)
- [服务间通信](#服务间通信)
- [数据管理](#数据管理)
- [服务治理](#服务治理)
- [部署规范](#部署规范)

## 服务边界

### 服务划分原则

按照业务领域（Domain）划分服务：

- **trade-rpc**: 交易域 - 订单、支付、退款
- **promotion-rpc**: 营销域 - 优惠券、秒杀、活动
- **user-rpc**: 用户域 - 用户信息、积分、会员

### 服务职责

每个服务应该：

1. **拥有独立的数据存储**：每个服务管理自己的数据库
2. **封装业务逻辑**：服务内部实现完整的业务功能
3. **提供清晰的接口**：通过 gRPC 暴露服务能力
4. **保持无状态**：服务实例可以水平扩展

### 禁止跨服务直接访问数据

❌ **错误做法**：

```go
// 在 trade-rpc 中直接访问 user 数据库
db.Query("SELECT * FROM users WHERE id = ?", userId)
```

✅ **正确做法**：

```go
// 通过 gRPC 调用 user-rpc
userClient := userrpc.NewUserClient(conn)
user, err := userClient.GetUser(ctx, &userrpc.GetUserRequest{UserId: userId})
```

## 服务间通信

### gRPC 通信规范

#### 1. 使用 gRPC 进行服务间调用

```go
import (
    "github.com/zeromicro/go-zero/zrpc"
    "github.com/aether-defense-system/service/user/rpc/userclient"
)

// 创建客户端
conn := zrpc.MustNewClient(zrpc.RpcClientConf{
    Etcd: discov.EtcdConf{
        Hosts: []string{"127.0.0.1:2379"},
        Key:   "user.rpc",
    },
}).Conn()

userClient := userclient.NewUserClient(conn)
```

#### 2. Context 传递

- **传递超时信息**：为每个 RPC 调用设置超时
- **传递请求 ID**：用于链路追踪
- **传递用户信息**：用于权限校验

```go
// 设置超时
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()

// 传递请求 ID
ctx = context.WithValue(ctx, "request_id", requestID)

// 调用 RPC
resp, err := userClient.GetUser(ctx, req)
```

#### 3. 错误处理

```go
resp, err := userClient.GetUser(ctx, req)
if err != nil {
    // 检查是否是 gRPC 错误
    if st, ok := status.FromError(err); ok {
        switch st.Code() {
        case codes.NotFound:
            return nil, errors.New(ErrCodeUserNotFound, "user not found")
        case codes.DeadlineExceeded:
            return nil, errors.New(ErrCodeTimeout, "service timeout")
        default:
            return nil, errors.New(ErrCodeInternal, "internal error")
        }
    }
    return nil, err
}
```

### 异步通信（消息队列）

#### 使用 RocketMQ 进行异步通信

```go
import (
    "github.com/apache/rocketmq-client-go/v2"
    "github.com/apache/rocketmq-client-go/v2/producer"
)

// 发送事务消息
p, _ := rocketmq.NewTransactionProducer(
    func(ctx context.Context, msg *primitive.MessageExt) (localTransactionState primitive.LocalTransactionState, err error) {
        // 执行本地事务
        if err := s.createOrder(ctx, order); err != nil {
            return primitive.RollbackMessageState, err
        }
        return primitive.CommitMessageState, nil
    },
    producer.WithNameServer([]string{"127.0.0.1:9876"}),
)

// 发送消息
msg := primitive.NewMessage("order-topic", []byte("order data"))
result, err := p.SendMessageInTransaction(ctx, msg)
```

## 数据管理

### 数据库设计

#### 1. 每个服务拥有独立的数据库

- **trade 服务**：`trade_db`
- **promotion 服务**：`promotion_db`
- **user 服务**：`user_db`

#### 2. 分库分表策略

- **分片键选择**：优先使用 `user_id`
- **分片算法**：Hash 取模 `user_id % 32`
- **绑定表**：父子表使用相同的分片键

#### 3. ID 生成

统一使用 Snowflake 算法：

```go
import "github.com/aether-defense-system/common/snowflake"

idGenerator := snowflake.NewGenerator(workerID)
orderId := idGenerator.Next()
```

### 数据一致性

#### 1. 本地事务

```go
// 在单个服务内使用数据库事务
tx, err := s.db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// 执行多个操作
if err := tx.CreateOrder(order); err != nil {
    return err
}
if err := tx.CreateOrderItems(items); err != nil {
    return err
}

// 提交事务
return tx.Commit()
```

#### 2. 分布式事务

使用 RocketMQ 事务消息实现最终一致性：

```go
// Phase 1: 发送半消息
halfMsg := primitive.NewMessage("order-topic", orderData)
result, err := producer.SendMessageInTransaction(ctx, halfMsg)

// Phase 2: 执行本地事务
if err := s.createOrder(ctx, order); err != nil {
    // Phase 3: 回滚
    return primitive.RollbackMessageState, err
}

// Phase 3: 提交
return primitive.CommitMessageState, nil
```

#### 3. 幂等性保证

```go
// 使用唯一索引防止重复
CREATE UNIQUE INDEX idx_user_coupon ON coupon_receive(user_id, template_id);

// 使用 Redis 去重
key := fmt.Sprintf("lock:order:%d", orderId)
exists, err := redis.SetNX(ctx, key, "1", 5*time.Minute)
if err != nil || !exists {
    return errors.New(ErrCodeDuplicate, "duplicate order")
}
```

## 服务治理

### 服务发现

使用 Etcd 作为服务注册中心：

```go
// 服务注册
zrpc.MustNewServer(c, func(grpcServer *grpc.Server) {
    trade.RegisterTradeServer(grpcServer, svcCtx)
}).Start()

// 服务发现
conn := zrpc.MustNewClient(zrpc.RpcClientConf{
    Etcd: discov.EtcdConf{
        Hosts: []string{"127.0.0.1:2379"},
        Key:   "trade.rpc",
    },
}).Conn()
```

### 负载均衡

Go-Zero 默认使用 P2C (Power of Two Choices) 算法：

- 随机选择两个节点
- 根据连接数和延迟选择负载更低的节点
- 避免"羊群效应"

### 限流

使用 Sentinel-Go：

```go
import "github.com/alibaba/sentinel-golang/api"

// 配置限流规则
_, err := flow.LoadRules([]*flow.Rule{
    {
        Resource: "trade:place_order",
        Threshold: 2500,  // QPS 阈值
        ControlBehavior: flow.WarmUp,
    },
})

// 在代码中使用
if entry, err := sentinel.Entry("trade:place_order"); err != nil {
    return nil, errors.New(ErrCodeRateLimit, "rate limit exceeded")
} else {
    defer entry.Exit()
    // 业务逻辑
}
```

### 熔断

```go
// 配置熔断规则
_, err := circuitbreaker.LoadRules([]*circuitbreaker.Rule{
    {
        Resource: "user:get_user",
        Strategy: circuitbreaker.SlowRequestRatio,
        RetryTimeoutMs: 5000,
        MinRequestAmount: 10,
        StatIntervalMs: 1000,
        MaxAllowedRtMs: 200,
        Threshold: 0.5,  // 慢调用比例 > 50%
    },
})
```

### 监控

暴露 Prometheus metrics：

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
        },
        []string{"method", "endpoint", "status"},
    )
)

func init() {
    prometheus.MustRegister(requestDuration)
}
```

## 部署规范

### 容器化

每个服务应该有独立的 Dockerfile：

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o trade-rpc cmd/rpc/trade-rpc/trade.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/trade-rpc .
CMD ["./trade-rpc"]
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: trade-rpc
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: trade-rpc
        image: trade-rpc:latest
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### 健康检查

```go
// 实现健康检查接口
func (s *TradeService) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
    // 检查数据库连接
    if err := s.db.Ping(); err != nil {
        return &HealthCheckResponse{Status: "unhealthy"}, nil
    }

    // 检查 Redis 连接
    if err := s.redis.Ping(ctx).Err(); err != nil {
        return &HealthCheckResponse{Status: "unhealthy"}, nil
    }

    return &HealthCheckResponse{Status: "healthy"}, nil
}
```

## 最佳实践总结

1. ✅ **服务边界清晰**：按业务领域划分，避免服务间耦合
2. ✅ **通信标准化**：使用 gRPC 进行同步调用，RocketMQ 进行异步通信
3. ✅ **数据独立**：每个服务拥有独立数据库，禁止跨服务直接访问
4. ✅ **最终一致性**：使用消息队列实现分布式事务
5. ✅ **服务治理**：实现限流、熔断、监控等治理能力
6. ✅ **可观测性**：日志、指标、链路追踪

## 参考资源

- [系统架构设计文档](../design/architecture.md)
- [性能优化指南](./performance-guidelines.md)
- [架构设计标准](../design/architecture-standards.md)
