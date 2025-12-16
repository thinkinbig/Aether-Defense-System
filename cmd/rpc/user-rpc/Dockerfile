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

# Build user-rpc
WORKDIR /build/cmd/rpc/user-rpc
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-rpc .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cmd/rpc/user-rpc/user-rpc /app/user-rpc

# Copy config file (will be overridden by ConfigMap mount)
COPY service/user/rpc/etc/user.yaml /app/user.yaml

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/user-rpc", "-f", "/etc/user-rpc/user.yaml"]
