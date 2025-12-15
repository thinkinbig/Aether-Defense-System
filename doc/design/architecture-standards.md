# 架构设计标准

本文档定义了 Aether-Defense-System 项目的架构设计标准，确保系统架构的一致性和可维护性。

## 目录

- [架构原则](#架构原则)
- [分层架构](#分层架构)
- [服务设计规范](#服务设计规范)
- [数据架构规范](#数据架构规范)
- [接口设计规范](#接口设计规范)
- [安全架构](#安全架构)

## 架构原则

### 1. 微服务原则

- **单一职责**：每个服务只负责一个业务领域
- **独立部署**：服务可以独立部署和扩展
- **数据独立**：每个服务拥有独立的数据库
- **技术无关**：服务间通过标准协议通信（gRPC）

### 2. 高可用原则

- **无状态设计**：服务实例无状态，可水平扩展
- **多活部署**：至少跨两个可用区部署
- **故障隔离**：服务故障不影响其他服务
- **优雅降级**：系统过载时提供有损服务

### 3. 性能原则

- **缓存优先**：读操作优先使用缓存
- **异步处理**：非关键路径使用异步处理
- **批量操作**：减少网络往返次数
- **连接复用**：复用数据库和 RPC 连接

### 4. 一致性原则

- **最终一致性**：接受最终一致性，使用消息队列
- **幂等性**：所有操作必须幂等
- **事务边界**：明确事务边界，避免分布式事务

## 分层架构

### 整体架构图

```
┌─────────────────────────────────────────────────┐
│              Client (Web/App/H5)                 │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│           API Gateway Layer                      │
│  - Authentication & Authorization               │
│  - Rate Limiting                                 │
│  - Request Routing                               │
└──────────┬───────────────────┬───────────────────┘
           │                   │
    ┌──────▼──────┐    ┌──────▼──────┐
    │  Trade API  │    │ Promotion   │
    │             │    │    API      │
    └──────┬──────┘    └──────┬──────┘
           │                   │
    ┌──────▼──────┐    ┌──────▼──────┐
    │  Trade RPC  │    │ Promotion  │
    │             │    │    RPC      │
    └──────┬──────┘    └──────┬──────┘
           │                   │
    ┌──────▼──────┐    ┌──────▼──────┐
    │   MySQL     │    │   Redis     │
    │  (Sharded)  │    │  (Cluster)  │
    └─────────────┘    └─────────────┘
           │                   │
           └───────┬───────────┘
                   │
            ┌──────▼──────┐
            │  RocketMQ   │
            │  (Message   │
            │   Queue)    │
            └─────────────┘
```

### 分层职责

#### 1. API Gateway 层

**位置**: `cmd/api/*`

**职责**:

- 协议转换（HTTP → gRPC）
- 请求聚合（调用多个 RPC 服务）
- 统一认证鉴权（JWT）
- 限流和熔断
- 请求路由

**规范**:

- 所有接口定义在 `.api` 文件中
- 使用 JWT 中间件进行认证
- 禁止在请求体中传递 `user_id`
- 统一错误响应格式

#### 2. RPC 服务层

**位置**: `service/*/rpc`

**职责**:

- 核心业务逻辑
- 领域服务实现
- 数据访问
- 服务间调用

**规范**:

- 服务间调用必须通过 gRPC
- 禁止跨服务直接访问数据库
- 使用 context 传递超时和取消信号
- 实现幂等性

#### 3. 数据层

**职责**:

- 数据持久化（MySQL）
- 缓存管理（Redis）
- 消息队列（RocketMQ）

**规范**:

- ID 生成使用 Snowflake 算法
- 分库分表策略（优先 `user_id`）
- Redis Key 命名规范
- 消息格式统一使用 Protobuf

## 服务设计规范

### 服务划分

按照业务领域（Domain）划分服务：

| 服务 | 领域 | 职责 |
|------|------|------|
| `trade-rpc` | 交易域 | 订单、支付、退款 |
| `promotion-rpc` | 营销域 | 优惠券、秒杀、活动 |
| `user-rpc` | 用户域 | 用户信息、积分、会员 |

### 服务接口设计

#### 1. gRPC 接口

```protobuf
syntax = "proto3";

package trade;

service TradeService {
  rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);
}

message PlaceOrderRequest {
  int64 user_id = 1;
  repeated int64 course_ids = 2;
  repeated int64 coupon_ids = 3;
}

message PlaceOrderResponse {
  int64 order_id = 1;
  int32 pay_amount = 2;
  int32 status = 3;
}
```

#### 2. HTTP 接口

```go
@server(
    group: order
    prefix: /v1/trade/order
    jwt: Auth
)
service trade-api {
    @doc "用户下单接口"
    @handler PlaceOrder
    post /place (PlaceOrderReq) returns (PlaceOrderResp)
}
```

### 服务间通信

#### 1. 同步调用（gRPC）

```go
// 创建客户端
conn := zrpc.MustNewClient(zrpc.RpcClientConf{
    Etcd: discov.EtcdConf{
        Hosts: []string{"127.0.0.1:2379"},
        Key:   "user.rpc",
    },
}).Conn()

userClient := userclient.NewUserClient(conn)

// 调用服务
resp, err := userClient.GetUser(ctx, &userrpc.GetUserRequest{
    UserId: userId,
})
```

#### 2. 异步调用（RocketMQ）

```go
// 发送事务消息
producer.SendMessageInTransaction(ctx, msg, func(ctx context.Context, msg *primitive.MessageExt) (primitive.LocalTransactionState, error) {
    // 执行本地事务
    if err := s.createOrder(ctx, order); err != nil {
        return primitive.RollbackMessageState, err
    }
    return primitive.CommitMessageState, nil
})
```

## 数据架构规范

### 数据库设计

#### 1. 分库分表策略

**分片键选择原则**:

- 优先使用 `user_id`（80% 查询按用户）
- 避免跨分片查询
- 使用绑定表策略（父子表相同分片键）

**分片算法**:

```go
shardIndex = user_id % 32  // 32 个分片
```

#### 2. ID 生成

**统一使用 Snowflake 算法**:

```go
import "github.com/aether-defense-system/common/snowflake"

idGenerator := snowflake.NewGenerator(workerID)
orderId := idGenerator.Next()
```

**Snowflake ID 结构**:

- 1 bit: 符号位（始终为 0）
- 41 bits: 时间戳（毫秒）
- 10 bits: 机器 ID（1024 个节点）
- 12 bits: 序列号（每毫秒 4096 个 ID）

#### 3. 数据一致性

**本地事务**:

```go
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// 执行操作
if err := tx.CreateOrder(order); err != nil {
    return err
}

return tx.Commit()
```

**分布式事务**:

- 使用 RocketMQ 事务消息
- 实现最终一致性
- 禁止使用 Seata（性能考虑）

### Redis 设计

#### 1. Key 命名规范

```
格式: {domain}:{resource}:{id}
示例:
  - promotion:stock:123        # 库存
  - promotion:coupon:456       # 优惠券
  - trade:order:789           # 订单
  - user:session:abc          # 会话
```

#### 2. 过期时间

**必须为所有 Key 设置过期时间**:

```go
// ✅ 正确
rdb.Set(ctx, "key", "value", 5*time.Minute)

// ❌ 错误
rdb.Set(ctx, "key", "value", 0)  // 永不过期
```

#### 3. Lua 脚本

**执行时间限制**: < 0.5ms

```lua
-- ✅ 正确：简单快速
local stock = redis.call('GET', KEYS[1])
if tonumber(stock) < tonumber(ARGV[1]) then
    return {err = "insufficient"}
end
redis.call('DECRBY', KEYS[1], ARGV[1])
return {ok = true}

-- ❌ 错误：复杂循环
for i = 1, 1000 do
    redis.call('GET', 'key' .. i)
end
```

## 接口设计规范

### RESTful API 设计

#### 1. URL 设计

```
格式: /v{version}/{domain}/{resource}/{action}
示例:
  - POST /v1/trade/order/place      # 下单
  - GET  /v1/trade/order/{id}      # 查询订单
  - PUT  /v1/trade/order/{id}/pay  # 支付订单
```

#### 2. 请求/响应格式

**请求**:

```json
{
  "courseIds": [1, 2, 3],
  "couponIds": [10, 20],
  "payType": 1
}
```

**响应**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "orderId": 123456,
    "payAmount": 10000,
    "status": 1
  }
}
```

#### 3. 错误处理

```json
{
  "code": 1001,
  "message": "invalid parameter",
  "data": null
}
```

### gRPC 接口设计

#### 1. 方法命名

- 使用动词：`PlaceOrder`, `GetOrder`, `CancelOrder`
- 避免使用 `Get`, `Set` 前缀（gRPC 风格）

#### 2. 消息设计

- 使用 `Request`/`Response` 后缀
- 字段使用 snake_case
- 必需字段使用 `required`，可选字段使用 `optional`

## 安全架构

### 认证授权

#### 1. JWT 认证

```go
@server(
    jwt: Auth
)
```

**Token 结构**:

- Header: 算法类型
- Payload: 用户信息（user_id, exp, iat）
- Signature: 签名

#### 2. 权限控制

- API 层：JWT 验证
- RPC 层：基于 user_id 的权限校验
- 数据层：行级权限控制

### 数据安全

#### 1. 输入验证

```go
// 验证用户输入
if req.UserId <= 0 {
    return errors.New(ErrCodeInvalidParam, "invalid user_id")
}
```

#### 2. SQL 注入防护

```go
// ✅ 正确：使用参数化查询
db.Query("SELECT * FROM orders WHERE user_id = ?", userId)

// ❌ 错误：字符串拼接
db.Query(fmt.Sprintf("SELECT * FROM orders WHERE user_id = %d", userId))
```

#### 3. 敏感信息保护

- 密码：使用 bcrypt 加密
- Token：不在日志中输出
- 支付信息：加密存储

## 监控和可观测性

### 日志规范

```go
// 结构化日志
logx.WithContext(ctx).Infof(
    "Processing order: orderId=%d, userId=%d, amount=%d",
    orderId, userId, amount,
)
```

### 指标收集

```go
// Prometheus 指标
requestDuration.Observe(duration)
requestCount.Inc()
errorCount.WithLabelValues("trade", "place_order").Inc()
```

### 链路追踪

- 使用 OpenTelemetry 或 Jaeger
- 传递 TraceID 和 SpanID
- 记录关键操作的时间戳

## 部署架构

### 容器化

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
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

## 最佳实践总结

1. ✅ **分层清晰**：API Gateway → RPC Services → Data Layer
2. ✅ **服务独立**：每个服务拥有独立数据库和部署单元
3. ✅ **通信标准化**：gRPC 同步，RocketMQ 异步
4. ✅ **数据一致性**：最终一致性，使用消息队列
5. ✅ **安全第一**：JWT 认证，输入验证，SQL 注入防护
6. ✅ **可观测性**：日志、指标、链路追踪

## 参考资源

- [系统架构图](./architecture.md)
- [项目设计指标](./project-design-indicators.md)
- [系统设计规范](./system-design-specification.md)
- [Go-Zero 约定](../coding-standards/go-zero-conventions.md)
- [微服务设计规则](../coding-standards/service-design-rules.md)
- [性能优化指南](../coding-standards/performance-guidelines.md)
