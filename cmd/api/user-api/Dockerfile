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

# Build user-api
WORKDIR /build/cmd/api/user-api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-api .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cmd/api/user-api/user-api /app/user-api

# Copy config file (will be overridden by ConfigMap mount)
COPY cmd/api/user-api/etc/user-api.yaml /app/user-api.yaml

# Expose port
EXPOSE 8888

# Run the binary
ENTRYPOINT ["/app/user-api", "-f", "/etc/user-api/user-api.yaml"]
