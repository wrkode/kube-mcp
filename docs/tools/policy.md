# Policy Toolset

## Overview

The Policy toolset provides tools for managing and inspecting policy engines in Kubernetes. It supports Kyverno and Gatekeeper. This toolset is **optional** and only available when the required CRDs are detected in the cluster.

## Dependencies

- **Kyverno CRDs** (any enables toolset):
  - `kyverno.io/v1/ClusterPolicy` (required for cluster policies)
  - `kyverno.io/v1/Policy` (required for namespaced policies)
  - `wgpolicyk8s.io/v1alpha2/PolicyReport` (optional, enhances violations functionality)
  - `wgpolicyk8s.io/v1alpha2/ClusterPolicyReport` (optional, enhances violations functionality)
- **Gatekeeper CRDs**:
  - `templates.gatekeeper.sh/v1beta1/ConstraintTemplate` (required for Gatekeeper support)

## Feature Gating

- **Tool Registration**: If neither Kyverno nor Gatekeeper CRDs are detected, **no tools from this toolset are registered**
- **CRD Detection**: The toolset automatically detects available CRDs on startup
- **Error Handling**: Tools return structured errors when CRDs are missing or operations fail

## Cluster Targeting

All tools in this toolset accept an optional `context` parameter for multi-cluster targeting:

- If `context` is omitted or empty, uses the provider's default context
- Each cluster is checked independently for Policy CRDs

## Tools

### policy.policies_list

**Description**: List policies from Kyverno or Gatekeeper.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Policy (CRDs required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | "" | Namespace name (for Kyverno namespaced Policy) |
| `engine` | string | No | "all" | Policy engine: "kyverno", "gatekeeper", or "all" |

#### Output Schema

**Success**:
```json
{
  "items": [
    {
      "engine": "kyverno",
      "kind": "ClusterPolicy",
      "name": "require-labels",
      "ready": true,
      "active": true,
      "message": "Policy is ready"
    },
    {
      "engine": "gatekeeper",
      "kind": "ConstraintTemplate",
      "name": "k8srequiredlabels",
      "ready": true,
      "active": true
    }
  ]
}
```

### policy.policy_get

**Description**: Get details for a specific policy.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Policy (CRDs required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `engine` | string | Yes | - | Policy engine: "kyverno" or "gatekeeper" |
| `kind` | string | Yes | - | Policy kind (e.g., "ClusterPolicy", "Policy", "ConstraintTemplate") |
| `name` | string | Yes | - | Policy name |
| `namespace` | string | No | "" | Namespace name (required for namespaced policies) |
| `raw` | boolean | No | false | Return raw object if true |

#### Output Schema

**Success**:
```json
{
  "summary": {
    "engine": "kyverno",
    "kind": "ClusterPolicy",
    "name": "require-labels",
    "ready": true,
    "active": true,
    "message": "Policy is ready"
  },
  "rules": ["check-labels", "check-annotations"]
}
```

### policy.violations_list

**Description**: List policy violations from Kyverno PolicyReports or Gatekeeper constraints.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Policy (CRDs required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `namespace` | string | No | "" | Namespace name (empty for all namespaces) |
| `engine` | string | No | "all" | Policy engine: "kyverno", "gatekeeper", or "all" |
| `limit` | integer | No | 0 | Maximum number of items to return (0 = no limit) |
| `continue` | string | No | "" | Token from previous paginated request |

#### Output Schema

**Success**:
```json
{
  "items": [
    {
      "engine": "kyverno",
      "policy": "require-labels",
      "rule": "check-labels",
      "resource": "v1/Pod",
      "name": "my-pod",
      "namespace": "default",
      "message": "validation error: required label 'app' is missing",
      "timestamp": "2024-01-15T10:30:00Z",
      "severity": "error"
    }
  ],
  "continue": "eyJ2IjoibWV0YS5rOHMuaW8vdjEiLCJydiI6MjE0NzQ4MzY0N30"
}
```

**Note**: The `continue` field is only present when pagination is used and more results are available.

### policy.explain_denial

**Description**: Explain an admission denial message by attempting to match it to policies. **Note**: This is a heuristic tool and may not always find matches.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Policy (CRDs required)

#### Input Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `context` | string | No | default | Kubeconfig context name |
| `message` | string | Yes | - | Admission denial message or event text |

#### Output Schema

**Success**:
```json
{
  "matches": [
    {
      "engine": "kyverno",
      "policy": "require-labels",
      "rule": "unknown",
      "confidence": 0.5,
      "explanation": "Message contains policy name 'require-labels'"
    }
  ]
}
```

**Note**: If no matches are found, returns an empty array (not an error).

## Examples

### List all policies

```json
{
  "tool": "policy.policies_list",
  "arguments": {
    "context": "dev-cluster",
    "engine": "all"
  }
}
```

### Get a specific Kyverno policy

```json
{
  "tool": "policy.policy_get",
  "arguments": {
    "context": "dev-cluster",
    "engine": "kyverno",
    "kind": "ClusterPolicy",
    "name": "require-labels"
  }
}
```

### List violations

```json
{
  "tool": "policy.violations_list",
  "arguments": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}
```

### Explain a denial message

```json
{
  "tool": "policy.explain_denial",
  "arguments": {
    "context": "dev-cluster",
    "message": "admission webhook \"validate.kyverno.svc\" denied the request: policy require-labels: validation error: required label 'app' is missing"
  }
}
```

## Error Codes

- **FeatureNotInstalled**: Returned if toolset is disabled (CRDs not detected) or if a specific CRD is missing (e.g., Kyverno CRDs not available when querying kyverno engine)
- **KubernetesError**: Returned for Kubernetes API errors (resource not found, permission denied, etc.)

## Notes

- The `explain_denial` tool uses heuristic matching and may not always accurately identify the policy causing a denial
- Gatekeeper violations extraction is best-effort and may not capture all violations if constraint CRDs are not installed
- PolicyReport CRDs are optional but enhance violations functionality for Kyverno
