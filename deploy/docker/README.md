# Local Development Environment

This directory contains Docker Compose configuration for local development and testing.

## Quick Start

### Option 1: Start Only Infrastructure (Recommended for Development)

Start infrastructure services and run applications locally with `go run`:

```bash
# Using convenience script
./deploy/docker/start-local.sh

# Or manually - start only infrastructure
docker-compose -f deploy/docker/docker-compose.yml up -d etcd redis mysql

# Then run services locally
go run ./service/user/rpc/cmd/user-rpc
go run ./service/user/api/cmd/user-api
```

### Option 2: Build and Start All Services (Containerized)

Build and run all services in Docker containers:

```bash
# Build and start all services
docker-compose -f deploy/docker/docker-compose.yml up --build -d

# Check service status
docker-compose -f deploy/docker/docker-compose.yml ps

# View logs
docker-compose -f deploy/docker/docker-compose.yml logs -f user-rpc
```

### Stop All Services

```bash
# Using convenience script
./deploy/docker/stop-local.sh

# Or manually
docker-compose -f deploy/docker/docker-compose.yml down

# Remove volumes (clean data)
docker-compose -f deploy/docker/docker-compose.yml down -v
```

## Services

### etcd (Service Discovery)
- **Port**: `2379`
- **Health Check**: `etcdctl endpoint health`
- **Usage**: Service registration and discovery for gRPC services

### Redis (Cache & Inventory)
- **Port**: `6379`
- **Health Check**: `redis-cli ping`
- **Usage**:
  - Inventory management (Lua scripts)
  - Cache layer
  - Rate limiting

### MySQL (Database)
- **Port**: `3306`
- **Credentials**:
  - Root: `root` / `root123`
  - User: `aether` / `aether123`
  - Database: `aether_defense`
- **Usage**: Persistent data storage

### RocketMQ (Message Queue)
- **NameServer Port**: `9876`
- **Broker Ports**: `10909` (VIP), `10911` (Main)
- **Console**: `http://localhost:8080` (RocketMQ Console)
- **Usage**: Distributed transaction messages, async processing

## Configuration

### Service Configuration Files

When running services locally, update configuration files to use local service names:

**For local development (docker-compose network):**
```yaml
Etcd:
  Hosts:
    - etcd:2379  # Use service name in docker-compose network
```

**For local development (host network):**
```yaml
Etcd:
  Hosts:
    - 127.0.0.1:2379  # Use localhost
```

### Environment Variables

You can override default settings using environment variables:

```bash
# Custom MySQL password
MYSQL_ROOT_PASSWORD=custom_password docker-compose -f deploy/docker/docker-compose.yml up -d
```

## Data Persistence

All data is persisted in Docker volumes:
- `etcd-data`: etcd cluster data
- `redis-data`: Redis persistence (AOF)
- `mysql-data`: MySQL database files
- `rocketmq-*-logs`: RocketMQ logs
- `rocketmq-broker-store`: RocketMQ message store

To reset all data:
```bash
docker-compose -f deploy/docker/docker-compose.yml down -v
```

## Troubleshooting

### Port Conflicts

If ports are already in use, modify port mappings in `docker-compose.yml`:

```yaml
ports:
  - "23790:2379"  # Change host port
```

### Service Health Checks

Check service health:
```bash
# etcd
docker exec aether-defense-etcd etcdctl endpoint health

# Redis
docker exec aether-defense-redis redis-cli ping

# MySQL
docker exec aether-defense-mysql mysqladmin ping -h localhost -u root -proot123
```

### View Logs

```bash
# All services
docker-compose -f deploy/docker/docker-compose.yml logs

# Specific service
docker-compose -f deploy/docker/docker-compose.yml logs etcd
docker-compose -f deploy/docker/docker-compose.yml logs redis
```

### Reset Services

```bash
# Stop and remove containers
docker-compose -f deploy/docker/docker-compose.yml down

# Remove volumes (deletes all data)
docker-compose -f deploy/docker/docker-compose.yml down -v

# Rebuild and start
docker-compose -f deploy/docker/docker-compose.yml up -d
```

## Integration with Services

### Running Services Locally (Option 1: Infrastructure in Docker, Apps Locally)

1. Start infrastructure:
   ```bash
   ./deploy/docker/start-local.sh
   # Or: docker-compose -f deploy/docker/docker-compose.yml up -d etcd redis mysql
   ```

2. Update service configs to use localhost:
   ```bash
   # Copy local config examples
   cp deploy/docker/config-examples/user-api.local.yaml service/user/api/etc/user-api.yaml
   cp deploy/docker/config-examples/user-rpc.local.yaml service/user/rpc/etc/user.yaml
   cp deploy/docker/config-examples/trade-rpc.local.yaml service/trade/rpc/etc/trade.yaml
   cp deploy/docker/config-examples/promotion-rpc.local.yaml service/promotion/rpc/etc/promotion.yaml
   ```

3. Run services locally:
   ```bash
   go run ./service/user/rpc/cmd/user-rpc
   go run ./service/user/api/cmd/user-api
   ```

### Running Services in Docker (Option 2: Everything Containerized)

1. Build and start all services:
   ```bash
   docker-compose -f deploy/docker/docker-compose.yml up --build -d
   ```

2. Services will use Docker network service names automatically:
   - etcd: `etcd:2379` (from config-docker/*.yaml)
   - Redis: `redis:6379`
   - MySQL: `mysql:3306`

## Network Configuration

All services are connected via `aether-defense-network` bridge network.

Services can communicate using:
- **Service names** (within docker network): `etcd:2379`, `redis:6379`
- **localhost** (from host): `127.0.0.1:2379`, `127.0.0.1:6379`
