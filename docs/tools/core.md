# Core Toolset

## Overview

The Core toolset provides fundamental Kubernetes operations for managing pods, generic resources, namespaces, nodes, and events. It includes tools for resource metrics, scaling, and server-side apply operations.

## Dependencies

- **Metrics Server**: `pods_top` and `nodes_top` require the `metrics.k8s.io` API (metrics-server). If unavailable, these tools return `MetricsUnavailable` errors but remain registered.

## Cluster Targeting

All tools in this toolset accept an optional `context` parameter for multi-cluster targeting:

- If `context` is omitted or empty, uses the provider's default context
- For in-cluster providers, the context parameter is ignored
- For single-cluster providers, only the configured context is allowed

## Tools

### pods_list

**Description**: List pods in a namespace or all namespaces.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes (uses `context` parameter)  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | all | Namespace to limit results to (empty for all namespaces) |

#### Output Schema

```json
{
  "pods": [
    {
      "name": "nginx-abc123",
      "namespace": "default",
      "status": "Running",
      "node": "node-1",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

#### Example Call

```json
{
  "tool": "pods_list",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}
```

#### Example Success Response

```json
{
  "pods": [
    {
      "name": "nginx-abc123",
      "namespace": "default",
      "status": "Running",
      "node": "node-1",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

#### Example Error Response

```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to list pods: namespaces \"invalid\" not found",
    "details": "namespaces \"invalid\" not found",
    "cluster": "dev-cluster",
    "tool": "pods_list"
  }
}
```

---

### pods_get

**Description**: Get detailed information about a specific pod.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | Yes | - | Pod name |
| `namespace` | string | Yes | - | Namespace |

#### Output Schema

```json
{
  "name": "nginx-abc123",
  "namespace": "default",
  "status": "Running",
  "node": "node-1",
  "created_at": "2024-01-15T10:30:00Z",
  "labels": {
    "app": "nginx",
    "version": "1.0"
  },
  "containers": [
    {
      "name": "nginx",
      "image": "nginx:1.21"
    }
  ]
}
```

#### Example Call

```json
{
  "tool": "pods_get",
  "params": {
    "context": "dev-cluster",
    "name": "nginx-abc123",
    "namespace": "default"
  }
}
```

---

### pods_delete

**Description**: Delete a pod.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | Yes | - | Pod name |
| `namespace` | string | Yes | - | Namespace |

#### Output Schema

Text result: `"Pod default/nginx-abc123 deleted successfully"`

#### Example Call

```json
{
  "tool": "pods_delete",
  "params": {
    "context": "dev-cluster",
    "name": "nginx-abc123",
    "namespace": "default"
  }
}
```

---

### pods_logs

**Description**: Fetch logs from a pod container.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | Yes | - | Pod name |
| `namespace` | string | Yes | - | Namespace |
| `container` | string | No | first container | Container name (if pod has multiple containers) |
| `tail_lines` | integer | No | all | Number of lines to retrieve from the end |

#### Output Schema

Text result containing pod logs.

#### Example Call

```json
{
  "tool": "pods_logs",
  "params": {
    "context": "dev-cluster",
    "name": "nginx-abc123",
    "namespace": "default",
    "container": "nginx",
    "tail_lines": 100
  }
}
```

---

### pods_exec

**Description**: Execute a command in a pod container.

**Read-only**: No  
**Destructive**: Yes (executes commands)  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | Yes | - | Pod name |
| `namespace` | string | Yes | - | Namespace |
| `container` | string | No | first container | Container name |
| `command` | array[string] | Yes | - | Command and arguments to execute |

#### Output Schema

Text result containing command output (stdout and stderr combined).

#### Example Call

```json
{
  "tool": "pods_exec",
  "params": {
    "context": "dev-cluster",
    "name": "nginx-abc123",
    "namespace": "default",
    "container": "nginx",
    "command": ["ls", "-la", "/usr/share/nginx/html"]
  }
}
```

---

### pods_top

**Description**: Get pod resource usage metrics (CPU and memory) from the metrics.k8s.io API.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: MetricsServer

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | all | Namespace to limit results to (empty for all namespaces) |

#### Output Schema

**Success**:
```json
{
  "pods": [
    {
      "name": "nginx-abc123",
      "namespace": "default",
      "containers": [
        {
          "name": "nginx",
          "cpu_m": 50,
          "memory": 10485760
        }
      ],
      "cpu_m": 50,
      "memory": 10485760
    }
  ],
  "metrics_available": true
}
```

**Metrics Unavailable**:
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

**Metrics Error**:
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

#### Notes

- `cpu_m`: CPU usage in millicores (1000m = 1 CPU core)
- `memory`: Memory usage in bytes
- Container-level metrics are included, plus pod-level totals
- If metrics server is unavailable, returns structured error but tool remains registered

#### Example Call

```json
{
  "tool": "pods_top",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}
```

---

### resources_list

**Description**: List resources by GroupVersionKind (GVK). Works with any Kubernetes resource including CRDs.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `group` | string | Yes | - | API group (empty string for core resources) |
| `version` | string | Yes | - | API version |
| `kind` | string | Yes | - | Resource kind |
| `namespace` | string | No | all | Namespace (empty for cluster-scoped or all namespaces) |

#### Output Schema

```json
{
  "resources": [
    {
      "name": "my-deployment",
      "namespace": "default",
      "kind": "Deployment",
      "apiVersion": "apps/v1"
    }
  ]
}
```

#### Example Call

```json
{
  "tool": "resources_list",
  "params": {
    "context": "dev-cluster",
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "namespace": "default"
  }
}
```

---

### resources_get

**Description**: Get a specific resource by GroupVersionKind, name, and namespace.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `group` | string | Yes | - | API group |
| `version` | string | Yes | - | API version |
| `kind` | string | Yes | - | Resource kind |
| `name` | string | Yes | - | Resource name |
| `namespace` | string | Yes | - | Namespace (empty for cluster-scoped resources) |

#### Output Schema

Full Kubernetes resource object (as unstructured JSON).

#### Example Call

```json
{
  "tool": "resources_get",
  "params": {
    "context": "dev-cluster",
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "my-deployment",
    "namespace": "default"
  }
}
```

---

### resources_apply

**Description**: Create or update a resource using Kubernetes server-side apply.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `manifest` | object | Yes | - | Complete Kubernetes resource manifest |
| `field_manager` | string | No | "kube-mcp" | Field manager name for server-side apply |

#### Output Schema

```json
{
  "name": "my-deployment",
  "namespace": "default",
  "kind": "Deployment",
  "status": "applied"
}
```

#### Example Call

```json
{
  "tool": "resources_apply",
  "params": {
    "context": "dev-cluster",
    "manifest": {
      "apiVersion": "apps/v1",
      "kind": "Deployment",
      "metadata": {
        "name": "my-deployment",
        "namespace": "default"
      },
      "spec": {
        "replicas": 3,
        "selector": {
          "matchLabels": {
            "app": "nginx"
          }
        },
        "template": {
          "metadata": {
            "labels": {
              "app": "nginx"
            }
          },
          "spec": {
            "containers": [
              {
                "name": "nginx",
                "image": "nginx:1.21"
              }
            ]
          }
        }
      }
    },
    "field_manager": "my-manager"
  }
}
```

---

### resources_delete

**Description**: Delete a resource by GroupVersionKind, name, and namespace.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `group` | string | Yes | - | API group |
| `version` | string | Yes | - | API version |
| `kind` | string | Yes | - | Resource kind |
| `name` | string | Yes | - | Resource name |
| `namespace` | string | Yes | - | Namespace (empty for cluster-scoped resources) |

#### Output Schema

Text result: `"Resource default/my-deployment deleted successfully"`

#### Example Call

```json
{
  "tool": "resources_delete",
  "params": {
    "context": "dev-cluster",
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "my-deployment",
    "namespace": "default"
  }
}
```

---

### resources_scale

**Description**: Scale a resource by updating its replica count via the `/scale` subresource.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `group` | string | Yes | - | API group |
| `version` | string | Yes | - | API version |
| `kind` | string | Yes | - | Resource kind |
| `name` | string | Yes | - | Resource name |
| `namespace` | string | Yes | - | Namespace |
| `replicas` | integer | Yes | - | Desired number of replicas |

#### Output Schema

**Success**:
```json
{
  "name": "my-deployment",
  "namespace": "default",
  "kind": "Deployment",
  "current_replicas": 3,
  "desired_replicas": 5,
  "generation": 2,
  "resource_version": "12345",
  "status": "scaled"
}
```

**Scaling Not Supported**:
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

#### Notes

- Only resources that expose a `/scale` subresource can be scaled (e.g., Deployments, StatefulSets, ReplicaSets)
- Returns current replica count before scaling and desired replica count after scaling
- Includes generation and resource version metadata

#### Example Call

```json
{
  "tool": "resources_scale",
  "params": {
    "context": "dev-cluster",
    "group": "apps",
    "version": "v1",
    "kind": "Deployment",
    "name": "my-deployment",
    "namespace": "default",
    "replicas": 5
  }
}
```

---

### namespaces_list

**Description**: List all namespaces in the cluster.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |

#### Output Schema

```json
{
  "namespaces": [
    {
      "name": "default",
      "status": "Active",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

#### Example Call

```json
{
  "tool": "namespaces_list",
  "params": {
    "context": "dev-cluster"
  }
}
```

---

### nodes_top

**Description**: Get node resource usage metrics (CPU and memory) from the metrics.k8s.io API.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: MetricsServer

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |

#### Output Schema

**Success**:
```json
{
  "nodes": [
    {
      "name": "node-1",
      "cpu_m": 5000,
      "memory": 10737418240,
      "cpu_percent": 50.0,
      "memory_percent": 60.0,
      "cpu_allocatable_m": 10000,
      "memory_allocatable": 17179869184
    }
  ],
  "metrics_available": true
}
```

**Metrics Unavailable**:
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

#### Notes

- `cpu_m`: CPU usage in millicores
- `memory`: Memory usage in bytes
- `cpu_percent`: CPU usage as percentage of allocatable (if available)
- `memory_percent`: Memory usage as percentage of allocatable (if available)
- `cpu_allocatable_m`: Allocatable CPU in millicores
- `memory_allocatable`: Allocatable memory in bytes

#### Example Call

```json
{
  "tool": "nodes_top",
  "params": {
    "context": "dev-cluster"
  }
}
```

---

### nodes_summary

**Description**: Get node summary statistics including capacity and allocatable resources.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | No | all | Node name to filter (empty for all nodes) |

#### Output Schema

```json
{
  "nodes": [
    {
      "name": "node-1",
      "status": "Ready",
      "cpu_capacity": 8,
      "memory_capacity": 34359738368,
      "cpu_allocatable": 8,
      "memory_allocatable": 17179869184
    }
  ]
}
```

#### Notes

- `cpu_capacity`: Total CPU cores
- `memory_capacity`: Total memory in bytes
- `cpu_allocatable`: Allocatable CPU cores
- `memory_allocatable`: Allocatable memory in bytes
- `status`: "Ready" or "NotReady"

#### Example Call

```json
{
  "tool": "nodes_summary",
  "params": {
    "context": "dev-cluster",
    "name": "node-1"
  }
}
```

---

### events_list

**Description**: List events in a namespace.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | all | Namespace (empty for all namespaces) |

#### Output Schema

```json
{
  "events": [
    {
      "name": "nginx-abc123.1234567890",
      "namespace": "default",
      "type": "Normal",
      "reason": "Started",
      "message": "Started container nginx",
      "involved_kind": "Pod",
      "involved_name": "nginx-abc123",
      "first_seen": "2024-01-15T10:30:00Z",
      "last_seen": "2024-01-15T10:30:05Z"
    }
  ]
}
```

#### Example Call

```json
{
  "tool": "events_list",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}
```

