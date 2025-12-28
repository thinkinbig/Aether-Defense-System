#!/bin/bash
# Test script for inventory deduction functionality
# This script tests the DecrStock API through gRPC

set -e

COURSE_ID=${1:-6001}
INITIAL_STOCK=${2:-100}
DEDUCTION_AMOUNT=${3:-30}
GRPC_ADDR=${GRPC_ADDR:-localhost:8082}

echo "=== 库存扣减功能测试 ==="
echo "课程ID: $COURSE_ID"
echo "初始库存: $INITIAL_STOCK"
echo "扣减数量: $DEDUCTION_AMOUNT"
echo ""

# Set initial inventory
echo "1. 设置初始库存..."
docker exec aether-defense-redis redis-cli SET "inventory:course:$COURSE_ID" "$INITIAL_STOCK" > /dev/null
current_stock=$(docker exec aether-defense-redis redis-cli GET "inventory:course:$COURSE_ID")
echo "   当前库存: $current_stock"
echo ""

# Call gRPC API (grpcurl needs proto because server reflection is disabled)
echo "2. 调用 gRPC DecrStock..."
if ! command -v grpcurl >/dev/null 2>&1; then
  echo "   ❌ 未找到 grpcurl，请先安装：go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
  exit 1
fi

cd "$(dirname "$0")/.."
resp=$(grpcurl -plaintext \
  -import-path service/promotion/rpc \
  -proto service/promotion/rpc/promotion.proto \
  -d "{\"courseId\":${COURSE_ID},\"num\":${DEDUCTION_AMOUNT}}" \
  "${GRPC_ADDR}" \
  promotion.PromotionService/DecrStock)
echo "   响应: ${resp}"
echo ""

# Verify final stock
echo "3. 验证最终库存..."
final_stock=$(docker exec aether-defense-redis redis-cli GET "inventory:course:$COURSE_ID")
expected_stock=$((INITIAL_STOCK - DEDUCTION_AMOUNT))
echo "   最终库存: $final_stock"
echo "   期望库存: $expected_stock"

if [ "$final_stock" = "$expected_stock" ]; then
    echo "   ✅ 库存扣减成功！"
    exit 0
else
    echo "   ❌ 库存扣减失败！期望 $expected_stock，实际 $final_stock"
    exit 1
fi
