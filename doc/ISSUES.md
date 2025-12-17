# 已知问题记录

本文档记录系统开发过程中发现的问题、排查过程和解决方案。

## 问题记录

### 问题 #001: 库存扣减 API 返回成功但库存值未实际减少

**状态**: ✅ 已解决  
**发现时间**: 2025-12-17  
**解决时间**: 2025-12-17  
**影响范围**: Promotion RPC 服务的 `DecrStock` 接口  
**优先级**: 高  
**测试类型**: 🔗 **整合测试（Integration Test）**

#### 问题描述

- **现象**:
  - gRPC API 调用 `DecrStock` 接口返回 `{"success": true, "message": "Inventory deduction successful"}`
  - 服务日志显示请求已成功处理
  - 但 Redis 中的库存值（`inventory:course:{courseId}`）未实际减少

#### 测试类型说明

这是一个**整合测试（Integration Test）**问题，因为：

- ✅ **测试范围**: 涉及多个组件的交互
  - gRPC API 层（接收请求）
  - 业务逻辑层（`DecrStockLogic`）
  - 数据访问层（Redis Client + Lua 脚本）

- ✅ **使用真实依赖**: 测试使用了真实的 Redis 实例（而非 mock）

- ❌ **不是端到端测试**: 未涉及完整的业务流程
  - 没有涉及订单创建、支付等完整用户流程
  - 只测试了单个服务的内部组件集成

根据项目的测试金字塔原则：
- **单元测试（70%）**: 测试单个函数/方法（使用 mocks）
- **整合测试（20%）**: 测试组件间交互（使用真实依赖）← **当前问题属于此类**
- **端到端测试（10%）**: 测试完整业务流程（跨多个服务）

- **测试场景**:
  ```bash
  # 设置初始库存
  docker exec aether-defense-redis redis-cli SET "inventory:course:100" "50"

  # 调用 API 扣减库存（服务未启用 reflection，需显式提供 proto）
  grpcurl -plaintext \
    -import-path service/promotion/rpc \
    -proto service/promotion/rpc/promotion.proto \
    -d '{"courseId": 100, "num": 5}' \
    localhost:8082 promotion.PromotionService/DecrStock

  # 检查库存（应从 50 变为 45）
  docker exec aether-defense-redis redis-cli GET "inventory:course:100"
  # 返回: 45
  ```

#### 排查过程

1. **已验证正常的部分**:
   - ✅ Redis 连接正常（`PING` 成功）
   - ✅ Promotion RPC 服务正常运行
   - ✅ gRPC API 调用成功
   - ✅ 请求验证逻辑正常（courseId、num 校验通过）
   - ✅ 服务日志正常记录请求

2. **直接测试 Lua 脚本**:
   ```bash
   # 直接执行 Lua 脚本可以正常扣减库存
   docker exec aether-defense-redis redis-cli EVAL \
     "local stock = redis.call('GET', KEYS[1]); \
      if (stock == false) then return {err = 'Inventory Key does not exist'}; end; \
      if (tonumber(stock) < tonumber(ARGV[1])) then return {err = 'Insufficient inventory'}; end; \
      local newStock = redis.call('DECRBY', KEYS[1], ARGV[1]); \
      return newStock;" \
     1 "inventory:course:900" 20
   # 返回: 80 (库存从 100 减少到 80，正常)
   ```

3. **最终原因与修复**:
   - ✅ **配置 key 冲突**：`zrpc.RpcServerConf` 内置字段 `Redis`（用于 RPC Auth），与业务配置 `Redis` 同名导致 `conf.MustLoad` 报 `conflict key redis` 并重启。
     - **修复**：将业务 Redis 配置改名为 `InventoryRedis`，并同步修改 Docker/Helm/示例配置。
   - ✅ **Database 配置被强制校验**：`database.Config` 的字段未标记 optional，导致仅配置 `dsn: ""` 时仍报 `database.max_open_conns is not set`。
     - **修复**：将 `common/database.Config` 的字段标记为 optional（默认值在 `NewClient()` 内仍会填充），并在配置中保留 `Database: { dsn: "" }` 让服务可不连 DB 启动。
   - ✅ **扣减链路验证通过**：gRPC 返回成功，同时 Redis 库存实际减少（50 → 45），日志打印 before/after。

#### 可能原因

1. **Lua 脚本返回值解析问题**:
   - Redis 返回 Lua 表时格式可能为 `map[string]interface{}`、`map[interface{}]interface{}` 或 `[]interface{}`
   - 当前解析逻辑可能未正确处理所有情况
   - 脚本返回整数时，解析逻辑可能未正确识别

2. **脚本执行但结果被忽略**:
   - 脚本可能执行成功，但返回值解析失败
   - 错误可能被静默忽略，导致函数返回 `nil`（无错误）

3. **Redis 连接或脚本加载问题**:
   - 脚本可能未正确加载到 Redis
   - 可能存在连接池或上下文问题

#### 相关代码

- **Lua 脚本定义**: `common/redis/client.go:99-115`
- **脚本执行逻辑**: `common/redis/client.go:197-218`
- **业务逻辑调用**: `service/promotion/rpc/internal/logic/decrstocklogic.go:71`
- **服务上下文初始化 Redis**: `service/promotion/rpc/svc/servicecontext.go:38`
- **Docker 配置**: `deploy/docker/config-docker/promotion-rpc.yaml`

#### 验证方式

- 使用脚本：`scripts/test-inventory-deduction.sh 100 50 5`
- 或使用上面的 `grpcurl + redis-cli GET` 手工验证

---

## 已解决的问题

### 问题 #002: RocketMQ Broker 启动失败 (NullPointerException)

**状态**: ✅ 已解决  
**解决时间**: 2025-12-17

#### 问题描述

RocketMQ Broker 启动时出现 `NullPointerException`，错误信息：
```
java.lang.NullPointerException
	at org.apache.rocketmq.broker.schedule.ScheduleMessageService.configFilePath(ScheduleMessageService.java:272)
```

#### 解决方案

1. 添加了 `rocketmq-broker-init` 初始化容器（使用 `init` profile）
2. 在启动前创建 `/home/rocketmq/store/schedule` 目录
3. 在 `broker.conf` 中配置 `storePathScheduleMessage = /home/rocketmq/store/schedule`

**相关文件**:
- `deploy/docker/docker-compose.yml:123-131`
- `deploy/docker/rocketmq/broker.conf:22`

---

### 问题 #003: RocketMQ 健康检查失败

**状态**: ✅ 已解决  
**解决时间**: 2025-12-17

#### 问题描述

RocketMQ NameServer 和 Broker 的健康检查失败，错误信息：
```
sh: netstat: command not found
```

#### 解决方案

将健康检查从 `netstat` 改为进程检查：
- NameServer: `ps aux | grep -v grep | grep mqnamesrv`
- Broker: `ps aux | grep -v grep | grep mqbroker`

**相关文件**:
- `deploy/docker/docker-compose.yml:114-119` (NameServer)
- `deploy/docker/docker-compose.yml:150-155` (Broker)
- `scripts/test-docker-compose.sh:71-84`

---

### 问题 #004: Promotion RPC 服务配置错误

**状态**: ✅ 已解决  
**解决时间**: 2025-12-17

#### 问题描述

Promotion RPC 服务启动失败，错误信息：
```
error: config file /app/promotion.yaml, field "redis.Host" is not set
error: config file /app/promotion.yaml, field "redis.Key" is not set
```

#### 解决方案

1. 在 `redis.Config` 结构体中添加 `Host` 和 `Key` 字段（用于 go-zero 兼容性）
2. 在配置文件中添加相应字段
3. 在 `NewClient` 方法中优先使用 `Addr`，如果为空则使用 `Host`

**相关文件**:
- `common/redis/client.go:24-33` (Config 结构体)
- `common/redis/client.go:56-65` (NewClient 方法)
- `deploy/docker/config-docker/promotion-rpc.yaml:12-23`

---

## 问题分类

### 按状态分类

- 🔴 **待解决**: 0 个
- ✅ **已解决**: 4 个

### 按优先级分类

- **高优先级**: 0 个
- **中优先级**: 0 个
- **低优先级**: 0 个

### 按模块分类

- **Promotion RPC**: 2 个（1 个待解决，1 个已解决）
- **RocketMQ**: 2 个（均已解决）
- **Redis**: 1 个（已解决，但可能与库存扣减问题相关）

---

## 更新日志

- **2025-12-17**: 创建问题记录文档，记录库存扣减问题和已解决的 RocketMQ 相关问题
- **2025-12-17**: 修复并验证库存扣减链路（配置冲突 + Database optional + gRPC→Redis 扣减成功）
