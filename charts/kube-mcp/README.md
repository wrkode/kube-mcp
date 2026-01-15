# kube-mcp Helm Chart

This Helm chart deploys [kube-mcp](https://github.com/wrkode/kube-mcp), a Model Context Protocol (MCP) server for Kubernetes management, on a Kubernetes cluster.

## Introduction

kube-mcp is a production-grade MCP server that provides comprehensive Kubernetes management capabilities through the Model Context Protocol. It supports multiple transports (STDIO, HTTP), multiple Kubernetes clusters, and includes toolsets for core Kubernetes operations, Helm, KubeVirt, and optional Kiali integration.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.0+
- (Optional) Prometheus Operator for ServiceMonitor/PodMonitor support
- (Optional) Metrics Server for pod/node metrics

## Installation

### Quick Start

```bash
# Option 1: Install from Helm repository (requires GitHub Pages setup)
helm repo add kube-mcp https://wrkode.github.io/kube-mcp
helm repo update
helm install kube-mcp kube-mcp/kube-mcp

# Option 2: Install from local chart
helm install kube-mcp ./charts/kube-mcp

# Option 3: Install from downloaded chart package
# Download kube-mcp-1.0.0.tgz from GitHub releases, then:
helm install kube-mcp kube-mcp-1.0.0.tgz
```

### Custom Configuration

```bash
# Install with custom values
helm install kube-mcp ./charts/kube-mcp \
  --set server.transports[0]=http \
  --set kubernetes.provider=in-cluster \
  --set security.readOnly=false
```

### Using a Values File

```bash
# Create a custom values file
cat > my-values.yaml <<EOF
server:
  transports:
    - http
  http:
    address: "0.0.0.0:8080"
kubernetes:
  provider: in-cluster
security:
  readOnly: false
EOF

# Install with custom values file
helm install kube-mcp ./charts/kube-mcp -f my-values.yaml
```

## Configuration

The following table lists the configurable parameters and their default values.

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.registry` | Container image registry | `ghcr.io` |
| `image.repository` | Container image repository | `wrkode/kube-mcp` |
| `image.tag` | Container image tag | `1.0.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Server Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `server.transports` | Enabled transports (stdio, http) | `["http"]` |
| `server.logLevel` | Log level (debug, info, warn, error) | `info` |
| `server.http.address` | HTTP server address | `0.0.0.0:8080` |
| `server.http.oauth.enabled` | Enable OAuth2/OIDC | `false` |
| `server.http.cors.enabled` | Enable CORS | `true` |

### Kubernetes Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `kubernetes.provider` | Provider (kubeconfig, in-cluster, single) | `in-cluster` |
| `kubernetes.qps` | Kubernetes API QPS | `100` |
| `kubernetes.burst` | Kubernetes API burst | `200` |
| `kubernetes.timeout` | Kubernetes API timeout | `30s` |

### Security Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `security.readOnly` | Enable read-only mode | `false` |
| `security.nonDestructive` | Enable non-destructive mode | `false` |
| `security.requireRBAC` | Require RBAC checks | `true` |
| `security.validateToken` | Validate bearer tokens | `true` |
| `security.rbacCacheTTL` | RBAC cache TTL (seconds) | `5` |

### Deployment Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |

### Service Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `8080` |

### RBAC Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account | `true` |

### Additional Features

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `autoscaling.enabled` | Enable HPA | `false` |
| `podDisruptionBudget.enabled` | Enable PDB | `false` |
| `serviceMonitor.enabled` | Enable ServiceMonitor | `false` |

## Examples

### Basic HTTP Deployment

```yaml
server:
  transports:
    - http
  http:
    address: "0.0.0.0:8080"
kubernetes:
  provider: in-cluster
```

### Read-Only Mode

```yaml
security:
  readOnly: true
  requireRBAC: true
```

### With OAuth2/OIDC

```yaml
server:
  http:
    oauth:
      enabled: true
      provider: oidc
      issuerURL: "https://your-issuer.com"
      clientID: "your-client-id"
      clientSecret: "your-client-secret"
```

### With Ingress

```yaml
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: kube-mcp.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: kube-mcp-tls
      hosts:
        - kube-mcp.example.com
```

### With Horizontal Pod Autoscaler

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80
```

### With Prometheus Monitoring

```yaml
serviceMonitor:
  enabled: true
  interval: 30s
  scrapeTimeout: 10s
```

### Multi-Transport (HTTP)

```yaml
server:
  transports:
    - http
  http:
    address: "0.0.0.0:8080"
```

## Upgrading

```bash
# Upgrade with new values
helm upgrade kube-mcp ./charts/kube-mcp -f my-values.yaml

# Upgrade to specific version
helm upgrade kube-mcp ./charts/kube-mcp --version 1.0.0
```

## Uninstallation

```bash
helm uninstall kube-mcp
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app.kubernetes.io/name=kube-mcp
```

### View Logs

```bash
kubectl logs -l app.kubernetes.io/name=kube-mcp
```

### Check Configuration

```bash
kubectl get configmap kube-mcp-config -o yaml
```

### Test Health Endpoint

```bash
kubectl port-forward svc/kube-mcp 8080:8080
curl http://localhost:8080/health
```

## Security Considerations

1. **RBAC**: The chart creates a ClusterRole with broad permissions. Review and customize `rbac.rules` based on your needs.

2. **Service Account**: Uses a dedicated service account for better security isolation.

3. **Read-Only Mode**: Enable `security.readOnly: true` for read-only deployments.

4. **OAuth**: Configure OAuth2/OIDC for production deployments requiring authentication.

5. **Network Policies**: Consider adding NetworkPolicy resources to restrict network access.

## Support

- **Issues**: [GitHub Issues](https://github.com/wrkode/kube-mcp/issues)
- **Documentation**: [docs/](https://github.com/wrkode/kube-mcp/tree/main/docs)
- **Discussions**: [GitHub Discussions](https://github.com/wrkode/kube-mcp/discussions)

## License

Apache License 2.0 - See [LICENSE](https://github.com/wrkode/kube-mcp/blob/main/LICENSE) for details.

