# 开发流程规范

本文档定义了 Aether-Defense-System 项目的开发流程，确保代码质量和文档同步更新。

## 开发流程概述

每次进行新的代码修改时，必须遵循以下流程：

```
1. TDD（测试驱动开发）
   ↓
2. 实现代码
   ↓
3. 运行所有测试并确保通过
   ↓
4. 更新文档
   ↓
5. 代码审查和提交
```

## 详细流程

### 阶段一：测试驱动开发（TDD）

#### 1.1 编写测试用例

在实现功能之前，先编写测试用例：

```go
// service/trade/rpc/trade_test.go
package rpc

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestTradeService_PlaceOrder(t *testing.T) {
    // Arrange: 准备测试数据
    ctx := context.Background()
    req := &PlaceOrderRequest{
        UserId:    123,
        CourseIds: []int64{1, 2, 3},
        CouponIds: []int64{10},
    }

    // Act: 执行被测试的方法
    resp, err := service.PlaceOrder(ctx, req)

    // Assert: 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Greater(t, resp.OrderId, int64(0))
    assert.Greater(t, resp.PayAmount, int32(0))
}
```

#### 1.2 TDD 循环

遵循 **Red-Green-Refactor** 循环：

1. **Red（红色）**：编写失败的测试
   - 测试应该先失败（因为功能还未实现）
   - 测试应该描述期望的行为

2. **Green（绿色）**：实现最小代码使测试通过
   - 只实现足够让测试通过的代码
   - 不要过度设计

3. **Refactor（重构）**：重构代码提高质量
   - 在测试通过后重构
   - 确保重构后测试仍然通过

#### 1.3 测试类型

根据测试金字塔原则：

- **单元测试（70%）**：测试单个函数或方法
- **集成测试（20%）**：测试服务间交互
- **端到端测试（10%）**：测试完整业务流程

### 阶段二：实现代码

#### 2.1 实现功能

在测试用例编写完成后，实现功能代码：

```go
func (s *TradeService) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
    // 1. 参数校验
    if err := s.validateRequest(req); err != nil {
        return nil, err
    }

    // 2. 业务逻辑
    orderId, err := s.createOrder(ctx, req)
    if err != nil {
        return nil, err
    }

    // 3. 返回结果
    return &PlaceOrderResponse{
        OrderId:   orderId,
        PayAmount: calculateAmount(req),
        Status:    1,
    }, nil
}
```

#### 2.2 代码质量要求

- 遵循项目代码规范（见 `.cursor/rules/coding-standards.mdc`）
- 使用结构化日志
- 正确处理错误
- 添加必要的注释

### 阶段三：运行测试

#### 3.1 运行所有测试

在代码修改后，必须运行所有测试：

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./service/trade/rpc/...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 3.2 测试要求

- **所有测试必须通过**：不允许提交失败的测试
- **测试覆盖率**：核心业务逻辑覆盖率应 > 80%
- **测试性能**：单元测试应在 100ms 内完成
- **并发测试**：使用 `-race` 标志检测竞态条件

```bash
# 运行测试并检测竞态条件
go test -race ./...
```

#### 3.3 测试失败处理

如果测试失败：

1. **分析失败原因**：查看测试输出和日志
2. **修复代码**：修复导致测试失败的代码
3. **重新运行测试**：确保所有测试通过
4. **检查回归**：确保没有引入新的问题

### 阶段四：更新文档

#### 4.1 文档更新检查清单

在测试通过后，检查并更新以下文档：

- [ ] **API 文档**：如果修改了 API，更新 `.api` 文件中的 `@doc` 注释
- [ ] **代码注释**：更新函数和方法的 Go doc 注释
- [ ] **架构文档**：如果涉及架构变更，更新 `doc/design/` 下的文档
- [ ] **README**：如果添加了新功能，更新相关 README
- [ ] **变更日志**：记录重要的变更（如适用）

#### 4.2 API 文档更新

```go
@server(
    group: order
    prefix: /v1/trade/order
    jwt: Auth
)
service trade-api {
    @doc "用户下单接口，支持多课程和优惠券组合购买"
    @handler PlaceOrder
    post /place (PlaceOrderReq) returns (PlaceOrderResp)
}
```

#### 4.3 代码注释更新

```go
// PlaceOrder creates a new order for the user.
// It validates the request, calculates the total amount,
// and creates the order in the database.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - req: Order creation request containing course IDs and coupon IDs
//
// Returns:
//   - OrderResponse with order ID, payment amount, and status
//   - error if order creation fails
func (s *TradeService) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
    // ...
}
```

#### 4.4 文档同步原则

- **代码即文档**：代码应该自解释，注释补充说明
- **及时更新**：代码变更后立即更新文档
- **版本控制**：文档与代码一起提交到版本控制

### 阶段五：代码审查和提交

#### 5.1 提交前检查

在提交代码前，运行以下检查：

```bash
# 1. 运行所有测试
go test ./...

# 2. 运行 lint 检查
golangci-lint run

# 3. 运行 pre-commit hooks
pre-commit run --all-files

# 4. 检查代码格式
go fmt ./...
goimports -w .
```

#### 5.2 提交信息规范

使用规范的提交信息格式：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型（type）**：

- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `test`: 测试相关
- `refactor`: 重构
- `perf`: 性能优化
- `chore`: 构建/工具相关

**示例**：

```
feat(trade): add order placement with coupon support

- Implement PlaceOrder method with coupon validation
- Add unit tests for order creation
- Update API documentation

Closes #123
```

#### 5.3 Pull Request 检查清单

创建 PR 前，确保：

- [ ] 所有测试通过
- [ ] 代码通过 lint 检查
- [ ] 文档已更新
- [ ] 提交信息符合规范
- [ ] 代码已自审查
- [ ] 没有引入安全漏洞

## 自动化工具集成

### Pre-commit Hooks

项目已配置 pre-commit hooks，在提交前自动运行：

- 代码格式化（gofmt, goimports）
- 代码检查（go-vet, golangci-lint）
- 测试运行（go test）
- 文档检查

### CI/CD 集成

在 CI/CD 流程中自动执行：

```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

      - name: Run linter
        uses: golangci/golangci-lint-action@v3
        with:
          version: v1.55.2
```

## 开发流程检查清单

每次代码修改时，使用以下检查清单：

### 开发阶段

- [ ] 编写测试用例（TDD）
- [ ] 实现功能代码
- [ ] 代码符合项目规范

### 测试阶段

- [ ] 运行所有测试：`go test ./...`
- [ ] 所有测试通过
- [ ] 测试覆盖率满足要求（> 80%）
- [ ] 运行竞态检测：`go test -race ./...`

### 文档阶段

- [ ] 更新 API 文档（如适用）
- [ ] 更新代码注释
- [ ] 更新架构文档（如适用）
- [ ] 更新 README（如适用）

### 提交阶段

- [ ] 运行 lint 检查：`golangci-lint run`
- [ ] 运行 pre-commit hooks
- [ ] 提交信息符合规范
- [ ] 创建 PR 并填写检查清单

## 最佳实践

### 1. 小步迭代

- 每次只实现一个小功能
- 频繁提交代码
- 保持代码库始终处于可工作状态

### 2. 测试优先

- 先写测试，后写代码
- 测试应该描述期望的行为
- 测试应该快速、独立、可重复

### 3. 持续集成

- 每次提交都触发 CI
- 快速反馈测试结果
- 及时修复失败的测试

### 4. 文档同步

- 代码变更时同步更新文档
- 使用自动化工具生成文档
- 定期审查文档准确性

### 5. 代码审查

- 所有代码必须经过审查
- 审查时检查测试和文档
- 使用 PR 模板确保完整性

## 常见问题

### Q: 如果测试运行时间太长怎么办？

A:

1. 优化测试，减少不必要的等待
2. 使用测试并行执行：`go test -parallel 4 ./...`
3. 将慢测试标记为集成测试，单独运行

### Q: 如何确保文档不遗漏？

A:

1. 使用 PR 模板，包含文档检查项
2. 在 CI 中检查文档更新
3. 代码审查时检查文档

### Q: 测试覆盖率要求是多少？

A:

- 核心业务逻辑：> 80%
- 工具函数：> 90%
- 简单 getter/setter：可适当降低

### Q: 什么时候可以跳过测试？

A:

- **永远不要跳过测试**
- 如果测试失败，修复代码或测试
- 如果测试过时，更新测试而不是删除

## 参考资源

- [Go 测试文档](https://go.dev/doc/tutorial/add-a-test)
- [TDD 最佳实践](https://www.thoughtworks.com/insights/blog/test-driven-development)
- [项目代码规范](./README.md)
- [Go-Zero 约定](./go-zero-conventions.md)
