## Aether Defense System

A Go-based high-concurrency demo system built on **go-zero**, showcasing service decomposition,
RPC communication, and an HTTP API gateway for user, trade, and promotion domains.

### Tech Stack

- **Language**: Go
- **Framework**: [go-zero](https://go-zero.dev/)
- **RPC**: gRPC + go-zero `zrpc`
- **Config**: YAML
- **Others**: Snowflake ID, MQ, Prometheus (introduced gradually in submodules)

### Project Layout

- `cmd/`
  - `api/user-api/`: User HTTP API entrypoint, routing, and minimal business logic.
  - `rpc/user-rpc/`: User RPC service entrypoint.
  - Other services’ API/RPC entrypoints follow the same pattern.
- `service/`
  - `user/rpc/`: User-domain RPC service (business logic, config, server wiring).
  - `trade/rpc/`: Trade-domain RPC service (place order, cancel order, etc.).
  - `promotion/rpc/`: Promotion / inventory-related RPC service.
- `common/`: Shared middleware, MQ wrappers, Snowflake ID generator, etc.
- `deploy/`: Docker / K8s / Prometheus and other deployment-related configs.
- `doc/`: Architecture design, coding standards, performance guidelines, etc.

### Getting Started (Local Development)

#### 1. Prerequisites

- Go 1.22+ (recommended to manage via `goenv` or `gvm`).
- `protoc` and go-zero codegen plugins installed (only needed if you regenerate proto code).

#### 2. Start the User RPC Service

From the project root:

```bash
go run ./cmd/rpc/user-rpc
```

By default it loads `service/user/rpc/etc/user.yaml` and listens on `0.0.0.0:8080`.

#### 3. Start the User HTTP API

From the project root:

```bash
go run ./service/user/api/cmd/user-api
```

It loads `service/user/api/etc/user-api.yaml`, listens on `0.0.0.0:8888`,
and calls the local `user.rpc` directly.

#### 4. Verify the Endpoint

```bash
curl "http://localhost:8888/v1/users/1"
```

You should see a JSON response containing `userId`, `username`, and `mobile`.
For now this is a deterministic stub user; persistence and real data sources will
be added in future iterations.

### Testing

From the project root:

```bash
go test ./...
```

Core business logic (user, trade, promotion RPC logic) is covered by unit tests
that validate input and basic flows.

### Development Guidelines & Architecture Docs

- Coding standards, commit/PR workflow and review rules: see `doc/coding-standards/`.
- Overall architecture, service boundaries and design principles: see `doc/design/`.

When adding new services or endpoints, please follow the patterns used in
`user`, `trade`, and `promotion` services to keep layering and style consistent.
