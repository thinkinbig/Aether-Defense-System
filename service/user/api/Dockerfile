# syntax=docker/dockerfile:1
# Ultra-slim optimized build
FROM golang:1.25-alpine AS builder

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
WORKDIR /build/service/user/api/cmd/user-api
RUN --mount=type=cache,target=/go-mod-cache \
    --mount=type=cache,target=/go-cache \
    go build \
    -ldflags="-w -s -X main.version=1.0.0" \
    -trimpath \
    -buildmode=exe \
    -o user-api .

# Use distroless for minimal runtime
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary and config
COPY --from=builder /build/service/user/api/cmd/user-api/user-api /app/user-api
COPY service/user/api/etc/user-api.yaml /app/user-api.yaml

# Set working directory and user
WORKDIR /app
USER nonroot:nonroot

# Expose port
EXPOSE 8888

# Run the binary
ENTRYPOINT ["/app/user-api", "-f", "/app/user-api.yaml"]
