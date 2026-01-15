# Error Contract

## Standard Error JSON Shape

All tools return errors in a consistent JSON structure:

```json
{
  "error": {
    "type": "ErrorCode",
    "message": "Human-readable error message",
    "details": "Additional error details or Kubernetes error message",
    "cluster": "context-name",
    "tool": "tool_name"
  }
}
```

When tools return errors via `NewErrorResult()`, the error message is returned as text content. When tools return structured errors via `NewJSONResult()` with an error object, the full error structure is available.

## Error Codes

### KubernetesError

**When used**: General Kubernetes API errors (permission denied, resource not found, validation errors, etc.)

**HTTP/Kubernetes equivalents**: Any non-2xx status code from Kubernetes API

**Example tools**: All tools that interact with Kubernetes API

**Example JSON**:
```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to get pod: pods \"nginx\" not found",
    "details": "pods \"nginx\" not found",
    "cluster": "dev-cluster",
    "tool": "pods_get"
  }
}
```

### MetricsUnavailable

**When used**: When the metrics.k8s.io API is not available or metrics server is not installed

**HTTP/Kubernetes equivalents**: N/A (API not available)

**Example tools**: `pods_top`, `nodes_top`

**Example JSON**:
```json
{
  "error": {
    "type": "MetricsUnavailable",
    "message": "Metrics server is not available",
    "details": "The metrics.k8s.io API is not available. Ensure metrics-server is installed and running."
  },
  "metrics_available": false
}
```

### MetricsError

**When used**: When metrics API is available but fails to retrieve metrics

**HTTP/Kubernetes equivalents**: Errors from metrics.k8s.io API

**Example tools**: `pods_top`, `nodes_top`

**Example JSON**:
```json
{
  "error": {
    "type": "MetricsError",
    "message": "Failed to retrieve pod metrics: connection refused",
    "details": "connection refused"
  },
  "metrics_available": false
}
```

### ScalingNotSupported

**When used**: When attempting to scale a resource that does not expose a `/scale` subresource

**HTTP/Kubernetes equivalents**: 404 Not Found on scale subresource

**Example tools**: `resources_scale`

**Example JSON**:
```json
{
  "error": {
    "type": "ScalingNotSupported",
    "message": "Resource apps/v1/ConfigMap does not support scaling",
    "details": "the server could not find the requested resource (get configmaps.scale)",
    "gvk": {
      "group": "apps",
      "version": "v1",
      "kind": "ConfigMap"
    }
  }
}
```

### InvalidManifest

**When used**: When a resource manifest is invalid or missing required fields

**HTTP/Kubernetes equivalents**: 422 Unprocessable Entity

**Example tools**: `resources_apply`, `kubevirt_vm_create`

**Example JSON**:
```json
{
  "error": {
    "type": "InvalidManifest",
    "message": "VirtualMachine manifest must have a name",
    "details": "The manifest is missing the 'metadata.name' field"
  }
}
```

### UnknownContext

**When used**: When a specified cluster context is not found or not available

**HTTP/Kubernetes equivalents**: N/A (client-side error)

**Example tools**: All cluster-aware tools

**Example JSON**:
```json
{
  "error": {
    "type": "UnknownContext",
    "message": "Context 'dev-cluster' not found",
    "details": "failed to get client set for context dev-cluster: context not found",
    "cluster": "dev-cluster"
  }
}
```

### FeatureDisabled

**When used**: When a feature or toolset is disabled via configuration or server mode (read_only/disable_destructive)

**HTTP/Kubernetes equivalents**: N/A

**Example tools**: Tools that require write access when read_only mode is enabled

**Example JSON**:
```json
{
  "error": {
    "type": "FeatureDisabled",
    "message": "Feature is disabled",
    "details": "Read-only mode is enabled"
  }
}
```

### FeatureNotInstalled

**When used**: When a required CRD or API is not present in the cluster

**HTTP/Kubernetes equivalents**: N/A (CRD/API not available)

**Example tools**: GitOps tools (when CRDs not detected), Policy tools (when CRDs not detected), CAPI tools (when CRDs not detected)

**Example JSON**:
```json
{
  "error": {
    "type": "FeatureNotInstalled",
    "message": "GitOps toolset is not enabled",
    "details": "Required GitOps CRDs (Flux Kustomization/HelmRelease or Argo CD Application) not detected in cluster"
  }
}
```

**Note**: If a toolset is not registered due to missing CRDs, no runtime error is returned (tools simply don't exist). This error is returned when a toolset is registered but a specific capability requires an optional CRD that is missing.

### ExternalServiceUnavailable

**When used**: When a required external service endpoint is configured but unreachable

**HTTP/Kubernetes equivalents**: Connection errors, timeouts

**Example tools**: Kiali tools (when Kiali server is unreachable)

**Example JSON**:
```json
{
  "error": {
    "type": "ExternalServiceUnavailable",
    "message": "Kiali server is unreachable",
    "details": "Connection refused"
  }
}
```

### ValidationError

**When used**: When input validation fails (bad input format, missing required fields, etc.)

**HTTP/Kubernetes equivalents**: 422 Unprocessable Entity

**Example tools**: All tools with input validation

**Example JSON**:
```json
{
  "error": {
    "type": "ValidationError",
    "message": "Invalid input",
    "details": "namespace is required for namespaced resources"
  }
}
```

## Error Response Format

### Via NewErrorResult()

When tools use `NewErrorResult(err)`, the error is returned as text content:

```json
{
  "isError": true,
  "content": [
    {
      "type": "text",
      "text": "Failed to get pod: pods \"nginx\" not found"
    }
  ]
}
```

### Via NewJSONResult() with Error Object

When tools use `NewJSONResult()` with an error object, the structured error is available:

```json
{
  "isError": false,
  "content": [
    {
      "type": "text",
      "text": "{\"error\":{\"type\":\"MetricsUnavailable\",\"message\":\"Metrics server is not available\",\"details\":\"The metrics.k8s.io API is not available. Ensure metrics-server is installed and running.\"},\"metrics_available\":false}"
    }
  ]
}
```

## Error Handling Best Practices

1. **Check `isError` field**: Always check the `isError` boolean in the response
2. **Parse error JSON**: For structured errors, parse the JSON content to access error details
3. **Handle gracefully**: Tools like `pods_top` return errors but remain functional (they just indicate metrics unavailable)
4. **Check feature availability**: Use `config_contexts_list` to verify available contexts before targeting
5. **Validate inputs**: Many errors can be avoided by validating inputs before calling tools

## Common Error Scenarios

### Permission Denied

```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Forbidden: user does not have access",
    "details": "pods is forbidden: User \"system:serviceaccount:default:user\" cannot get resource \"pods\" in API group \"\" in the namespace \"default\""
  }
}
```

### Resource Not Found

```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to get pod: pods \"nginx\" not found",
    "details": "pods \"nginx\" not found"
  }
}
```

### Invalid GVK

```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to map GVK to GVR: no matches for kind \"InvalidKind\" in group \"invalid.group\"",
    "details": "no matches for kind \"InvalidKind\" in group \"invalid.group\""
  }
}
```

