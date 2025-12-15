# Go-Zero Framework Conventions

本文档定义了在 Aether-Defense-System 项目中使用 Go-Zero 框架的约定和最佳实践。

## 目录

- [API 定义规范](#api-定义规范)
- [RPC 服务规范](#rpc-服务规范)
- [错误处理](#错误处理)
- [中间件使用](#中间件使用)
- [代码生成](#代码生成)

## API 定义规范

### .api 文件结构

所有 HTTP API 必须定义在 `.api` 文件中，遵循以下结构：

```go
syntax = "v1"

type (
    // Request types
    PlaceOrderReq {
        CourseIds  []int64 `json:"courseIds"`
        CouponIds  []int64 `json:"couponIds,optional"`
        OrderId    int64   `json:"orderId"`
    }

    // Response types
    PlaceOrderResp {
        OrderId     int64  `json:"orderId"`
        PayAmount   int    `json:"payAmount"`
        Status      int    `json:"status"`
    }
)

@server(
    group: order
    prefix: /v1/trade/order
    jwt: Auth
    middleware: RateLimit
)
service trade-api {
    @doc "用户下单接口"
    @handler PlaceOrder
    post /place (PlaceOrderReq) returns (PlaceOrderResp)
}
```

### 命名约定

- **服务名**：使用 `{domain}-api` 格式（如 `trade-api`, `promotion-api`）
- **路由前缀**：使用 `/v1/{domain}/{resource}` 格式
- **Handler 名**：使用 PascalCase，与 HTTP 方法语义对应
- **类型名**：Request 类型以 `Req` 结尾，Response 类型以 `Resp` 结尾

### 必需注解

1. **@doc**：所有接口必须包含文档注释

   ```go
   @doc "用户下单接口，支持多课程和优惠券"
   ```

2. **@handler**：指定处理函数名

   ```go
   @handler PlaceOrder
   ```

3. **@server**：配置服务级设置
   - `group`: 路由分组
   - `prefix`: URL 前缀
   - `jwt`: JWT 认证配置
   - `middleware`: 中间件列表

### 参数验证

使用 tag 进行参数验证：

```go
type PlaceOrderReq {
    CourseIds  []int64 `json:"courseIds" validate:"required,min=1"`
    CouponIds  []int64 `json:"couponIds,optional"`
    PayType    int8    `json:"payType,options=1|2"`  // 枚举值
    Amount     int     `json:"amount,range=[1:1000000]"`  // 范围
}
```

### 安全规范

- **禁止在请求体中传递 user_id**：必须从 JWT token 中解析
- **使用 JWT 中间件**：在 `@server` 块中配置 `jwt: Auth`
- **敏感信息**：密码、token 等敏感信息不得出现在 URL 或日志中

## RPC 服务规范

### .proto 文件定义

```protobuf
syntax = "proto3";

package trade;

option go_package = "github.com/aether-defense-system/service/trade/rpc/trade";

service TradeService {
  rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
}

message PlaceOrderRequest {
  int64 user_id = 1;
  repeated int64 course_ids = 2;
  repeated int64 coupon_ids = 3;
  int64 order_id = 4;
}

message PlaceOrderResponse {
  int64 order_id = 1;
  int32 pay_amount = 2;
  int32 status = 3;
}
```

### 服务实现规范

```go
package rpc

import (
    "context"
    "github.com/zeromicro/go-zero/core/logx"
)

type TradeService struct {
    // Dependencies
}

func (s *TradeService) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
    // 1. 参数校验
    if err := s.validateRequest(req); err != nil {
        return nil, err
    }

    // 2. 业务逻辑
    // ...

    // 3. 返回结果
    return &PlaceOrderResponse{
        OrderId: orderId,
        PayAmount: payAmount,
        Status: 1,
    }, nil
}
```

### Context 使用

- **所有 RPC 方法必须接收 `context.Context` 作为第一个参数**
- **使用 context 传递超时和取消信号**
- **使用 context 传递请求 ID、用户 ID 等元数据**

```go
func (s *TradeService) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
    // 从 context 获取请求 ID
    requestID := ctx.Value("request_id").(string)

    // 设置超时
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // ...
}
```

## 错误处理

### 错误码定义

在 `common/errors/` 目录下定义统一的错误码：

```go
package errors

const (
    ErrCodeInvalidParam = 1001
    ErrCodeNotFound     = 1002
    ErrCodeInternal     = 1003
)

var (
    ErrInvalidParam = errors.New("invalid parameter")
    ErrNotFound     = errors.New("resource not found")
)
```

### 错误返回规范

```go
import (
    "github.com/zeromicro/go-zero/core/logx"
    "github.com/zeromicro/go-zero/core/stores/errors"
)

func (s *TradeService) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
    // 参数校验失败
    if req.UserId <= 0 {
        logx.WithContext(ctx).Errorf("Invalid user_id: %d", req.UserId)
        return nil, errors.New(ErrCodeInvalidParam, "invalid user_id")
    }

    // 业务逻辑错误
    if err := s.processOrder(ctx, req); err != nil {
        logx.WithContext(ctx).Errorf("Failed to process order: %v", err)
        return nil, err
    }

    // ...
}
```

### 错误日志

- **记录错误时包含足够的上下文**：request ID, user ID, operation name
- **使用结构化日志**：`logx.WithContext(ctx).Errorf(...)`
- **区分错误级别**：Error（需要处理）、Warn（可恢复）、Info（信息性）

## 中间件使用

### JWT 认证中间件

在 `.api` 文件中配置：

```go
@server(
    jwt: Auth
)
```

在 handler 中获取用户信息：

```go
func PlaceOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 从 JWT 中获取 user_id
        userId := r.Context().Value("userId").(int64)
        // ...
    }
}
```

### 限流中间件

```go
@server(
    middleware: RateLimit
)
```

在 `common/middleware/ratelimit.go` 中实现：

```go
func RateLimitMiddleware() rest.Middleware {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // 限流逻辑
            // ...
            next(w, r)
        }
    }
}
```

### 日志中间件

Go-Zero 自动提供日志中间件，记录请求和响应信息。

## 代码生成

### 生成 API 代码

```bash
goctl api go -api api/trade.api -dir cmd/api/trade-api
```

### 生成 RPC 代码

```bash
goctl rpc protoc service/trade/rpc/trade.proto --go_out=service/trade/rpc --go-grpc_out=service/trade/rpc --zrpc_out=cmd/rpc/trade-rpc
```

### 生成 Model 代码

```bash
goctl model mysql datasource -url="user:password@tcp(host:port)/database" -table="table_name" -dir="model" -cache
```

## 最佳实践

1. **保持 .api 和 .proto 文件与实现同步**
2. **使用代码生成工具，不要手动编写模板代码**
3. **在生成代码的基础上添加业务逻辑**
4. **遵循 Go-Zero 的目录结构约定**
5. **使用 Go-Zero 提供的工具和中间件**

## 参考资源

- [Go-Zero 官方文档](https://go-zero.dev/)
- [Go-Zero 示例项目](https://github.com/zeromicro/zero-examples)
- [项目架构设计文档](../design/architecture.md)
