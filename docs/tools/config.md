# Config Toolset

## Overview

The Config toolset provides tools for inspecting kubeconfig files and managing Kubernetes contexts. These tools help clients discover available clusters and understand the current configuration.

## Dependencies

None. All tools in this toolset work without requiring any Kubernetes cluster connection.

## Cluster Targeting

The Config toolset does not use cluster targeting since it operates on kubeconfig files rather than cluster APIs.

## Tools

### config_contexts_list

**Description**: List all available Kubernetes contexts from the kubeconfig file.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: No  
**Feature-gated**: No

#### Input Schema

No input parameters required.

#### Output Schema

```json
{
  "contexts": [
    "dev-cluster",
    "prod-eu",
    "prod-us"
  ],
  "current_context": "dev-cluster"
}
```

#### Example Call

```json
{
  "tool": "config_contexts_list",
  "params": {}
}
```

#### Example Success Response

```json
{
  "contexts": [
    "dev-cluster",
    "prod-eu",
    "prod-us"
  ],
  "current_context": "dev-cluster"
}
```

#### Example Error Response

```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to list contexts: failed to load kubeconfig",
    "details": "failed to load kubeconfig: open /path/to/kubeconfig: no such file or directory"
  }
}
```

---

### config_kubeconfig_view

**Description**: View kubeconfig file contents (full or minified).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: No  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `minified` | boolean | No | false | If true, return minified version |

#### Output Schema

**Note**: Currently returns a placeholder message. Full implementation pending.

```json
{
  "message": "Kubeconfig viewing not yet implemented - requires kubeconfig path access",
  "minified": false
}
```

#### Example Call

```json
{
  "tool": "config_kubeconfig_view",
  "params": {
    "minified": false
  }
}
```

#### Example Success Response

```json
{
  "message": "Kubeconfig viewing not yet implemented - requires kubeconfig path access",
  "minified": false
}
```

