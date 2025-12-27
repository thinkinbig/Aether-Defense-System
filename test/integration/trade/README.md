# Trade Service Integration Tests

This directory contains integration tests for the trade service, verifying the complete order placement and cancellation flow.

## Test Cases

### Order Placement Tests

- **TestIntegration_PlaceOrder_Success**: Tests successful order placement
  - Verifies order creation in database
  - Verifies RocketMQ message sending
  - Verifies user validation
  
- **TestIntegration_PlaceOrder_InvalidUser**: Tests order placement with invalid user
  - Verifies error handling for non-existent users
  
- **TestIntegration_PlaceOrder_EmptyCourseIds**: Tests order placement with empty course list
  - Verifies validation of required fields

### Order Cancellation Tests

- **TestIntegration_CancelOrder_Success**: Tests successful order cancellation
  - Verifies order status update
  - Verifies ownership validation
  
- **TestIntegration_CancelOrder_InvalidUser**: Tests cancellation with wrong user
  - Verifies authorization checks

## Running Tests

### Prerequisites

1. **Services must be running**:
   - MySQL (for order storage)
   - Redis (for inventory)
   - RocketMQ (for transactional messages)
   - Trade RPC service (on port 8081)
   - User RPC service (on port 8080)

2. **Test user must exist**:
   - Create a test user with ID 1001 in the database
   - Or modify the test to use an existing user ID

### Run Tests

```bash
# Run all trade integration tests
go test -tags=integration ./test/integration/trade -v

# Run specific test
go test -tags=integration ./test/integration/trade -v -run TestIntegration_PlaceOrder_Success

# With custom service addresses
TRADE_RPC_ADDR=localhost:8081 USER_RPC_ADDR=localhost:8080 go test -tags=integration ./test/integration/trade -v
```

### Environment Variables

- `TRADE_RPC_ADDR`: Trade RPC service address (default: localhost:8081)
- `USER_RPC_ADDR`: User RPC service address (default: localhost:8080)
- `REDIS_ADDR`: Redis server address (default: localhost:6379)
- `REDIS_DB`: Redis database number (default: 0)
- `REDIS_PASSWORD`: Redis password (default: empty)

## Test Data

Tests use dynamic order IDs based on timestamps to avoid conflicts. Test course IDs:
- 5001: Used in success test
- 5002: Used in invalid user test
- 5003: Used in cancellation test

## Notes

- Tests will skip if services are not running
- Tests require a test user (ID 1001) to exist in the database
- Inventory deduction happens asynchronously via RocketMQ, so tests may need to wait for message consumption
- Tests use real services, not mocks, to verify end-to-end integration

