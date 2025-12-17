# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build trade-rpc
WORKDIR /build/service/trade/rpc
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o trade-rpc .

# Runtime stage
FROM alpine:latest

# Fix for ARM64 builds - install packages without triggers
RUN apk --no-cache add ca-certificates tzdata --no-scripts
ENV TZ=Europe/Berlin

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/service/trade/rpc/trade-rpc /app/trade-rpc

# Copy config file (will be overridden by ConfigMap mount)
COPY service/trade/rpc/etc/trade.yaml /app/trade.yaml

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/trade-rpc", "-f", "/etc/trade-rpc/trade.yaml"]
