# Docker Compose 测试指南

## 快速开始

### 1. 启动基础设施服务

```bash
# 使用便捷脚本（推荐）
./deploy/docker/start-local.sh

# 或手动启动
docker compose -f deploy/docker/docker-compose.yml up -d etcd redis mysql
```

### 2. 运行测试脚本

```bash
# 测试所有服务健康状态
./scripts/test-docker-compose.sh
```

## 测试场景

### 场景 1: 仅测试基础设施（推荐用于开发）

**启动基础设施：**
```bash
docker compose -f deploy/docker/docker-compose.yml up -d etcd redis mysql
```

**本地运行服务：**
```bash
# 使用本地配置文件
cp deploy/docker/config-examples/promotion-rpc.local.yaml service/promotion/rpc/etc/promotion.yaml

# 运行服务
go run ./service/promotion/rpc/promotion.go
```

**测试库存扣减：**
```bash
# 设置库存
docker exec aether-defense-redis redis-cli SET "inventory:course:1" "100"

# 测试 Lua 脚本扣减
docker exec aether-defense-redis redis-cli EVAL "
local stock = redis.call('GET', KEYS[1])
if (stock == false) then
    return {err = 'Inventory Key does not exist'}
end
if (tonumber(stock) < tonumber(ARGV[1])) then
    return {err = 'Insufficient inventory'}
end
redis.call('DECRBY', KEYS[1], ARGV[1])
return {result = 'Inventory deduction successful'}
" 1 "inventory:course:1" 10

# 查看剩余库存
docker exec aether-defense-redis redis-cli GET "inventory:course:1"
```

### 场景 2: 完整容器化测试

**构建并启动所有服务：**
```bash
docker compose -f deploy/docker/docker-compose.yml up --build -d
```

**查看服务日志：**
```bash
# 查看所有服务日志
docker compose -f deploy/docker/docker-compose.yml logs -f

# 查看特定服务日志
docker compose -f deploy/docker/docker-compose.yml logs -f promotion-rpc
```

**测试服务健康：**
```bash
# 检查服务状态
docker compose -f deploy/docker/docker-compose.yml ps

# 测试 Redis 连接
docker exec aether-defense-redis redis-cli ping

# 测试 etcd 连接
docker exec aether-defense-etcd etcdctl endpoint health
```

### 场景 3: 测试 Promotion RPC 服务

**1. 确保服务运行：**
```bash
docker compose -f deploy/docker/docker-compose.yml ps promotion-rpc
```

**2. 设置测试库存：**
```bash
docker exec aether-defense-redis redis-cli SET "inventory:course:100" "50"
```

**3. 使用 gRPC 客户端测试（需要 grpcurl 工具）：**
```bash
# 安装 grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 测试 DecrStock
grpcurl -plaintext \
  -d '{"courseId": 100, "num": 5}' \
  localhost:8082 \
  promotion.PromotionService/DecrStock
```

**4. 验证库存变化：**
```bash
docker exec aether-defense-redis redis-cli GET "inventory:course:100"
# 应该返回: 45 (50 - 5)
```

## 常用测试命令

### 检查服务状态
```bash
docker compose -f deploy/docker/docker-compose.yml ps
```

### 查看服务日志
```bash
# 所有服务
docker compose -f deploy/docker/docker-compose.yml logs

# 特定服务
docker compose -f deploy/docker/docker-compose.yml logs promotion-rpc

# 实时跟踪
docker compose -f deploy/docker/docker-compose.yml logs -f promotion-rpc
```

### 重启服务
```bash
# 重启单个服务
docker compose -f deploy/docker/docker-compose.yml restart promotion-rpc

# 重启所有服务
docker compose -f deploy/docker/docker-compose.yml restart
```

### 停止服务
```bash
# 停止所有服务
./deploy/docker/stop-local.sh

# 或手动
docker compose -f deploy/docker/docker-compose.yml down

# 停止并删除数据卷
docker compose -f deploy/docker/docker-compose.yml down -v
```

## Redis 测试

### 基本操作测试
```bash
# 设置值
docker exec aether-defense-redis redis-cli SET "test:key" "test:value"

# 获取值
docker exec aether-defense-redis redis-cli GET "test:key"

# 删除值
docker exec aether-defense-redis redis-cli DEL "test:key"
```

### 库存操作测试
```bash
# 设置库存
docker exec aether-defense-redis redis-cli SET "inventory:course:1" "100"

# 查看库存
docker exec aether-defense-redis redis-cli GET "inventory:course:1"

# 使用 Lua 脚本扣减库存（模拟 DecrStock）
docker exec aether-defense-redis redis-cli EVAL "
local stock = redis.call('GET', KEYS[1])
if (stock == false) then
    return {err = 'Inventory Key does not exist'}
end
if (tonumber(stock) < tonumber(ARGV[1])) then
    return {err = 'Insufficient inventory'}
end
redis.call('DECRBY', KEYS[1], ARGV[1])
return {result = 'Inventory deduction successful'}
" 1 "inventory:course:1" 10

# 验证结果
docker exec aether-defense-redis redis-cli GET "inventory:course:1"
```

## 故障排查

> **注意**: 详细的已知问题记录请参考 [ISSUES.md](../../doc/ISSUES.md)

### 服务无法启动
```bash
# 查看详细日志
docker compose -f deploy/docker/docker-compose.yml logs promotion-rpc

# 检查端口占用
netstat -tuln | grep 8082
```

### Redis 连接失败
```bash
# 检查 Redis 是否运行
docker exec aether-defense-redis redis-cli ping

# 检查网络连接
docker network inspect aether-defense-network
```

### 配置问题
```bash
# 验证配置文件
cat deploy/docker/config-docker/promotion-rpc.yaml

# 检查容器内的配置
docker exec aether-defense-promotion-rpc cat /app/promotion.yaml
```

## 性能测试

### 测试 Redis 并发性能
```bash
# 使用 redis-benchmark
docker exec aether-defense-redis redis-benchmark -h localhost -p 6379 -n 10000 -c 100
```

### 测试库存扣减并发
```bash
# 设置初始库存
docker exec aether-defense-redis redis-cli SET "inventory:course:999" "1000"

# 并发测试（需要编写测试脚本）
# 可以创建多个并发请求测试原子性
```
