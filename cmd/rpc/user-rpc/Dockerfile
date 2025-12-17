# syntax=docker/dockerfile:1
# Ultra-slim optimized build
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install minimal build dependencies
RUN apk add --no-cache git ca-certificates

# Set Go environment for maximum optimization
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GOCACHE=/go-cache
ENV GOMODCACHE=/go-mod-cache
ENV CGO_ENABLED=0
ENV GOOS=linux

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go-mod-cache \
    go mod download

# Copy source code
COPY . .

# Build with maximum optimization
WORKDIR /build/cmd/rpc/user-rpc
RUN --mount=type=cache,target=/go-mod-cache \
    --mount=type=cache,target=/go-cache \
    go build \
    -ldflags="-w -s -X main.version=1.0.0" \
    -trimpath \
    -buildmode=exe \
    -o user-rpc .

# Use distroless for minimal runtime
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary and config
COPY --from=builder /build/cmd/rpc/user-rpc/user-rpc /app/user-rpc
COPY service/user/rpc/etc/user.yaml /app/user.yaml

# Set working directory and user
WORKDIR /app
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/app/user-rpc", "-f", "/etc/user-rpc/user.yaml"]
