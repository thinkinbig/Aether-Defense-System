#!/bin/bash
# Stop local development environment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.yml"

echo "🛑 Stopping Aether Defense System Local Development Environment"
echo ""

# Use docker compose (newer) or docker-compose (older)
if command -v docker &> /dev/null && docker compose version &> /dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
elif command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
else
    echo "❌ Error: docker compose is not available"
    exit 1
fi

# Stop services
$COMPOSE_CMD -f "$COMPOSE_FILE" down

echo "✅ Services stopped"
echo ""
echo "💡 To remove all data (volumes):"
echo "   $COMPOSE_CMD -f $COMPOSE_FILE down -v"
echo ""
