# syntax=docker/dockerfile:1
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Set Go environment for better caching
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/go-mod-cache

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go-mod-cache \
    go mod download

# Copy source code
COPY . .

# Build user-api with cache mounts and optimizations
WORKDIR /build/cmd/api/user-api
RUN --mount=type=cache,target=/go-mod-cache \
    --mount=type=cache,target=/go-cache \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o user-api .

# Runtime stage
FROM alpine:latest

# Fix for ARM64 builds - install packages without triggers
RUN apk --no-cache add ca-certificates tzdata --no-scripts
ENV TZ=Europe/Berlin

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cmd/api/user-api/user-api /app/user-api

# Copy config file (will be overridden by ConfigMap mount)
COPY cmd/api/user-api/etc/user-api.yaml /app/user-api.yaml

# Expose port
EXPOSE 8888

# Run the binary
ENTRYPOINT ["/app/user-api", "-f", "/etc/user-api/user-api.yaml"]
