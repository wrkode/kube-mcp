# Helm Toolset

## Overview

The Helm toolset provides tools for managing Helm charts and releases. It supports installing charts, listing releases, and uninstalling releases across multiple Kubernetes clusters.

## Dependencies

- **Helm**: Requires Helm 3.x to be available
- **Kubernetes Cluster**: Requires access to a Kubernetes cluster for release management

## Cluster Targeting

All tools in this toolset accept an optional `context` parameter for multi-cluster targeting:

- If `context` is omitted or empty, uses the provider's default context
- Helm releases are stored per-cluster

## Tools

### helm_install

**Description**: Install a Helm chart into a Kubernetes cluster.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes (uses `context` parameter)  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | Yes | - | Release name |
| `chart` | string | Yes | - | Chart name or path (e.g., "nginx" or "./charts/my-chart") |
| `namespace` | string | Yes | - | Namespace to install into |
| `values` | object | No | {} | Chart values (YAML values as JSON object) |
| `version` | string | No | latest | Chart version |

#### Output Schema

```json
{
  "name": "my-release",
  "namespace": "default",
  "status": "deployed",
  "version": 1
}
```

#### Example Call

```json
{
  "tool": "helm_install",
  "params": {
    "context": "dev-cluster",
    "name": "my-release",
    "chart": "nginx",
    "namespace": "default",
    "values": {
      "replicaCount": 3,
      "image": {
        "tag": "1.21"
      }
    },
    "version": "13.2.0"
  }
}
```

#### Example Success Response

```json
{
  "name": "my-release",
  "namespace": "default",
  "status": "deployed",
  "version": 1
}
```

#### Example Error Response

```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to install chart: chart \"invalid-chart\" not found",
    "details": "chart \"invalid-chart\" not found",
    "cluster": "dev-cluster",
    "tool": "helm_install"
  }
}
```

---

### helm_releases_list

**Description**: List Helm releases in a namespace or all namespaces.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | all | Namespace to limit results to (empty for all namespaces) |

#### Output Schema

```json
{
  "releases": [
    {
      "name": "my-release",
      "namespace": "default",
      "status": "deployed",
      "version": 1,
      "chart": "nginx-13.2.0"
    }
  ]
}
```

#### Example Call

```json
{
  "tool": "helm_releases_list",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}
```

#### Example Success Response

```json
{
  "releases": [
    {
      "name": "my-release",
      "namespace": "default",
      "status": "deployed",
      "version": 1,
      "chart": "nginx-13.2.0"
    }
  ]
}
```

---

### helm_uninstall

**Description**: Uninstall a Helm release from a Kubernetes cluster.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `name` | string | Yes | - | Release name |
| `namespace` | string | Yes | - | Namespace |

#### Output Schema

```json
{
  "name": "my-release",
  "status": "uninstalled",
  "message": "Release \"my-release\" uninstalled"
}
```

#### Example Call

```json
{
  "tool": "helm_uninstall",
  "params": {
    "context": "dev-cluster",
    "name": "my-release",
    "namespace": "default"
  }
}
```

#### Example Success Response

```json
{
  "name": "my-release",
  "status": "uninstalled",
  "message": "Release \"my-release\" uninstalled"
}
```

#### Example Error Response

```json
{
  "error": {
    "type": "KubernetesError",
    "message": "Failed to uninstall release: release: not found",
    "details": "release: not found",
    "cluster": "dev-cluster",
    "tool": "helm_uninstall"
  }
}
```

