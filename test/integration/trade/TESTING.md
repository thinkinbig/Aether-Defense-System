# Trade Service Integration Testing Guide

## Overview

Integration tests for the trade service verify the complete order placement and cancellation flow, including:
- User validation via User RPC
- Order creation in MySQL database
- RocketMQ transactional message sending
- Order cancellation with status updates

## Quick Start

### 1. Start Required Services

```bash
# Start infrastructure services
docker-compose -f deploy/docker/docker-compose.yml up -d mysql redis rocketmq-nameserver rocketmq-broker etcd

# Start RPC services (in separate terminals)
go run ./cmd/rpc/user-rpc
go run ./cmd/rpc/trade-rpc

# Or use Docker
docker-compose -f deploy/docker/docker-compose.yml up -d user-rpc trade-rpc
```

### 2. Prepare Test Data

**Option 1: Automated Setup (Recommended)**

Run the setup script to automatically create test data:

```bash
# Using default settings (localhost)
./test/integration/trade/setup_test_data.sh

# Or with custom settings
MYSQL_HOST=localhost MYSQL_PORT=3306 \
MYSQL_USER=aether MYSQL_PASSWORD=aether123 \
REDIS_HOST=localhost REDIS_PORT=6379 \
./test/integration/trade/setup_test_data.sh
```

**Option 2: Manual Setup**

Create a test user in the database:

```sql
INSERT INTO user (id, username, mobile, status, create_time, update_time)
VALUES (1001, 'testuser', '13800138000', 1, NOW(), NOW());
```

Or run the SQL script directly:

```bash
mysql -h localhost -u aether -paether123 aether_defense < test/integration/trade/setup_test_data.sql
```

Set initial inventory in Redis:

```bash
redis-cli SET "inventory:course:5001" "100"
redis-cli SET "inventory:course:5003" "100"
redis-cli SET "inventory:course:5004" "100"
```

### 3. Run Tests

```bash
# Run all trade integration tests
go test -tags=integration ./test/integration/trade -v

# Run specific test
go test -tags=integration ./test/integration/trade -v -run TestIntegration_PlaceOrder_Success
```

## Test Scenarios

### Success Cases

1. **PlaceOrder_Success**: 
   - Valid user places order with valid course IDs
   - Order is created in database
   - RocketMQ message is sent
   - Returns success response

2. **CancelOrder_Success**:
   - User cancels their own pending order
   - Order status updated to Closed
   - Returns success response

### Error Cases

1. **PlaceOrder_InvalidUser**:
   - Non-existent user tries to place order
   - Returns error with user validation message

2. **PlaceOrder_EmptyCourseIds**:
   - User tries to place order with empty course list
   - Returns validation error

3. **CancelOrder_InvalidUser**:
   - User tries to cancel order belonging to another user
   - Returns authorization error

## Troubleshooting

### Tests Skip Automatically

If tests skip with "Required services not running":
- Verify services are running: `docker ps` or check service logs
- Check service addresses match environment variables
- Ensure ports are not blocked by firewall

### Database Connection Issues

- Verify MySQL is accessible
- Check database credentials in service configs
- Ensure test user exists in database

### RocketMQ Issues

- Verify RocketMQ NameServer and Broker are running
- Check RocketMQ configuration in trade-rpc config
- Review RocketMQ logs for connection errors

### Redis Connection Issues

- Verify Redis is accessible
- Check Redis address and database number
- Ensure inventory keys are set correctly

## Expected Test Results

All tests should pass when:
- All services are running
- Test data is prepared
- Network connectivity is available

Tests will skip gracefully if services are unavailable, so they won't fail in CI/CD when services aren't running.

