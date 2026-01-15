# GitOps Toolset

## Overview

The GitOps toolset provides tools for managing GitOps applications via CRDs. It supports Flux (Kustomization and HelmRelease) and Argo CD (Application). This toolset is **optional** and only available when the required CRDs are detected in the cluster.

## Dependencies

- **Flux CRDs** (any combination enables toolset):
  - `kustomize.toolkit.fluxcd.io/v1/Kustomization` (required for Kustomization tools)
  - `helm.toolkit.fluxcd.io/v2/HelmRelease` (required for HelmRelease tools)
- **Argo CD CRD**:
  - `argoproj.io/v1alpha1/Application` (required for Application tools)

## Feature Gating

- **Tool Registration**: If none of the required CRDs are detected, **no tools from this toolset are registered**
- **CRD Detection**: The toolset automatically detects available CRDs on startup
- **Error Handling**: Tools return structured errors when CRDs are missing or operations fail

## Cluster Targeting

All tools in this toolset accept an optional `context` parameter for multi-cluster targeting:

- If `context` is omitted or empty, uses the provider's default context
- Each cluster is checked independently for GitOps CRDs

## Tools

### gitops.apps_list

**Description**: List GitOps applications (Flux Kustomization/HelmRelease or Argo CD Application).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: GitOps (CRDs required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | "" | Namespace name (empty for all namespaces) |
| `label_selector` | string | No | "" | Label selector |
| `kinds` | array | No | all available | Array of kinds: "Kustomization", "HelmRelease", "Application" |
| `limit` | integer | No | 0 | Maximum number of items to return (0 = no limit) |
| `continue` | string | No | "" | Token from previous paginated request |

#### Output Schema

**Success**:
```json
{
  "items": [
    {
      "kind": "Kustomization",
      "name": "my-app",
      "namespace": "default",
      "ready": true,
      "status": "Ready",
      "last_updated": "2024-01-15T10:30:00Z",
      "revision": "main/abc123",
      "message": "Applied revision: main/abc123"
    }
  ],
  "continue": "eyJ2IjoibWV0YS5rOHMuaW8vdjEiLCJydiI6MjE0NzQ4MzY0N30"
}
```

**Note**: The `continue` field is only present when pagination is used and more results are available.

**Error (FeatureNotInstalled)**:
```json
{
  "error": {
    "type": "FeatureNotInstalled",
    "message": "GitOps toolset is not enabled",
    "details": "Required GitOps CRDs not detected in cluster"
  }
}
```

### gitops.app_get

**Description**: Get details for a specific GitOps application.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: GitOps (CRDs required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `kind` | string | Yes | - | Application kind: "Kustomization", "HelmRelease", or "Application" |
| `name` | string | Yes | - | Application name |
| `namespace` | string | Yes | - | Namespace name (required for namespaced kinds) |
| `raw` | boolean | No | false | Return raw object if true |

#### Output Schema

**Success**:
```json
{
  "summary": {
    "kind": "Kustomization",
    "name": "my-app",
    "namespace": "default",
    "ready": true,
    "status": "Ready",
    "last_updated": "2024-01-15T10:30:00Z",
    "revision": "main/abc123",
    "message": "Applied revision: main/abc123",
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "reason": "ReconciliationSucceeded",
        "message": "Applied revision: main/abc123",
        "time": "2024-01-15T10:30:00Z"
      }
    ]
  },
  "raw_object": { ... }
}
```

### gitops.app_reconcile

**Description**: Trigger reconciliation for a Flux Kustomization or HelmRelease. **Note**: Reconcile is only supported for Flux resources. Argo CD Application reconcile requires Argo CD API access and is not supported in CRD-only mode.

**Read-only**: No  
**Destructive**: Yes (mutates annotations)  
**Cluster-aware**: Yes  
**Feature-gated**: GitOps (CRDs required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `kind` | string | Yes | - | Application kind: "Kustomization" or "HelmRelease" |
| `name` | string | Yes | - | Application name |
| `namespace` | string | Yes | - | Namespace name |
| `confirm` | boolean | Yes | - | Must be true to reconcile |

#### Output Schema

**Success**:
```json
{
  "result": {
    "annotation_applied": "kustomize.toolkit.fluxcd.io/reconcile",
    "timestamp": "2024-01-15T10:35:00Z"
  },
  "summary": {
    "kind": "Kustomization",
    "name": "my-app",
    "namespace": "default",
    "ready": true,
    "status": "Ready",
    ...
  }
}
```

**Error (Argo CD Application)**:
```json
{
  "error": {
    "type": "FeatureDisabled",
    "message": "Reconcile is not supported for Argo CD Application in CRD-only mode",
    "details": "Argo CD Application reconcile requires Argo CD API access. Use Flux Kustomization or HelmRelease for reconcile operations."
  }
}
```

## Examples

### List all GitOps applications

```json
{
  "tool": "gitops.apps_list",
  "arguments": {
    "context": "dev-cluster"
  }
}
```

### Get a specific Kustomization

```json
{
  "tool": "gitops.app_get",
  "arguments": {
    "context": "dev-cluster",
    "kind": "Kustomization",
    "name": "my-app",
    "namespace": "default"
  }
}
```

### Reconcile a HelmRelease

```json
{
  "tool": "gitops.app_reconcile",
  "arguments": {
    "context": "dev-cluster",
    "kind": "HelmRelease",
    "name": "my-chart",
    "namespace": "default",
    "confirm": true
  }
}
```

## Error Codes

- **FeatureNotInstalled**: Returned if toolset is disabled (CRDs not detected) or if a specific CRD is missing (e.g., Kustomization CRD not available when querying Kustomization)
- **FeatureDisabled**: Returned if reconcile is attempted on Argo CD Application (not supported in CRD-only mode)
- **KubernetesError**: Returned for Kubernetes API errors (resource not found, permission denied, etc.)
