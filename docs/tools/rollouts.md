# Progressive Delivery Toolset

## Overview

The Progressive Delivery toolset provides tools for managing progressive delivery resources via CRDs. It supports Argo Rollouts (Rollout) and Flagger (Canary). This toolset is **optional** and only available when the required CRDs are detected in the cluster.

## Dependencies

- **Argo Rollouts CRD**:
  - `argoproj.io/v1alpha1/Rollout` (enables toolset)
- **Flagger CRD**:
  - `flagger.app/v1beta1/Canary` (enables toolset)

## Feature Gating

- **Tool Registration**: If none of the required CRDs are detected, **no tools from this toolset are registered**
- **CRD Detection**: The toolset automatically detects available CRDs on startup
- **Error Handling**: Tools return structured errors when CRDs are missing or operations fail

## Tools

### rollouts.list

**Description**: List progressive delivery resources (Argo Rollouts Rollout or Flagger Canary).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Rollouts (CRDs required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | "" | Namespace name (empty for all namespaces) |
| `label_selector` | string | No | "" | Label selector |
| `limit` | integer | No | 0 | Maximum number of items to return |
| `continue` | string | No | "" | Token from previous paginated request |

#### Output Schema

```json
{
  "items": [
    {
      "kind": "Rollout",
      "name": "my-app",
      "namespace": "default",
      "phase": "Progressing",
      "ready": false,
      "message": "Rolling out new revision",
      "revision": "abc123"
    }
  ],
  "continue": "token..."
}
```

### rollouts.get_status

**Description**: Get detailed status of a progressive delivery resource.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Rollouts (CRDs required)

#### Input Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `context` | string | No | Kubeconfig context name |
| `kind` | string | Yes | Resource kind: "Rollout" or "Canary" |
| `name` | string | Yes | Resource name |
| `namespace` | string | Yes | Namespace name |
| `raw` | boolean | No | Return raw object if true |

#### Output Schema

```json
{
  "kind": "Rollout",
  "name": "my-app",
  "namespace": "default",
  "phase": "Progressing",
  "ready": false,
  "revisions": ["abc123", "def456"],
  "current_step": 2,
  "total_steps": 5,
  "traffic_weight": 50,
  "paused": false,
  "conditions": []
}
```

### rollouts.promote

**Description**: Promote a rollout to the next step (Argo Rollouts only).

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: Rollouts (CRDs required)  
**Requires**: `confirm: true`, RBAC check

#### Input Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `context` | string | No | Kubeconfig context name |
| `kind` | string | Yes | Resource kind: "Rollout" |
| `name` | string | Yes | Resource name |
| `namespace` | string | Yes | Namespace name |
| `confirm` | boolean | Yes | Must be true to promote |

**Note**: Flagger Canary promotion is not yet supported (returns FeatureDisabled).

### rollouts.abort

**Description**: Abort a rollout (Argo Rollouts only).

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: Rollouts (CRDs required)  
**Requires**: `confirm: true`, RBAC check

### rollouts.retry

**Description**: Retry a rollout analysis or progression (Argo Rollouts only).

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: Rollouts (CRDs required)  
**Requires**: `confirm: true`, RBAC check

## Error Codes

- **FeatureNotInstalled**: Required CRDs not detected
- **FeatureDisabled**: Operation not supported for this engine (e.g., Flagger promote)
- **KubernetesError**: Kubernetes API errors (Forbidden, NotFound, etc.)

## Examples

See [Usage Guide](../USAGE_GUIDE.md) for examples.
