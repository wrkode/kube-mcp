# Best Practices for kube-mcp

## Overview

This document provides best practices, recommendations, and patterns for using kube-mcp effectively and securely in production environments.

## Security Best Practices

### 1. Enable RBAC Checks

Always enable RBAC checks in production:

```toml
[security]
require_rbac = true
rbac_cache_ttl = 5  # Cache TTL in seconds
```

**Benefits:**
- Prevents unauthorized operations
- Validates permissions before execution
- Provides audit trail

### 2. Use Single-Cluster Provider for Production

For production deployments, use the single-cluster provider to prevent accidental cluster switches:

```toml
[kubernetes]
provider = "single"
kubeconfig_path = "/etc/kube-mcp/kubeconfig"
context = "production-cluster"  # Locked to this context
```

**Benefits:**
- Prevents accidental cluster switches
- Reduces attack surface
- Simplifies security model

### 3. Restrict Kubeconfig Access

Limit kubeconfig file permissions:

```bash
chmod 600 ~/.kube/config
```

**Benefits:**
- Prevents unauthorized access
- Protects credentials
- Follows principle of least privilege

### 4. Use Service Accounts in Kubernetes

When running in-cluster, use dedicated service accounts with minimal required permissions:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-mcp
  namespace: kube-mcp
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-mcp-reader
rules:
  - apiGroups: [""]
    resources: ["pods", "namespaces"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kube-mcp-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-mcp-reader
subjects:
  - kind: ServiceAccount
    name: kube-mcp
    namespace: kube-mcp
```

**Benefits:**
- Principle of least privilege
- Reduced blast radius
- Better auditability

### 5. Enable Token Validation

For HTTP transport, enable token validation:

```toml
[server.http.oauth]
enabled = true
provider = "oidc"
issuer_url = "https://auth.example.com"
client_id = "kube-mcp"
client_secret = "${OAUTH_CLIENT_SECRET}"

[security]
validate_token = true  # Validate Bearer tokens using TokenReview
```

**Benefits:**
- Validates user identity
- Integrates with Kubernetes authentication
- Provides audit trail

### 6. Use Read-Only Mode for Audit

For audit and inspection purposes, use read-only mode:

```toml
[security]
read_only = true
require_rbac = true
```

**Benefits:**
- Prevents accidental modifications
- Safe for audit tools
- Reduces risk

## Configuration Best Practices

### 1. Use Drop-In Configuration

Organize configuration using drop-in files:

```
/etc/kube-mcp/
├── config.toml          # Base configuration
└── conf.d/
    ├── 10-security.toml # Security overrides
    ├── 20-kiali.toml    # Kiali configuration
    └── 30-custom.toml    # Custom settings
```

**Benefits:**
- Better organization
- Easier maintenance
- Supports hot reload

### 2. Use Environment Variables

Use environment variables for sensitive values:

```toml
[server.http.oauth]
client_secret = "${OAUTH_CLIENT_SECRET}"

[kiali]
token = "${KIALI_TOKEN}"
```

**Benefits:**
- Avoids hardcoding secrets
- Easier secret rotation
- Better security

### 3. Configure Rate Limiting

Set appropriate rate limits:

```toml
[kubernetes]
qps = 100        # Queries per second
burst = 200      # Burst limit
timeout = "30s"  # Request timeout
```

**Benefits:**
- Prevents API server overload
- Better performance
- More predictable behavior

### 4. Enable Observability

Enable logging and metrics:

```toml
[server]
log_level = "info"  # Use "debug" for troubleshooting

# Metrics are automatically enabled
```

**Benefits:**
- Better troubleshooting
- Performance monitoring
- Audit trail

## Operational Best Practices

### 1. Always Validate Before Applying

Validate manifests before applying:

```json
// Step 1: Validate
{
  "tool": "resources_validate",
  "params": {
    "manifest": { /* your manifest */ }
  }
}

// Step 2: Dry-run
{
  "tool": "resources_apply",
  "params": {
    "namespace": "default",
    "manifest": { /* your manifest */ },
    "dry_run": true
  }
}

// Step 3: Apply
{
  "tool": "resources_apply",
  "params": {
    "namespace": "default",
    "manifest": { /* your manifest */ }
  }
}
```

**Benefits:**
- Catches errors early
- Prevents bad deployments
- Safer operations

### 2. Use Dry-Run for Destructive Operations

Always use dry-run for destructive operations:

```json
// Delete with dry-run
{
  "tool": "resources_delete",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "nginx-deployment",
    "namespace": "default",
    "dry_run": true
  }
}
```

**Benefits:**
- Validates operation
- Shows what would be deleted
- Prevents accidents

### 3. Use Label Selectors

Use label selectors for efficient filtering:

```json
{
  "tool": "pods_list",
  "params": {
    "namespace": "default",
    "label_selector": "app=nginx,version=1.21"
  }
}
```

**Benefits:**
- More efficient queries
- Better performance
- Cleaner results

### 4. Use Pagination for Large Result Sets

Use pagination for large result sets:

```json
{
  "tool": "resources_list",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "",
    "limit": 100,
    "continue": ""  // Use continue token from previous response
  }
}
```

**Benefits:**
- Handles large clusters
- Better memory usage
- More predictable performance

### 5. Monitor with Events

Use events for monitoring:

```json
{
  "tool": "events_list",
  "params": {
    "namespace": "default",
    "field_selector": "involvedObject.name=nginx-deployment"
  }
}
```

**Benefits:**
- Real-time monitoring
- Troubleshooting
- Audit trail

### 6. Use Resource Watching

Watch resources for real-time updates:

```json
{
  "tool": "resources_watch",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "default",
    "timeout": 60
  }
}
```

**Benefits:**
- Real-time updates
- Event-driven workflows
- Better monitoring

## Multi-Cluster Best Practices

### 1. Explicitly Specify Context

Always specify context explicitly in multi-cluster scenarios:

```json
{
  "tool": "pods_list",
  "params": {
    "context": "prod-cluster",  // Explicit context
    "namespace": "production"
  }
}
```

**Benefits:**
- Prevents accidental cluster switches
- Clearer intent
- Better debugging

### 2. Verify Context Before Operations

Verify context exists before operations:

```json
// Step 1: List contexts
{
  "tool": "config_contexts_list",
  "params": {}
}

// Step 2: Use verified context
{
  "tool": "pods_list",
  "params": {
    "context": "prod-cluster",
    "namespace": "production"
  }
}
```

**Benefits:**
- Prevents errors
- Better error handling
- Clearer workflows

### 3. Use Single-Cluster Provider for Dedicated Deployments

For dedicated cluster deployments, use single-cluster provider:

```toml
[kubernetes]
provider = "single"
kubeconfig_path = "/etc/kube-mcp/kubeconfig"
context = "production-cluster"
```

**Benefits:**
- Simpler configuration
- Better security
- Prevents mistakes

## Performance Best Practices

### 1. Configure Appropriate Rate Limits

Set rate limits based on cluster size:

```toml
# Small cluster (< 100 nodes)
[kubernetes]
qps = 50
burst = 100

# Medium cluster (100-500 nodes)
[kubernetes]
qps = 100
burst = 200

# Large cluster (> 500 nodes)
[kubernetes]
qps = 200
burst = 400
```

### 2. Enable RBAC Caching

Enable RBAC caching for better performance:

```toml
[security]
require_rbac = true
rbac_cache_ttl = 5  # 5 seconds cache
```

**Benefits:**
- Reduces API calls
- Better performance
- Lower latency

### 3. Use Field Selectors

Use field selectors for efficient filtering:

```json
{
  "tool": "events_list",
  "params": {
    "namespace": "default",
    "field_selector": "involvedObject.name=nginx-deployment"
  }
}
```

**Benefits:**
- Server-side filtering
- Less data transfer
- Better performance

### 4. Batch Operations

Batch operations when possible:

```json
// Instead of multiple calls, use label selectors
{
  "tool": "resources_list",
  "params": {
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "default",
    "label_selector": "app=nginx"
  }
}
```

## Error Handling Best Practices

### 1. Always Check for Errors

Always check `isError` field:

```json
{
  "isError": true,
  "content": [
    {
      "type": "text",
      "text": "{\"error\":{\"type\":\"KubernetesError\",\"message\":\"...\"}}"
    }
  ]
}
```

### 2. Handle Specific Error Types

Handle specific error types appropriately:

- `KubernetesError` - Kubernetes API errors
- `UnknownContext` - Cluster context not found
- `MetricsUnavailable` - Metrics server not available
- `ScalingNotSupported` - Resource doesn't support scaling

### 3. Retry Transient Errors

Retry transient errors:

```json
{
  "error": {
    "type": "KubernetesError",
    "message": "etcdserver: request timed out"
  }
}
```

**Benefits:**
- Handles transient failures
- Better reliability
- Improved user experience

## Monitoring Best Practices

### 1. Enable Metrics

Metrics are automatically enabled. Access via:

```bash
curl http://localhost:8080/metrics
```

### 2. Monitor Health Endpoint

Monitor health endpoint:

```bash
curl http://localhost:8080/health
```

### 3. Use Structured Logging

Use structured logging:

```toml
[server]
log_level = "info"  # Use "debug" for troubleshooting
```

### 4. Monitor Resource Usage

Monitor resource usage:

```json
{
  "tool": "pods_top",
  "params": {
    "namespace": "default"
  }
}
```

## Deployment Best Practices

### 1. Use Helm Chart for Kubernetes

Use Helm chart for Kubernetes deployments:

```bash
helm install kube-mcp kube-mcp/kube-mcp
```

**Benefits:**
- Standardized deployment
- Easy upgrades
- Configuration management

### 2. Use Distroless Container

Use distroless container for production:

```dockerfile
FROM gcr.io/distroless/static-debian12:nonroot
```

**Benefits:**
- Minimal attack surface
- Smaller image size
- Better security

### 3. Set Resource Limits

Set resource limits:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### 4. Use Health Probes

Use health probes:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
```

## Summary

Key best practices:

1. **Security**: Enable RBAC, use single-cluster provider, restrict access
2. **Configuration**: Use drop-in configs, environment variables, rate limiting
3. **Operations**: Validate, dry-run, use selectors, pagination
4. **Multi-Cluster**: Specify context explicitly, verify before operations
5. **Performance**: Configure rate limits, enable caching, use selectors
6. **Error Handling**: Check errors, handle types, retry transient errors
7. **Monitoring**: Enable metrics, monitor health, use structured logging
8. **Deployment**: Use Helm, distroless, resource limits, health probes

Following these best practices will help you:
- Operate kube-mcp securely
- Achieve better performance
- Improve reliability
- Simplify maintenance
- Enhance observability

