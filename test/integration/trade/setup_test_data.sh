#!/bin/bash
# Setup script for trade service integration tests
# This script prepares test data in MySQL and Redis

set -e

# Default values
MYSQL_HOST="${MYSQL_HOST:-localhost}"
MYSQL_PORT="${MYSQL_PORT:-3306}"
MYSQL_USER="${MYSQL_USER:-aether}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-aether123}"
MYSQL_DATABASE="${MYSQL_DATABASE:-aether_defense}"

REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Setting up test data for trade service integration tests..."

# Setup MySQL test user
echo "Creating test user in MySQL..."
mysql -h"${MYSQL_HOST}" -P"${MYSQL_PORT}" -u"${MYSQL_USER}" -p"${MYSQL_PASSWORD}" "${MYSQL_DATABASE}" < "${SCRIPT_DIR}/setup_test_data.sql"

# Setup Redis inventory
echo "Setting initial inventory in Redis..."
redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" SET "inventory:course:5001" "100"
redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" SET "inventory:course:5003" "100"
redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" SET "inventory:course:5004" "100"

echo "Test data setup complete!"
echo ""
echo "Test user: ID=1001, username=testuser"
echo "Inventory keys set:"
echo "  - inventory:course:5001 = 100"
echo "  - inventory:course:5003 = 100"
echo "  - inventory:course:5004 = 100"
