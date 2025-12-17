#!/bin/bash
# Start local development environment
# This script starts all infrastructure services needed for local testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.yml"

echo "🚀 Starting Aether Defense System Local Development Environment"
echo ""

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null && ! command -v docker &> /dev/null; then
    echo "❌ Error: docker-compose or docker is not installed"
    exit 1
fi

# Use docker compose (newer) or docker-compose (older)
if command -v docker &> /dev/null && docker compose version &> /dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
elif command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
else
    echo "❌ Error: docker compose is not available"
    exit 1
fi

# Start only infrastructure services (for local Go development)
echo "📦 Starting infrastructure services..."
$COMPOSE_CMD -f "$COMPOSE_FILE" up -d etcd redis mysql rocketmq-nameserver rocketmq-broker

echo ""
echo "⏳ Waiting for services to be healthy..."

# Wait for etcd
echo -n "  - etcd: "
timeout=30
while [ $timeout -gt 0 ]; do
    if docker exec aether-defense-etcd etcdctl endpoint health &> /dev/null 2>&1; then
        echo "✅"
        break
    fi
    echo -n "."
    sleep 1
    timeout=$((timeout - 1))
done
if [ $timeout -eq 0 ]; then
    echo "❌ (timeout)"
fi

# Wait for Redis
echo -n "  - Redis: "
timeout=30
while [ $timeout -gt 0 ]; do
    if docker exec aether-defense-redis redis-cli ping &> /dev/null 2>&1; then
        echo "✅"
        break
    fi
    echo -n "."
    sleep 1
    timeout=$((timeout - 1))
done
if [ $timeout -eq 0 ]; then
    echo "❌ (timeout)"
fi

# Wait for MySQL
echo -n "  - MySQL: "
timeout=60
while [ $timeout -gt 0 ]; do
    if docker exec aether-defense-mysql mysqladmin ping -h localhost -u root -proot123 &> /dev/null 2>&1; then
        echo "✅"
        break
    fi
    echo -n "."
    sleep 2
    timeout=$((timeout - 2))
done
if [ $timeout -eq 0 ]; then
    echo "❌ (timeout)"
fi

# Wait for RocketMQ NameServer
echo -n "  - RocketMQ NameServer: "
timeout=30
while [ $timeout -gt 0 ]; do
    if docker exec aether-defense-rocketmq-nameserver sh -c "netstat -tuln | grep 9876" &> /dev/null 2>&1; then
        echo "✅"
        break
    fi
    echo -n "."
    sleep 1
    timeout=$((timeout - 1))
done
if [ $timeout -eq 0 ]; then
    echo "❌ (timeout)"
fi

echo ""
echo "✅ Local development environment is ready!"
echo ""
echo "📍 Service Endpoints:"
echo "   - etcd:              localhost:2379"
echo "   - Redis:             localhost:6379"
echo "   - MySQL:             localhost:3306"
echo "     Username:          root / aether"
echo "     Password:          root123 / aether123"
echo "     Database:          aether_defense"
echo "   - RocketMQ NameServer: localhost:9876"
echo "   - RocketMQ Broker:     localhost:10911"
echo "   - RocketMQ Console:    http://localhost:8080"
echo ""
echo "📝 Next steps:"
echo "   1. Update service configs to use local endpoints (127.0.0.1)"
echo "   2. Run your services: go run ./cmd/rpc/user-rpc"
echo ""
echo "🛑 To stop services:"
echo "   $COMPOSE_CMD -f $COMPOSE_FILE down"
echo ""
