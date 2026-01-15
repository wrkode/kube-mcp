# Network Toolset

## Overview

The Network toolset provides tools for managing NetworkPolicies, Cilium policies, and querying Hubble flows. NetworkPolicy tools are always available (native Kubernetes), while Cilium and Hubble tools require additional dependencies.

## Dependencies

- **NetworkPolicy**: Always available (native Kubernetes API)
- **Cilium**: Requires `cilium.io/v2/CiliumNetworkPolicy` or `CiliumClusterwideNetworkPolicy` CRDs
- **Hubble**: Requires `hubble_api_url` configuration

## Tools

### net.networkpolicies_list

**Description**: List NetworkPolicies.

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

### net.networkpolicy_explain

**Description**: Explain NetworkPolicy rules in a normalized format.

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Output Schema

```json
{
  "name": "my-policy",
  "namespace": "default",
  "ingress": [
    {
      "direction": "ingress",
      "peers": [
        {
          "pod_selector": {"app": "frontend"},
          "namespace_selector": {"name": "production"}
        }
      ],
      "ports": [
        {"protocol": "TCP", "port": "80"}
      ]
    }
  ],
  "egress": []
}
```

### net.connectivity_hint

**Description**: Analyze connectivity between pods (best-effort).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: No

#### Input Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `context` | string | No | Kubeconfig context name |
| `src_namespace` | string | Yes | Source namespace |
| `src_labels` | object | Yes | Source pod labels |
| `dst_namespace` | string | Yes | Destination namespace |
| `dst_labels` | object | Yes | Destination pod labels |
| `port` | string | Yes | Port number |
| `protocol` | string | Yes | Protocol (TCP, UDP, SCTP) |

#### Output Schema

```json
{
  "likely_allowed": "true",
  "reasons": [
    "Egress rule in default/my-policy may allow traffic"
  ],
  "evaluated_policies": [
    "default/my-policy",
    "default/other-policy"
  ]
}
```

**Note**: This is a best-effort analysis. Actual policy evaluation is complex and depends on pod selectors, IP blocks, and other factors.

### net.cilium_policies_list

**Description**: List Cilium network policies (requires Cilium CRDs).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Cilium (CRDs required)

### net.cilium_policy_get

**Description**: Get Cilium policy details (requires Cilium CRDs).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Cilium (CRDs required)

### net.hubble_flows_query

**Description**: Query Hubble flows (requires Hubble API configuration).

**Read-only**: Yes  
**Destructive**: No  
**Cluster-aware**: Yes  
**Feature-gated**: Hubble (API URL required)

#### Input Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `context` | string | No | Kubeconfig context name |
| `namespace` | string | No | Filter by namespace |
| `pod` | string | No | Filter by pod name |
| `verdict` | string | No | Filter by verdict |
| `protocol` | string | No | Filter by protocol |
| `since_seconds` | integer | No | Query flows since N seconds ago |
| `limit` | integer | No | Maximum number of flows |

#### Output Schema

```json
{
  "flows": [
    {
      "time": "2024-01-15T10:30:00Z",
      "verdict": "FORWARDED",
      "source": {
        "namespace": "default",
        "pod": "frontend-pod",
        "ip": "10.0.0.1"
      },
      "destination": {
        "namespace": "default",
        "pod": "backend-pod",
        "ip": "10.0.0.2"
      },
      "protocol": "TCP",
      "port": 8080
    }
  ]
}
```

## Error Codes

- **FeatureNotInstalled**: Cilium CRDs not detected (for Cilium tools)
- **FeatureDisabled**: Hubble API not configured (for Hubble tool)
- **ExternalServiceUnavailable**: Hubble API unreachable
- **KubernetesError**: Kubernetes API errors
