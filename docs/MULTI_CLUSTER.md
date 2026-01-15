# Multi-Cluster Management Guide

## Overview

kube-mcp provides comprehensive multi-cluster support, allowing you to manage multiple Kubernetes clusters from a single server instance. This guide covers how to configure, use, and optimize multi-cluster operations.

> **New to kube-mcp?** Start with the [Getting Started Guide](GETTING_STARTED.md) for basic setup.
> 
> **Need usage examples?** See the [Usage Guide](USAGE_GUIDE.md) for practical multi-cluster workflows.

## Provider Types

kube-mcp supports three provider types, each with different multi-cluster capabilities:

### 1. Kubeconfig Provider (Multi-Cluster)

The `kubeconfig` provider supports full multi-cluster operations by reading contexts from a kubeconfig file. This is the recommended provider for multi-cluster scenarios.

**Configuration:**
```toml
[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"
# context is empty - allows switching between contexts per tool call
```

**Features:**
- Access to all contexts defined in kubeconfig
- Dynamic context switching per tool call
- Supports multiple clusters, users, and namespaces
- Works with merged kubeconfig files

**Example kubeconfig structure:**
```yaml
apiVersion: v1
kind: Config
contexts:
  - name: dev-cluster
    context:
      cluster: dev-cluster
      user: dev-user
      namespace: default
  - name: prod-cluster
    context:
      cluster: prod-cluster
      user: prod-user
      namespace: production
  - name: staging-cluster
    context:
      cluster: staging-cluster
      user: staging-user
      namespace: staging
current-context: dev-cluster
```

### 2. In-Cluster Provider (Single Cluster)

The `in-cluster` provider connects to the cluster where kube-mcp is running as a pod. It uses the pod's service account credentials.

**Configuration:**
```toml
[kubernetes]
provider = "in-cluster"
# Uses service account automatically
```

**Features:**
- Single cluster only (the cluster where kube-mcp runs)
- Uses service account credentials
- No kubeconfig required
- Context parameter is ignored (always uses "in-cluster")

**Use Cases:**
- Running kube-mcp as a pod in Kubernetes
- Service mesh or cluster-internal operations
- When you don't need multi-cluster support

### 3. Single-Cluster Provider (Fixed Context)

The `single` provider locks kube-mcp to a specific context from a kubeconfig file. This provides security by preventing accidental cluster switches.

**Configuration:**
```toml
[kubernetes]
provider = "single"
kubeconfig_path = "~/.kube/config"
context = "production-cluster"  # Fixed context
```

**Features:**
- Single cluster only (the specified context)
- Prevents accidental cluster switches
- Context parameter in tool calls is ignored
- Useful for production deployments with strict access control

**Use Cases:**
- Production deployments where cluster switching should be prevented
- Security-sensitive environments
- When you want to ensure only one cluster is accessible

## Configuring Multi-Cluster Access

### Setting Up Kubeconfig

To use multi-cluster features, you need a kubeconfig file with multiple contexts:

```bash
# View available contexts
kubectl config get-contexts

# Switch context (for kubectl)
kubectl config use-context dev-cluster

# Merge multiple kubeconfig files
KUBECONFIG=~/.kube/config:~/.kube/config-dev:~/.kube/config-prod \
  kubectl config view --flatten > ~/.kube/config-merged
```

### kube-mcp Configuration

Configure kube-mcp to use the kubeconfig provider:

```toml
[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"  # Path to your merged kubeconfig
# Leave context empty to allow per-call context switching
qps = 100
burst = 200
timeout = "30s"
```

**Path Expansion:**
- `~/.kube/config` - Expands to user's home directory
- `/etc/kube-mcp/kubeconfig` - Absolute path
- `./config` - Relative to working directory

### Docker Configuration

When running kube-mcp in Docker, mount your kubeconfig:

```bash
docker run -v ~/.kube/config:/etc/kube-mcp/kubeconfig:ro \
           -v /path/to/config.toml:/etc/kube-mcp/config.toml:ro \
           -p 8080:8080 \
           ghcr.io/wrkode/kube-mcp:1.0.0
```

Update `config.toml`:
```toml
[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "/etc/kube-mcp/kubeconfig"  # Path inside container
```

### Kubernetes Deployment

When deploying kube-mcp in Kubernetes, use a ConfigMap or Secret for kubeconfig:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-mcp-kubeconfig
data:
  config: |
    # Your kubeconfig content here
```

Mount in deployment:
```yaml
volumes:
  - name: kubeconfig
    configMap:
      name: kube-mcp-kubeconfig
volumeMounts:
  - name: kubeconfig
    mountPath: /etc/kube-mcp/kubeconfig
    subPath: config
```

## Using Multi-Cluster Features

### Listing Available Contexts

Use the `config_contexts_list` tool to see all available clusters:

```json
{
  "tool": "config_contexts_list",
  "params": {}
}
```

**Response:**
```json
{
  "contexts": [
    {
      "name": "dev-cluster",
      "cluster": "dev-cluster",
      "user": "dev-user",
      "namespace": "default"
    },
    {
      "name": "prod-cluster",
      "cluster": "prod-cluster",
      "user": "prod-user",
      "namespace": "production"
    }
  ],
  "current_context": "dev-cluster"
}
```

### Specifying Context in Tool Calls

All cluster-aware tools accept an optional `context` parameter:

**Example: List pods in dev cluster**
```json
{
  "tool": "pods_list",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}
```

**Example: List pods in prod cluster**
```json
{
  "tool": "pods_list",
  "params": {
    "context": "prod-cluster",
    "namespace": "production"
  }
}
```

**Example: Use default context (omit context parameter)**
```json
{
  "tool": "pods_list",
  "params": {
    "namespace": "default"
  }
}
```

### Default Context Behavior

When `context` is omitted or empty:
- **Kubeconfig provider**: Uses the `current-context` from kubeconfig
- **In-cluster provider**: Always uses "in-cluster" (context parameter ignored)
- **Single-cluster provider**: Uses the configured context (context parameter ignored)

### Error Handling

**Unknown Context:**
```json
{
  "error": {
    "type": "UnknownContext",
    "message": "Context 'invalid-cluster' not found in kubeconfig",
    "details": "Available contexts: dev-cluster, prod-cluster",
    "cluster": "invalid-cluster",
    "tool": "pods_list"
  }
}
```

**Context Not Available (Single-Cluster Mode):**
```json
{
  "error": {
    "type": "UnknownContext",
    "message": "Context 'dev-cluster' not available in single-cluster mode (configured: production-cluster)",
    "details": "Single-cluster provider only allows the configured context",
    "cluster": "dev-cluster",
    "tool": "pods_list"
  }
}
```

## Multi-Cluster Use Cases

### 1. Development and Production Management

Manage both dev and production clusters from a single kube-mcp instance:

```json
// Deploy to dev cluster
{
  "tool": "resources_apply",
  "params": {
    "context": "dev-cluster",
    "namespace": "default",
    "manifest": { /* deployment manifest */ }
  }
}

// Verify in production cluster
{
  "tool": "pods_list",
  "params": {
    "context": "prod-cluster",
    "namespace": "production"
  }
}
```

### 2. Cross-Cluster Resource Comparison

Compare resources across clusters:

```json
// Get deployment from dev
{
  "tool": "resources_get",
  "params": {
    "context": "dev-cluster",
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "my-app",
    "namespace": "default"
  }
}

// Get same deployment from prod
{
  "tool": "resources_get",
  "params": {
    "context": "prod-cluster",
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "my-app",
    "namespace": "production"
  }
}
```

### 3. Multi-Cluster Monitoring

Monitor multiple clusters simultaneously:

```json
// Check node health across clusters
{
  "tool": "nodes_summary",
  "params": {
    "context": "dev-cluster"
  }
}

{
  "tool": "nodes_summary",
  "params": {
    "context": "prod-cluster"
  }
}
```

### 4. Disaster Recovery

Use multi-cluster support for disaster recovery scenarios:

```json
// Backup resources from primary cluster
{
  "tool": "resources_list",
  "params": {
    "context": "primary-cluster",
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "production"
  }
}

// Restore to backup cluster
{
  "tool": "resources_apply",
  "params": {
    "context": "backup-cluster",
    "namespace": "production",
    "manifest": { /* restored manifest */ }
  }
}
```

## Performance Considerations

### Rate Limiting

Configure rate limits per cluster to avoid overwhelming API servers:

```toml
[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"
qps = 100        # Queries per second per cluster
burst = 200      # Burst limit per cluster
timeout = "30s"  # Request timeout
```

### Connection Pooling

kube-mcp maintains separate connection pools for each context:
- Connections are reused within the same context
- Each context has its own rate limiter
- Connections are established on-demand

### RBAC Caching

RBAC checks are cached per context to improve performance:

```toml
[security]
require_rbac = true
rbac_cache_ttl = 5  # Cache TTL in seconds
```

## Security Best Practices

### 1. Use Single-Cluster Provider for Production

For production deployments, consider using the single-cluster provider to prevent accidental cluster switches:

```toml
[kubernetes]
provider = "single"
kubeconfig_path = "/etc/kube-mcp/kubeconfig"
context = "production-cluster"  # Locked to this context
```

### 2. Restrict Kubeconfig Access

Limit kubeconfig file permissions:
```bash
chmod 600 ~/.kube/config
```

### 3. Use Service Accounts in Kubernetes

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

### 4. Enable RBAC Checks

Always enable RBAC checks in production:

```toml
[security]
require_rbac = true
rbac_cache_ttl = 5
```

### 5. Use Read-Only Mode for Audit

For audit and inspection purposes, use read-only mode:

```toml
[security]
read_only = true
require_rbac = true
```

## Troubleshooting

### Context Not Found

**Problem:** Tool call fails with "UnknownContext" error.

**Solutions:**
1. Verify context exists: `kubectl config get-contexts`
2. Check kubeconfig path in config.toml
3. Ensure kubeconfig is readable by kube-mcp process
4. Verify context name spelling (case-sensitive)

### Permission Denied

**Problem:** Operations fail with permission errors.

**Solutions:**
1. Check RBAC permissions for the user/context
2. Verify service account permissions (in-cluster mode)
3. Review security configuration in config.toml
4. Check kubeconfig credentials are valid

### Connection Timeouts

**Problem:** Requests timeout when accessing clusters.

**Solutions:**
1. Increase timeout in config.toml: `timeout = "60s"`
2. Verify network connectivity to cluster API servers
3. Check firewall rules
4. Verify cluster API server is accessible

### Performance Issues

**Problem:** Multi-cluster operations are slow.

**Solutions:**
1. Increase rate limits: `qps = 200`, `burst = 400`
2. Enable RBAC caching: `rbac_cache_ttl = 10`
3. Use connection pooling (automatic)
4. Consider using single-cluster provider for dedicated deployments

## Advanced Topics

### Dynamic Context Discovery

kube-mcp automatically discovers contexts from kubeconfig. When contexts are added or removed, they become available without restarting kube-mcp (for kubeconfig provider).

### Context Switching Performance

Context switching has minimal overhead:
- Connection pools are maintained per context
- RBAC caches are per-context
- No global state that needs clearing

### Merged Kubeconfig Files

kube-mcp supports merged kubeconfig files created by:
- `KUBECONFIG` environment variable
- Manual merging with `kubectl config view --flatten`
- Tools like `kubectx` and `kubens`

All contexts in merged files are available for use.

## Examples

### Complete Multi-Cluster Configuration

```toml
[server]
transports = ["stdio", "http"]
log_level = "info"

[server.http]
address = "0.0.0.0:8080"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"
# Empty context allows per-call switching
qps = 100
burst = 200
timeout = "30s"

[security]
require_rbac = true
rbac_cache_ttl = 5
read_only = false
```

### Multi-Cluster Tool Call Pattern

```json
// Pattern: Always specify context for clarity
{
  "tool": "pods_list",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}

// Pattern: Use default context when appropriate
{
  "tool": "pods_list",
  "params": {
    "namespace": "default"
  }
}
```

## Summary

kube-mcp's multi-cluster support provides:
- **Flexibility**: Choose the provider type that fits your needs
- **Security**: Single-cluster provider prevents accidental switches
- **Performance**: Efficient connection pooling and caching
- **Simplicity**: Easy context switching via tool parameters

For most use cases, the `kubeconfig` provider with an empty context provides the best balance of flexibility and control.

