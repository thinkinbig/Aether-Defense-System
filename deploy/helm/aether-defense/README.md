# Aether Defense System Helm Chart

This Helm chart deploys the Aether Defense System, including etcd for service discovery and all microservices.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Minikube or a Kubernetes cluster

## Installation

### Install with default values

```bash
helm install aether-defense ./deploy/helm/aether-defense
```

### Install with custom values

```bash
helm install aether-defense ./deploy/helm/aether-defense -f my-values.yaml
```

### Upgrade existing release

```bash
helm upgrade aether-defense ./deploy/helm/aether-defense
```

## Configuration

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `etcd.enabled` | Enable etcd deployment | `true` |
| `etcd.replicaCount` | Number of etcd replicas | `3` |
| `userRpc.enabled` | Enable user RPC service | `true` |
| `userRpc.replicaCount` | Number of user RPC replicas | `2` |
| `userApi.enabled` | Enable user API service | `true` |
| `userApi.replicaCount` | Number of user API replicas | `2` |

## Components

### etcd

- **StatefulSet**: 3 replicas for high availability
- **Service**: Headless service for StatefulSet communication
- **Storage**: Persistent volumes (8Gi each by default)

### User RPC

- **Deployment**: Configurable replicas
- **Service**: ClusterIP service on port 8080
- **Config**: Automatically configured to use etcd for service discovery

### User API

- **Deployment**: Configurable replicas
- **Service**: ClusterIP service on port 8888
- **Config**: Automatically configured to discover user RPC via etcd

## Testing

After installation, verify the deployment:

```bash
# Check all pods
kubectl get pods

# Check services
kubectl get svc

# Port forward to test user-api
kubectl port-forward svc/aether-defense-user-api 8888:8888

# Test the API
curl http://localhost:8888/v1/users/1
```

## Uninstallation

```bash
helm uninstall aether-defense
```

