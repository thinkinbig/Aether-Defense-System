# Configuration Management Across Environments

This document explains how configuration files are managed across different deployment environments.

## Configuration Strategy

The system uses different configuration files for different environments:

### 1. Local Development (`service/*/etc/*.yaml`)

**Purpose**: Default configs for running services locally with `go run`

**Location**: `service/{service}/rpc/etc/{service}.yaml`

**Configuration**:
- etcd: `127.0.0.1:2379`
- Database: `127.0.0.1:3306`
- Redis: `127.0.0.1:6379`
- RocketMQ: `127.0.0.1:9876`

**Usage**:
```bash
go run ./cmd/rpc/user-rpc
# Uses: service/user/rpc/etc/user.yaml
```

### 2. Docker Compose (`deploy/docker/config-docker/*.yaml`)

**Purpose**: Configs for Docker Compose deployments using service names

**Location**: `deploy/docker/config-docker/*.yaml`

**Configuration**:
- etcd: `etcd:2379` (Docker service name)
- Database: `mysql:3306` (Docker service name)
- Redis: `redis:6379` (Docker service name)
- RocketMQ: `rocketmq-nameserver:9876` (Docker service name)

**Usage**:
```bash
docker-compose -f deploy/docker/docker-compose.yml up
# Mounts config-docker/*.yaml files as volumes, overriding default configs
```

**How it works**:
- Docker Compose mounts `config-docker/*.yaml` files to `/app/{service}.yaml`
- Environment variable `CONFIG_FILE` points to the mounted config
- This overrides the default config copied in Dockerfile

### 3. Kubernetes (`deploy/helm/aether-defense/templates/*/configmap.yaml`)

**Purpose**: Configs generated from Helm values for K8s deployments

**Location**: Generated ConfigMaps from Helm templates

**Configuration**:
- etcd: `{release-name}-etcd-headless:2379` (K8s service name)
- Database: K8s service names
- Redis: K8s service names
- RocketMQ: K8s service names

**Usage**:
```bash
helm install aether-defense ./deploy/helm/aether-defense
# Generates ConfigMaps from values.yaml and templates
```

**How it works**:
- Helm templates generate ConfigMaps from `values.yaml`
- ConfigMaps are mounted to `/etc/{service}/` in pods
- Services use `-f /etc/{service}/{service}.yaml` flag

## Configuration File Hierarchy

```
service/user/rpc/etc/user.yaml          # Local dev (localhost)
    ↓ (overridden by)
deploy/docker/config-docker/user-rpc.yaml  # Docker (service names)
    ↓ (overridden by)
K8s ConfigMap (from Helm)                  # K8s (K8s service names)
```

## Best Practices

1. **Never commit K8s-specific configs to service directories**
   - Service directories should only have local dev configs
   - K8s configs are generated from Helm templates

2. **Docker configs are separate**
   - Keep Docker configs in `deploy/docker/config-docker/`
   - Use Docker service names (e.g., `etcd:2379`)

3. **Local dev configs use localhost**
   - Use `127.0.0.1` or `localhost` for local development
   - Assume services are running on the same machine

4. **Environment-specific overrides**
   - Use environment variables when possible
   - Use volume mounts in Docker
   - Use ConfigMaps in K8s

## Updating Configurations

### For Local Development
Edit: `service/{service}/rpc/etc/{service}.yaml`

### For Docker Compose
Edit: `deploy/docker/config-docker/{service}.yaml`

### For Kubernetes
Edit: `deploy/helm/aether-defense/values.yaml` or environment-specific values files

## Verification

To verify which config is being used:

1. **Local**: Check the file path in the service startup command
2. **Docker**: Check the volume mount in `docker-compose.yml`
3. **K8s**: Check the ConfigMap and volume mount in the deployment

