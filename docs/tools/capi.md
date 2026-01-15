# CAPI Toolset

## Overview

The CAPI (Cluster API) toolset provides tools for managing Cluster API resources. This toolset is **optional** and only available when the required CRDs are detected in the cluster.

## Dependencies

- **CAPI CRDs**:
  - `cluster.x-k8s.io/v1beta1/Cluster` (required - enables toolset)
  - `cluster.x-k8s.io/v1beta1/Machine` (optional - enables machine tools)
  - `cluster.x-k8s.io/v1beta1/MachineDeployment` (optional - enables machine deployment tools)

## Feature Gating

- **Tool Registration**: If Cluster CRD is not detected, **no tools from this toolset are registered**
- **CRD Detection**: The toolset automatically detects available CRDs on startup
- **Partial Functionality**: If only Cluster CRD exists, cluster tools are available but machine tools are not
- **Error Handling**: Tools return structured errors when CRDs are missing or operations fail

## Cluster Targeting

All tools in this toolset accept an optional `context` parameter for multi-cluster targeting:

- If `context` is omitted or empty, uses the provider's default context
- Each cluster is checked independently for CAPI CRDs

## Tools

### capi.clusters_list

**Description**: List Cluster API clusters.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: CAPI (Cluster CRD required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | "" | Namespace name (empty for all namespaces) |
| `label_selector` | string | No | "" | Label selector |

#### Output Schema

**Success**:
```json
{
  "items": [
    {
      "name": "my-cluster",
      "namespace": "default",
      "ready": true,
      "control_plane_ready": true,
      "infrastructure_ready": true,
      "kubernetes_version": "v1.28.0",
      "control_plane_ref": {
        "api_version": "controlplane.cluster.x-k8s.io/v1beta1",
        "kind": "KubeadmControlPlane",
        "name": "my-cluster-control-plane",
        "namespace": "default"
      },
      "infrastructure_ref": {
        "api_version": "infrastructure.cluster.x-k8s.io/v1beta1",
        "kind": "AWSCluster",
        "name": "my-cluster",
        "namespace": "default"
      },
      "conditions": [
        {
          "type": "Ready",
          "status": "True",
          "reason": "AllReplicasReady",
          "message": "Cluster is ready",
          "time": "2024-01-15T10:30:00Z"
        }
      ]
    }
  ]
}
```

### capi.cluster_get

**Description**: Get details for a specific Cluster API cluster.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: CAPI (Cluster CRD required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | Yes | - | Namespace name |
| `name` | string | Yes | - | Cluster name |
| `raw` | boolean | No | false | Return raw object if true |

#### Output Schema

Same as `capi.clusters_list` but for a single cluster.

### capi.machines_list

**Description**: List machines for a cluster. **Note**: Requires Machine CRD.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: CAPI (Machine CRD required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `cluster_namespace` | string | Yes | - | Cluster namespace |
| `cluster_name` | string | Yes | - | Cluster name |

#### Output Schema

**Success**:
```json
{
  "items": [
    {
      "name": "my-cluster-control-plane-0",
      "namespace": "default",
      "phase": "Running",
      "ready": true,
      "node_ref": "my-cluster-control-plane-0",
      "version": "v1.28.0",
      "conditions": [
        {
          "type": "Ready",
          "status": "True",
          "reason": "NodeReady",
          "message": "Machine is ready",
          "time": "2024-01-15T10:30:00Z"
        }
      ]
    }
  ],
  "continue": "eyJ2IjoibWV0YS5rOHMuaW8vdjEiLCJydiI6MjE0NzQ4MzY0N30"
}
```

**Note**: The `continue` field is only present when pagination is used and more results are available.

### capi.machinedeployments_list

**Description**: List machine deployments for a cluster. **Note**: Requires MachineDeployment CRD.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: CAPI (MachineDeployment CRD required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `cluster_namespace` | string | Yes | - | Cluster namespace |
| `cluster_name` | string | Yes | - | Cluster name |

#### Output Schema

**Success**:
```json
{
  "items": [
    {
      "name": "my-cluster-md-0",
      "namespace": "default",
      "replicas_desired": 3,
      "replicas_ready": 3,
      "replicas_updated": 3,
      "paused": false,
      "conditions": [
        {
          "type": "Ready",
          "status": "True",
          "reason": "AllReplicasReady",
          "message": "MachineDeployment is ready",
          "time": "2024-01-15T10:30:00Z"
        }
      ]
    }
  ]
}
```

### capi.rollout_status

**Description**: Get high-level rollout status for a cluster.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: CAPI (Cluster CRD required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `cluster_namespace` | string | Yes | - | Cluster namespace |
| `cluster_name` | string | Yes | - | Cluster name |

#### Output Schema

**Success**:
```json
{
  "status": {
    "ready": true,
    "message": "Cluster is ready",
    "counts": {
      "machines_desired": 4,
      "machines_ready": 4,
      "machines_updated": 4
    },
    "blockers": []
  }
}
```

### capi.scale_machinedeployment

**Description**: Scale a machine deployment. **Note**: Requires MachineDeployment CRD and RBAC permissions.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: CAPI (MachineDeployment CRD required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | Yes | - | Namespace name |
| `name` | string | Yes | - | MachineDeployment name |
| `replicas` | integer | Yes | - | Number of replicas |
| `confirm` | boolean | Yes | - | Must be true to scale |

#### Output Schema

**Success**:
```json
{
  "summary": {
    "name": "my-cluster-md-0",
    "namespace": "default",
    "replicas_desired": 5,
    "replicas_ready": 3,
    "replicas_updated": 3,
    "paused": false,
    "conditions": [...]
  }
}
```

## Examples

### List all clusters

```json
{
  "tool": "capi.clusters_list",
  "arguments": {
    "context": "dev-cluster"
  }
}
```

### Get cluster details

```json
{
  "tool": "capi.cluster_get",
  "arguments": {
    "context": "dev-cluster",
    "namespace": "default",
    "name": "my-cluster"
  }
}
```

### List machines for a cluster

```json
{
  "tool": "capi.machines_list",
  "arguments": {
    "context": "dev-cluster",
    "cluster_namespace": "default",
    "cluster_name": "my-cluster"
  }
}
```

### Get rollout status

```json
{
  "tool": "capi.rollout_status",
  "arguments": {
    "context": "dev-cluster",
    "cluster_namespace": "default",
    "cluster_name": "my-cluster"
  }
}
```

### Scale a machine deployment

```json
{
  "tool": "capi.scale_machinedeployment",
  "arguments": {
    "context": "dev-cluster",
    "namespace": "default",
    "name": "my-cluster-md-0",
    "replicas": 5,
    "confirm": true
  }
}
```

## Error Codes

- **FeatureNotInstalled**: Returned if toolset is disabled (Cluster CRD not detected) or if machine tools are called without Machine/MachineDeployment CRDs
- **KubernetesError**: Returned for Kubernetes API errors (resource not found, permission denied, etc.)

## Notes

- Machine and MachineDeployment tools are only available if their respective CRDs are detected
- The `rollout_status` tool provides best-effort machine counts if Machine CRD is available
- Scaling operations require RBAC permissions and `confirm: true`
