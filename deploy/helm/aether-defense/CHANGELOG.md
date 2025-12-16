# Helm Chart Changelog

## Production-Ready Improvements

This Helm chart has been upgraded to follow production deployment best practices as defined in `.cursor/rules/13-18-production-deployment-*.mdc`.

### ✅ Completed Improvements

#### 1. Environment Isolation
- ✅ Added namespace configuration support
- ✅ Created environment-specific values files:
  - `values-dev.yaml` - Development environment
  - `values-staging.yaml` - Staging environment  
  - `values-prod.yaml` - Production environment
- ✅ Added NetworkPolicy template for pod-to-pod communication control
- ✅ Added ServiceAccount and RBAC templates for each service

#### 2. Resource Limits & Security
- ✅ Added ResourceQuota template for namespace-level resource limits
- ✅ Added LimitRange template for default container limits
- ✅ Added SecurityContext configuration:
  - `runAsNonRoot: true`
  - `readOnlyRootFilesystem: true` (configurable)
  - `allowPrivilegeEscalation: false`
  - Capabilities dropped to minimum
- ✅ Fixed image tag warnings (documented that production should use specific versions)

#### 3. High Availability & Disaster Recovery
- ✅ Added PodDisruptionBudget (PDB) for all services with multiple replicas
- ✅ Added Pod Anti-Affinity configuration to distribute pods across nodes
- ✅ Enhanced health checks:
  - Added `startupProbe` for slow-starting applications
  - Added `timeoutSeconds`, `failureThreshold`, `successThreshold`
- ✅ Added RollingUpdate strategy configuration:
  - `maxUnavailable: 0` for zero-downtime deployments
  - `maxSurge: 1` for controlled rollout speed

#### 4. Autoscaling
- ✅ Added HorizontalPodAutoscaler (HPA) template
- ✅ Configurable CPU and memory-based autoscaling
- ✅ Configurable min/max replicas per service

#### 5. Monitoring
- ✅ Added ServiceMonitor template for Prometheus integration
- ✅ Configurable monitoring settings

#### 6. Bug Fixes
- ✅ Fixed etcd service name references in ConfigMaps (now uses full service name)
- ✅ Fixed Kubernetes naming conventions (all resource names now lowercase)
- ✅ Fixed ServiceAccount references in Deployments

### 📋 New Configuration Options

#### Global Configuration
```yaml
global:
  namespace: ""  # Environment-specific namespace
  imageRegistry: ""
  imagePullPolicy: IfNotPresent
```

#### Security Configuration
```yaml
security:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

#### Resource Quotas
```yaml
resourceQuota:
  enabled: false
  requests:
    cpu: "20"
    memory: 40Gi
  limits:
    cpu: "40"
    memory: 80Gi
```

#### Service-Specific Configuration
Each service now supports:
- `pdb.minAvailable` - Pod Disruption Budget
- `podAntiAffinity` - Enable pod anti-affinity
- `autoscaling` - HPA configuration
- `strategy` - Rolling update strategy

### 🚀 Deployment Commands

#### Development
```bash
helm install aether-defense ./deploy/helm/aether-defense \
  -f deploy/helm/aether-defense/values.yaml \
  -f deploy/helm/aether-defense/values-dev.yaml \
  -n aether-defense-dev --create-namespace
```

#### Staging
```bash
helm install aether-defense ./deploy/helm/aether-defense \
  -f deploy/helm/aether-defense/values.yaml \
  -f deploy/helm/aether-defense/values-staging.yaml \
  -n aether-defense-staging --create-namespace
```

#### Production
```bash
helm install aether-defense ./deploy/helm/aether-defense \
  -f deploy/helm/aether-defense/values.yaml \
  -f deploy/helm/aether-defense/values-prod.yaml \
  -n aether-defense-prod --create-namespace
```

### ⚠️ Important Notes

1. **Image Tags**: Production deployments should use specific image tags, not `latest`
2. **Replica Counts**: Production requires minimum 3 replicas for HA
3. **Security**: Production enables strict security settings (read-only root filesystem, etc.)
4. **Resource Quotas**: Production enables resource quotas to prevent resource exhaustion
5. **Network Policies**: Production enables network policies for pod-to-pod communication control

### 📚 References

See `.cursor/rules/` for detailed production deployment guidelines:
- `13-production-deployment-isolation.mdc` - Environment isolation
- `14-production-deployment-security.mdc` - Security and resource limits
- `15-production-deployment-ha.mdc` - High availability
- `16-production-deployment-testing.mdc` - Testing strategies
- `17-production-deployment-canary.mdc` - Canary deployment
- `18-production-deployment-checklist.mdc` - Deployment checklist
