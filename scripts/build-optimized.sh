#!/bin/bash

# Optimized Docker build script with BuildKit cache
set -e

# Enable BuildKit
export DOCKER_BUILDKIT=1

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Starting ultra-slim Docker builds with BuildKit cache...${NC}"

# Services to build
SERVICES=(
    "cmd/api/user-api:user-api"
    "cmd/rpc/user-rpc:user-rpc"
    "service/trade/rpc:trade-rpc"
    "service/promotion/rpc:promotion-rpc"
)

# Build each service with cache
for service in "${SERVICES[@]}"; do
    IFS=':' read -r dockerfile_path service_name <<< "$service"

    echo -e "${YELLOW}📦 Building ${service_name}...${NC}"

    # Build with cache mounts and registry cache
    docker build \
        --file "${dockerfile_path}/Dockerfile" \
        --tag "aether-defense/${service_name}:latest" \
        --cache-from "aether-defense/${service_name}:cache" \
        --build-arg BUILDKIT_INLINE_CACHE=1 \
        .

    echo -e "${GREEN}✅ ${service_name} built successfully${NC}"
done

echo -e "${GREEN}🎉 All services built successfully with ultra-slim optimization!${NC}"

# Show cache usage
echo -e "${YELLOW}📊 Docker system info:${NC}"
docker system df
