# Kiali Toolset

## Overview

The Kiali toolset provides tools for service mesh observability through the Kiali API. This toolset is **optional** and only available when Kiali is configured and reachable.

## Dependencies

- **Kiali Server**: Requires a configured and reachable Kiali instance
- **Configuration**: Kiali URL must be configured in kube-mcp configuration
- **Authentication**: Optional token-based authentication (can use OAuth)

## Feature Gating

- **Tool Registration**: If Kiali is not configured (`enabled: false` or URL not set), **no tools from this toolset are registered**
- **Configuration Validation**: Kiali configuration is validated on startup
- **Error Handling**: Tools return structured errors when Kiali is unreachable or misconfigured

## Cluster Targeting

Kiali tools operate on a single Kiali instance configured in kube-mcp. The `context` parameter is **not used** for Kiali tools since Kiali itself manages multi-cluster service mesh visibility.

## Tools

### kiali_mesh_graph

**Description**: Get the service mesh graph for a namespace or all namespaces.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: No (Kiali manages cluster visibility)  
**Feature-gated**: Kiali

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `namespace` | string | No | all | Namespace to limit results to (empty for all namespaces) |

#### Output Schema

Returns the Kiali mesh graph JSON structure (varies by Kiali version).

```json
{
  "nodes": [...],
  "edges": [...],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Example Call

```json
{
  "tool": "kiali_mesh_graph",
  "params": {
    "namespace": "default"
  }
}
```

#### Example Error Response

```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to get mesh graph: failed to reach Kiali server at https://kiali.example.com: connection refused",
    "details": "failed to reach Kiali server at https://kiali.example.com: connection refused"
  }
}
```

---

### kiali_istio_config_get

**Description**: Get Istio configuration for a namespace, optionally filtered by object type.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: No  
**Feature-gated**: Kiali

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `namespace` | string | No | all | Namespace |
| `object_type` | string | No | all | Istio object type (e.g., "virtualservices", "destinationrules") |

#### Output Schema

Returns Istio configuration objects (structure varies by object type).

#### Example Call

```json
{
  "tool": "kiali_istio_config_get",
  "params": {
    "namespace": "default",
    "object_type": "virtualservices"
  }
}
```

---

### kiali_metrics

**Description**: Get metrics for a service in a namespace.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: No  
**Feature-gated**: Kiali

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `namespace` | string | Yes | - | Namespace |
| `service` | string | No | all | Service name (empty for all services in namespace) |

#### Output Schema

Returns Kiali metrics structure (varies by Kiali version and metrics type).

#### Example Call

```json
{
  "tool": "kiali_metrics",
  "params": {
    "namespace": "default",
    "service": "nginx"
  }
}
```

---

### kiali_logs

**Description**: Get logs for a workload in a namespace.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: No  
**Feature-gated**: Kiali

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `namespace` | string | Yes | - | Namespace |
| `workload` | string | No | all | Workload name (empty for all workloads in namespace) |

#### Output Schema

Returns Kiali logs structure.

#### Example Call

```json
{
  "tool": "kiali_logs",
  "params": {
    "namespace": "default",
    "workload": "nginx-deployment"
  }
}
```

---

### kiali_traces

**Description**: Get distributed tracing data for a service in a namespace.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: No  
**Feature-gated**: Kiali

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `namespace` | string | Yes | - | Namespace |
| `service` | string | No | all | Service name (empty for all services in namespace) |

#### Output Schema

Returns Kiali traces structure.

#### Example Call

```json
{
  "tool": "kiali_traces",
  "params": {
    "namespace": "default",
    "service": "nginx"
  }
}
```

## Error Handling

### Toolset Not Available

If Kiali is not configured (`enabled: false` or URL not set), the toolset is not registered and tools will not appear in the tool list. This is not an error condition - it's expected behavior when Kiali is not configured.

### Common Errors

- **KubernetesError**: Used for Kiali API errors (connection refused, authentication failures, HTTP errors)
- **FeatureDisabled**: Returned if toolset is disabled in configuration

### Example Error Responses

**Kiali Unreachable**:
```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to get mesh graph: failed to reach Kiali server at https://kiali.example.com: connection refused",
    "details": "failed to reach Kiali server at https://kiali.example.com: connection refused"
  }
}
```

**Authentication Failure**:
```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to get mesh graph: Kiali API returned status 401: Unauthorized",
    "details": "Kiali API returned status 401: Unauthorized"
  }
}
```

**Invalid Configuration**:
```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to get mesh graph: failed to create Kiali client: Kiali URL is required",
    "details": "Kiali URL is required"
  }
}
```

## Configuration

Kiali must be configured in the kube-mcp configuration file:

```toml
[kiali]
enabled = true
url = "https://kiali.example.com"
token = "optional-auth-token"
timeout = "30s"

[kiali.tls]
enabled = true
ca_file = "/path/to/ca.crt"
cert_file = "/path/to/client.crt"
key_file = "/path/to/client.key"
insecure_skip_verify = false
```

## TLS Support

Kiali tools support TLS configuration:
- **CA Certificate**: For custom CA bundles
- **Client Certificates**: For mutual TLS authentication
- **Insecure Skip Verify**: For development/testing (not recommended for production)

