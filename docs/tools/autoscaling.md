# Autoscaling Toolset

## Overview

The Autoscaling toolset provides tools for managing HorizontalPodAutoscalers (HPA) and KEDA ScaledObjects. HPA tools are always available (native Kubernetes), while KEDA tools require KEDA CRDs.

## Dependencies

- **HPA**: Always available (native Kubernetes API)
- **KEDA**: Requires `keda.sh/v1alpha1/ScaledObject` CRD

## Tools

### autoscaling.hpa_list

**Description**: List HorizontalPodAutoscalers.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No (always available)

#### Input Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `context` | string | No | Kubeconfig context name |
| `namespace` | string | No | Namespace name (empty for all) |
| `label_selector` | string | No | Label selector |
| `limit` | integer | No | Maximum number of items |
| `continue` | string | No | Pagination token |

#### Output Schema

```json
{
  "items": [
    {
      "name": "my-hpa",
      "namespace": "default",
      "current_replicas": 3,
      "desired_replicas": 5,
      "min_replicas": 1,
      "max_replicas": 10,
      "metrics": [
        {
          "type": "Resource",
          "name": "cpu",
          "target_avg": "70%",
          "current_avg": "85%"
        }
      ],
      "conditions": [],
      "last_scale_time": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### autoscaling.hpa_explain

**Description**: Explain HPA status and metrics.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

### autoscaling.keda_scaledobjects_list

**Description**: List KEDA ScaledObjects.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: KEDA (CRDs required)

### autoscaling.keda_scaledobject_get

**Description**: Get KEDA ScaledObject details.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: KEDA (CRDs required)

### autoscaling.keda_triggers_explain

**Description**: Explain KEDA triggers.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: KEDA (CRDs required)

### autoscaling.keda_pause

**Description**: Pause KEDA autoscaling.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: KEDA (CRDs required)  
**Requires**: `confirm: true`, RBAC check

**Implementation**: Uses `autoscaling.keda.sh/paused` annotation.

### autoscaling.keda_resume

**Description**: Resume KEDA autoscaling.

**Read-only**: No  
**Destructive**: Yes  
**Cluster-aware**: Yes  
**Feature-gated**: KEDA (CRDs required)  
**Requires**: `confirm: true`, RBAC check

## Error Codes

- **FeatureNotInstalled**: KEDA CRDs not detected (for KEDA tools)
- **KubernetesError**: Kubernetes API errors
