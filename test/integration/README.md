# 整合测试（Integration Tests）

本目录包含系统的整合测试，用于测试多个组件之间的交互。

## 目录结构

```
test/integration/
├── README.md              # 本文档
├── setup.go               # 测试环境设置工具
├── promotion/             # Promotion 服务整合测试
│   └── stock_test.go      # 库存扣减整合测试
└── testdata/              # 测试数据文件
```

## 运行整合测试

### 前置条件

1. **Redis 服务必须运行**
   - 本地开发：确保 Redis 在 `localhost:6379` 运行
   - Docker 环境：使用 `docker compose -f deploy/docker/docker-compose.yml up -d redis`

2. **环境变量（可选）**
   ```bash
   export REDIS_ADDR=localhost:6379  # Redis 地址（默认：localhost:6379）
   export REDIS_DB=15                 # Redis 数据库（默认：15，用于测试）
   export REDIS_PASSWORD=             # Redis 密码（默认：空）
   ```

### 运行所有整合测试

```bash
# 从项目根目录运行
go test -tags=integration ./test/integration/... -v

# 或者指定 Redis 地址
REDIS_ADDR=localhost:6379 go test -tags=integration ./test/integration/... -v
```

### 运行特定测试

```bash
# 运行 Promotion 服务的库存扣减测试
go test -tags=integration ./test/integration/promotion -v -run TestIntegration_StockDeduction

# 运行特定测试用例
go test -tags=integration ./test/integration/promotion -v -run TestIntegration_StockDeduction_Success
```

### 在 Docker 环境中运行

```bash
# 启动测试环境
docker compose -f deploy/docker/docker-compose.yml up -d redis

# 运行测试（从容器内或本地）
REDIS_ADDR=localhost:6379 go test -tags=integration ./test/integration/... -v
```

## 测试用例

### Promotion 服务 - 库存扣减测试

#### TestIntegration_StockDeduction_Success
- **目的**: 测试成功的库存扣减流程
- **验证**:
  - API 返回成功
  - Redis 中的库存值实际减少
  - 业务逻辑层和数据层的集成正常

#### TestIntegration_StockDeduction_InsufficientStock
- **目的**: 测试库存不足的情况
- **验证**:
  - API 返回错误
  - Redis 中的库存值不变
  - 错误消息正确

#### TestIntegration_StockDeduction_NonExistentKey
- **目的**: 测试不存在的库存键
- **验证**:
  - API 返回错误
  - 错误消息正确

#### TestIntegration_StockDeduction_MultipleDeductions
- **目的**: 测试多次顺序扣减
- **验证**:
  - 每次扣减都成功
  - 最终库存值正确

#### TestIntegration_StockDeduction_ConcurrentDeductions
- **目的**: 测试并发扣减的原子性
- **验证**:
  - 所有并发请求都成功
  - 最终库存值正确（验证 Lua 脚本的原子性）

## 测试环境

整合测试使用真实的 Redis 实例（而非 mock），但使用独立的测试数据库（DB 15）以避免与生产数据冲突。

### 自动清理

每个测试都会：
1. 在开始前清理测试数据库（`FlushDB`）
2. 在结束后再次清理（通过 `t.Cleanup()`）

### 跳过测试

如果 Redis 不可用，测试会自动跳过（使用 `t.Skipf`），不会导致测试失败。

## 与单元测试的区别

| 特性 | 单元测试 | 整合测试 |
|------|---------|---------|
| **位置** | 源码目录（`*_test.go`） | `test/integration/` |
| **依赖** | 使用 mocks | 使用真实依赖（Redis） |
| **范围** | 单个函数/方法 | 多个组件交互 |
| **速度** | 快 | 较慢 |
| **覆盖率** | 70% | 20% |

## 注意事项

1. **数据隔离**: 测试使用独立的 Redis 数据库（DB 15），但仍需确保测试之间不相互影响
2. **清理资源**: 测试会自动清理，但如有问题请手动检查
3. **并发安全**: 并发测试验证了 Lua 脚本的原子性，这是整合测试的重要价值
4. **CI/CD**: 在 CI/CD 环境中，需要确保 Redis 服务可用或使用 testcontainers

## 相关文档

- [测试结构规范](../../.cursor/rules/11-testing-structure.mdc)
- [已知问题记录](../../doc/ISSUES.md) - 问题 #001: 库存扣减问题
