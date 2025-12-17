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

# Build promotion-rpc
WORKDIR /build/service/promotion/rpc
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o promotion-rpc .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/service/promotion/rpc/promotion-rpc /app/promotion-rpc

# Copy config file (will be overridden by ConfigMap mount)
COPY service/promotion/rpc/etc/promotion.yaml /app/promotion.yaml

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/promotion-rpc", "-f", "/etc/promotion-rpc/promotion.yaml"]
