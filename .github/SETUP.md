# GitHub Actions CI/CD Setup Guide

## Required GitHub Secrets

To make the CI/CD pipeline work properly, you need to configure the following secrets in your GitHub repository settings:

### 1. Kubernetes Cluster Configuration

Add the following in GitHub repository Settings > Secrets and variables > Actions:

```
KUBE_CONFIG_DEV      # kubeconfig file content for development environment
KUBE_CONFIG_STAGING  # kubeconfig file content for staging environment  
KUBE_CONFIG_PROD     # kubeconfig file content for production environment
```

### 2. Container Image Registry

GitHub Container Registry (ghcr.io) will automatically use `GITHUB_TOKEN`, no additional configuration needed.

## Environment Configuration

### 1. Update Image Registry in Helm Values

In `deploy/helm/aether-defense/values.yaml`, change:

```yaml
global:
  imageRegistry: "ghcr.io/your-github-username/aether-defense-system"
```

Replace `your-github-username` with your actual GitHub username or organization name.

### 2. Environment Deployment Configuration

Ensure your Kubernetes cluster has the following namespaces:
- `aether-defense-dev` (development environment)
- `aether-defense-staging` (staging environment)
- `aether-defense-prod` (production environment)

### 3. Branch Strategy

- `develop` branch push → automatically deploy to development environment
- `v*` tag push → automatically deploy to staging and production environments
- Pull Request → run CI checks

## Workflow Description

### CI Process (`.github/workflows/ci.yml`)
1. Code formatting check and linting
2. Unit tests and coverage check
3. Multi-platform build verification

### CD Process (`.github/workflows/cd.yml`)
1. Docker image build and push to GitHub Container Registry
2. Security vulnerability scanning
3. Automatic deployment to corresponding environments
4. Health checks and smoke tests

## Usage Steps

1. **Configure Secrets**: Add necessary secrets in GitHub repository settings
2. **Update Image Registry**: Modify `imageRegistry` in `values.yaml`
3. **Push Code**:
   - Push to `develop` branch to trigger development environment deployment
   - Create `v1.0.0` format tags to trigger production environment deployment

## Image Tag Strategy

- `latest`: Latest version of main branch
- `develop`: Latest version of develop branch
- `v1.0.0`: Semantic version tags
- `main-abc1234`: Branch name-commit SHA

## Monitoring and Logs

After deployment is complete, you can monitor the application through the following methods:

```bash
# Check Pod status
kubectl get pods -n aether-defense-dev

# Check service logs
kubectl logs -f deployment/aether-defense-dev-user-api -n aether-defense-dev

# Check service status
kubectl get svc -n aether-defense-dev
```

## Troubleshooting

### 1. Image Pull Failure
Ensure GitHub Container Registry permissions are set correctly, images should be public or the cluster has correct pull secrets.

### 2. Deployment Timeout
Check if resource configuration is reasonable, ensure the cluster has sufficient resources.

### 3. Health Check Failure
Check if the application's health check endpoints are correctly implemented.
