# 性能优化指南

本文档定义了 Aether-Defense-System 项目的性能优化指南，确保系统能够满足 40,000 QPS 峰值流量和 P99 < 100ms 的延迟要求。

## 目录

- [性能指标](#性能指标)
- [Go 语言性能优化](#go-语言性能优化)
- [Redis 性能优化](#redis-性能优化)
- [数据库性能优化](#数据库性能优化)
- [消息队列优化](#消息队列优化)
- [网络优化](#网络优化)

## 性能指标

### 系统性能目标

- **峰值 QPS**: 40,000 QPS
- **P50 延迟**: < 20ms
- **P99 延迟**: < 100ms
- **P99.9 延迟**: < 300ms
- **可用性**: 99.99%

### 各组件性能预算

| 组件 | 性能指标 | 说明 |
|------|---------|------|
| API Gateway | 40,000 RPS | 入口流量 |
| 业务服务 | 2,000-3,000 RPS/节点 | 单 Pod 处理能力 |
| Redis | 60,000 QPS | 缓存和原子操作 |
| MySQL | 800 TPS (Write) | 写入事务处理 |
| RocketMQ | < 2ms 写入延迟 | 消息发送延迟 |

## Go 语言性能优化

### 内存管理

#### 1. 使用 sync.Pool 复用对象

在热点代码路径中使用对象池：

```go
var orderPool = sync.Pool{
    New: func() interface{} {
        return &Order{}
    },
}

func processOrder(ctx context.Context) {
    // 从池中获取对象
    order := orderPool.Get().(*Order)
    defer orderPool.Put(order)
    
    // 重置对象状态
    order.Reset()
    
    // 使用对象
    // ...
}
```

#### 2. 避免不必要的内存分配

```go
// ❌ 错误：每次调用都分配新的 slice
func getItems() []Item {
    return []Item{...}  // 分配新内存
}

// ✅ 正确：复用预分配的 slice
var itemsBuffer = make([]Item, 0, 100)
func getItems() []Item {
    itemsBuffer = itemsBuffer[:0]  // 复用底层数组
    // ... 填充数据
    return itemsBuffer
}
```

#### 3. 使用预分配 slice

```go
// ❌ 错误：动态扩容
items := []Item{}
for i := 0; i < 1000; i++ {
    items = append(items, item)  // 可能多次扩容
}

// ✅ 正确：预分配容量
items := make([]Item, 0, 1000)
for i := 0; i < 1000; i++ {
    items = append(items, item)  // 无需扩容
}
```

### GC 优化

#### 1. 减少堆内存分配

```go
// ❌ 错误：频繁分配小对象
func processRequest(req *Request) {
    log := fmt.Sprintf("Processing: %s", req.ID)  // 分配字符串
    // ...
}

// ✅ 正确：使用对象池或复用
var logBuffer = &strings.Builder{}
func processRequest(req *Request) {
    logBuffer.Reset()
    logBuffer.WriteString("Processing: ")
    logBuffer.WriteString(req.ID)
    // ...
}
```

#### 2. 避免反射

```go
// ❌ 错误：使用反射（性能差）
func serialize(v interface{}) []byte {
    return json.Marshal(v)  // 使用反射
}

// ✅ 正确：使用代码生成或预编译
func serializeOrder(o *Order) []byte {
    // 手动序列化或使用代码生成
    // ...
}
```

#### 3. 使用 json-iterator

```go
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func marshal(v interface{}) ([]byte, error) {
    return json.Marshal(v)  // 比标准库快 2-3 倍
}
```

### 并发优化

#### 1. 控制 Goroutine 数量

```go
// ❌ 错误：无限制创建 goroutine
for _, item := range items {
    go processItem(item)  // 可能创建大量 goroutine
}

// ✅ 正确：使用 worker pool
type WorkerPool struct {
    workers int
    jobs    chan Item
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        go p.worker()
    }
}

func (p *WorkerPool) worker() {
    for item := range p.jobs {
        processItem(item)
    }
}
```

#### 2. 使用 Context 控制超时

```go
// 为每个请求设置超时
ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
defer cancel()

// 在超时时间内完成操作
result, err := doWork(ctx)
if err != nil {
    if err == context.DeadlineExceeded {
        return errors.New(ErrCodeTimeout, "operation timeout")
    }
    return err
}
```

## Redis 性能优化

### Lua 脚本优化

#### 1. 脚本执行时间限制

**硬性要求：Lua 脚本执行时间必须 < 0.5ms**

```lua
-- ✅ 正确：简单快速的脚本
-- KEYS[1]: stock key
-- ARGV[1]: quantity
local stock = redis.call('GET', KEYS[1])
if tonumber(stock) < tonumber(ARGV[1]) then
    return {err = "insufficient"}
end
redis.call('DECRBY', KEYS[1], ARGV[1])
return {ok = true}

-- ❌ 错误：复杂的循环操作
for i = 1, 1000 do
    redis.call('GET', 'key' .. i)  -- 执行时间过长
end
```

#### 2. 脚本缓存

```go
// 预加载并缓存脚本
script := redis.NewScript(`
    local stock = redis.call('GET', KEYS[1])
    -- ...
`)

// 执行时使用缓存的 SHA
result, err := script.Run(ctx, rdb, []string{"stock:123"}, 1).Result()
```

### Key 设计

#### 1. 命名规范

```
格式: {domain}:{resource}:{id}
示例:
  - promotion:stock:123
  - trade:order:456
  - user:session:789
```

#### 2. 过期时间

**必须为所有 Key 设置过期时间**，避免内存泄漏：

```go
// ✅ 正确：设置过期时间
err := rdb.Set(ctx, "key", "value", 5*time.Minute).Err()

// ❌ 错误：未设置过期时间
err := rdb.Set(ctx, "key", "value", 0).Err()  // 永不过期
```

#### 3. 热点 Key 处理

```go
// 使用本地缓存 + Redis 二级缓存
type Cache struct {
    local  *ristretto.Cache  // 本地缓存
    redis  *redis.Client     // Redis 缓存
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
    // 先查本地缓存
    if val, ok := c.local.Get(key); ok {
        return val.(string), nil
    }
    
    // 再查 Redis
    val, err := c.redis.Get(ctx, key).Result()
    if err == nil {
        c.local.Set(key, val, 1*time.Second)  // 缓存 1 秒
    }
    return val, err
}
```

### Pipeline 使用

```go
// 批量操作使用 Pipeline
pipe := rdb.Pipeline()
for _, key := range keys {
    pipe.Get(ctx, key)
}
results, err := pipe.Exec(ctx)
```

## 数据库性能优化

### 查询优化

#### 1. 索引设计

```sql
-- ✅ 正确：为常用查询创建联合索引
CREATE INDEX idx_user_status ON trade_order(user_id, status);

-- ❌ 错误：缺少索引导致全表扫描
SELECT * FROM trade_order WHERE user_id = 123 AND status = 1;
```

#### 2. 避免 N+1 查询

```go
// ❌ 错误：N+1 查询
for _, order := range orders {
    items, _ := db.GetOrderItems(order.ID)  // 每次循环都查询
}

// ✅ 正确：批量查询
orderIDs := make([]int64, len(orders))
for i, order := range orders {
    orderIDs[i] = order.ID
}
itemsMap, _ := db.GetOrderItemsBatch(orderIDs)  // 一次查询
```

#### 3. 使用连接池

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 写入优化

#### 1. 批量插入

```go
// ❌ 错误：逐条插入
for _, item := range items {
    db.Insert(item)  // 每次插入都开启事务
}

// ✅ 正确：批量插入
db.InsertBatch(items)  // 一次事务插入多条
```

#### 2. 异步写入

```go
// 使用消息队列异步写入数据库
func createOrder(ctx context.Context, order *Order) error {
    // 1. 先写入 Redis（快速响应）
    redis.Set(ctx, fmt.Sprintf("order:%d", order.ID), order, 5*time.Minute)
    
    // 2. 发送消息到 MQ（异步写入数据库）
    mq.SendMessage("order-create", order)
    
    return nil
}
```

## 消息队列优化

### RocketMQ 优化

#### 1. 事务消息性能

```go
// 事务消息必须实现高效的 check-back 接口
func (s *TradeService) CheckLocalTransaction(ctx context.Context, msg *primitive.MessageExt) (primitive.LocalTransactionState, error) {
    // ✅ 正确：使用 Redis 快速查询
    exists, err := redis.Exists(ctx, fmt.Sprintf("order:%s", msg.MsgId)).Result()
    if err != nil {
        return primitive.UnknownTransactionState, err
    }
    if exists {
        return primitive.CommitMessageState, nil
    }
    return primitive.RollbackMessageState, nil
}
```

#### 2. 消息批量发送

```go
// 批量发送消息
messages := make([]*primitive.Message, 0, 100)
for _, order := range orders {
    msg := primitive.NewMessage("order-topic", orderData)
    messages = append(messages, msg)
}
producer.SendBatch(ctx, messages)
```

## 网络优化

### gRPC 优化

#### 1. 连接复用

```go
// 复用 gRPC 连接
var clientConn *grpc.ClientConn

func init() {
    conn, err := grpc.Dial("user-rpc:8080", grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    clientConn = conn
}
```

#### 2. 压缩

```go
// 启用 gzip 压缩
conn, err := grpc.Dial(
    "user-rpc:8080",
    grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")),
)
```

### HTTP 优化

#### 1. 连接池

```go
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}
client := &http.Client{Transport: transport}
```

## 性能测试

### 基准测试

```go
func BenchmarkPlaceOrder(b *testing.B) {
    s := setupService()
    req := &PlaceOrderRequest{...}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = s.PlaceOrder(context.Background(), req)
    }
}
```

### 压力测试

使用工具进行压力测试：

```bash
# 使用 wrk 进行 HTTP 压力测试
wrk -t12 -c400 -d30s --latency http://localhost:8080/v1/trade/order/place

# 使用 go-wrk 进行 gRPC 压力测试
go-wrk -c 100 -d 30s -M POST -body '{"courseIds":[1,2]}' http://localhost:8080/v1/trade/order/place
```

## 性能监控

### 指标收集

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
            Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5},
        },
        []string{"method", "endpoint"},
    )
)
```

### 告警阈值

- **P99 延迟 > 100ms**: 触发告警
- **错误率 > 1%**: 触发告警
- **QPS > 80% 容量**: 触发告警

## 最佳实践总结

1. ✅ **内存管理**：使用 sync.Pool，避免频繁分配
2. ✅ **GC 优化**：减少堆分配，避免反射
3. ✅ **并发控制**：使用 worker pool，控制 goroutine 数量
4. ✅ **Redis 优化**：Lua 脚本 < 0.5ms，设置过期时间
5. ✅ **数据库优化**：合理索引，批量操作，异步写入
6. ✅ **消息队列**：批量发送，高效 check-back
7. ✅ **性能测试**：定期进行压力测试和基准测试

## 参考资源

- [项目设计指标文档](../design/project-design-indicators.md)
- [系统设计规范文档](../design/system-design-specification.md)

