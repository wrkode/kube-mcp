# kube-mcp Tools Reference

## Overview

kube-mcp is a Model Context Protocol (MCP) server that exposes comprehensive Kubernetes management capabilities through a set of tools organized into toolsets. This document provides an overview of all available tools and links to detailed documentation for each toolset.

> **Looking for usage examples?** See the [Usage Guide](USAGE_GUIDE.md) for practical examples and workflows.
> 
> **Need multi-cluster help?** See the [Multi-Cluster Guide](MULTI_CLUSTER.md) for cluster targeting and context management.

## Toolsets

kube-mcp organizes tools into the following toolsets:

- **config** - Kubeconfig inspection and context management
- **core** - Core Kubernetes operations (pods, resources, namespaces, nodes, events, metrics, scaling)
- **helm** - Helm chart and release management
- **gitops** - GitOps application management (Flux and Argo CD) (optional, requires GitOps CRDs)
- **policy** - Policy engine visibility (Kyverno and Gatekeeper) (optional, requires Policy CRDs)
- **capi** - Cluster API cluster lifecycle management (optional, requires CAPI CRDs)
- **rollouts** - Progressive delivery management (Argo Rollouts and Flagger) (optional, requires Rollouts CRDs)
- **certs** - Cert-Manager certificate lifecycle management (optional, requires Cert-Manager CRDs)
- **autoscaling** - HPA and KEDA autoscaling management (HPA always available, KEDA optional)
- **backup** - Velero backup and restore management (optional, requires Velero CRDs)
- **net** - NetworkPolicy, Cilium, and Hubble network observability (NetworkPolicy always available)
- **kubevirt** - KubeVirt VM lifecycle management (optional, requires KubeVirt CRDs)
- **kiali** - Service mesh observability via Kiali (optional, requires Kiali configuration)

## Cluster Targeting

All cluster-aware tools support multi-cluster operations through the `ClusterTarget` pattern. Each tool accepts an optional `context` parameter that specifies which Kubernetes cluster context to target.

### ClusterTarget Input

```json
{
  "context": "string (optional)"
}
```

- **context**: The kubeconfig context name to target. If omitted or empty, uses the provider's default context.
- For in-cluster providers, the context parameter is ignored.
- For single-cluster providers, only the configured context is allowed.

See [Cluster Targeting](tools/core.md#cluster-targeting) for detailed information.

## Read-Only vs Destructive Tools

Tools are categorized as:

- **Read-only**: Tools that only retrieve information and do not modify cluster state
- **Destructive**: Tools that modify cluster state (create, update, delete, scale)

This distinction helps IDEs and agents understand which tools require caution.

## Feature Gating

Some tools require specific dependencies:

- **Metrics Server**: `pods_top` and `nodes_top` require the `metrics.k8s.io` API (metrics-server)
- **GitOps CRDs**: All `gitops.*` tools require Flux or Argo CD CRDs to be installed
- **Policy CRDs**: All `policy.*` tools require Kyverno or Gatekeeper CRDs to be installed
- **CAPI CRDs**: All `capi.*` tools require Cluster API CRDs to be installed
- **Rollouts CRDs**: All `rollouts.*` tools require Argo Rollouts or Flagger CRDs to be installed
- **Cert-Manager CRDs**: All `certs.*` tools require Cert-Manager CRDs to be installed
- **KEDA CRDs**: KEDA-specific `autoscaling.keda_*` tools require KEDA CRDs to be installed
- **Velero CRDs**: All `backup.*` tools require Velero CRDs to be installed
- **Cilium CRDs**: Cilium-specific `net.cilium_*` tools require Cilium CRDs to be installed
- **Hubble**: `net.hubble_flows_query` requires Hubble API to be configured
- **KubeVirt CRDs**: All `kubevirt_*` tools require KubeVirt CRDs to be installed
- **Kiali**: All `kiali_*` tools require Kiali to be configured and reachable

When dependencies are missing:
- Metrics tools return `MetricsUnavailable` errors but remain registered
- GitOps tools are not registered if CRDs are not detected
- Policy tools are not registered if CRDs are not detected
- CAPI tools are not registered if Cluster CRD is not detected
- Rollouts tools are not registered if CRDs are not detected
- Certs tools are not registered if Certificate CRD is not detected
- KEDA tools are not registered if ScaledObject CRD is not detected (HPA tools always available)
- Backup tools are not registered if Backup CRD is not detected
- Cilium tools are not registered if CRDs are not detected (NetworkPolicy tools always available)
- Hubble tool is not registered if Hubble API is not configured
- KubeVirt tools are not registered if CRDs are not detected
- Kiali tools are not registered if Kiali is not configured

## Tools Index

| Toolset | Tool Name | Description | Read-only | Destructive | Feature-gated |
|---------|-----------|-------------|-----------|-------------|---------------|
| config | `config_contexts_list` | List all available Kubernetes contexts from kubeconfig | [OK] | [NO] | No |
| config | `config_kubeconfig_view` | View kubeconfig file contents (full or minified) | [OK] | [NO] | No |
| core | `pods_list` | List pods in a namespace or all namespaces | [OK] | [NO] | No |
| core | `pods_get` | Get pod details | [OK] | [NO] | No |
| core | `pods_delete` | Delete a pod | [NO] | [OK] | No |
| core | `pods_logs` | Fetch pod logs | [OK] | [NO] | No |
| core | `pods_exec` | Execute command in pod | [NO] | [OK] | No |
| core | `pods_top` | Get pod resource usage metrics from metrics.k8s.io API | [OK] | [NO] | MetricsServer |
| core | `resources_list` | List resources by GroupVersionKind | [OK] | [NO] | No |
| core | `resources_get` | Get a resource | [OK] | [NO] | No |
| core | `resources_apply` | Create or update a resource using server-side apply | [NO] | [OK] | No |
| core | `resources_delete` | Delete a resource | [NO] | [OK] | No |
| core | `resources_scale` | Scale a resource (get or change replicas) | [NO] | [OK] | No |
| core | `namespaces_list` | List all namespaces | [OK] | [NO] | No |
| core | `nodes_top` | Get node resource usage metrics from metrics.k8s.io API | [OK] | [NO] | MetricsServer |
| core | `nodes_summary` | Get node summary statistics | [OK] | [NO] | No |
| core | `events_list` | List events | [OK] | [NO] | No |
| helm | `helm_install` | Install a Helm chart | [NO] | [OK] | No |
| helm | `helm_releases_list` | List Helm releases | [OK] | [NO] | No |
| helm | `helm_uninstall` | Uninstall a Helm release | [NO] | [OK] | No |
| gitops | `gitops.apps_list` | List GitOps applications (Flux/Argo CD) | [OK] | [NO] | GitOps |
| gitops | `gitops.app_get` | Get GitOps application details | [OK] | [NO] | GitOps |
| gitops | `gitops.app_reconcile` | Trigger reconciliation for Flux app | [NO] | [OK] | GitOps |
| policy | `policy.policies_list` | List policies (Kyverno/Gatekeeper) | [OK] | [NO] | Policy |
| policy | `policy.policy_get` | Get policy details | [OK] | [NO] | Policy |
| policy | `policy.violations_list` | List policy violations | [OK] | [NO] | Policy |
| policy | `policy.explain_denial` | Explain admission denial (heuristic) | [OK] | [NO] | Policy |
| capi | `capi.clusters_list` | List Cluster API clusters | [OK] | [NO] | CAPI |
| capi | `capi.cluster_get` | Get Cluster API cluster details | [OK] | [NO] | CAPI |
| capi | `capi.machines_list` | List machines for a cluster | [OK] | [NO] | CAPI |
| capi | `capi.machinedeployments_list` | List machine deployments for a cluster | [OK] | [NO] | CAPI |
| capi | `capi.rollout_status` | Get cluster rollout status | [OK] | [NO] | CAPI |
| capi | `capi.scale_machinedeployment` | Scale a machine deployment | [NO] | [OK] | CAPI |
| kubevirt | `kubevirt_vm_create` | Create a VirtualMachine | [NO] | [OK] | KubeVirt |
| kubevirt | `kubevirt_vm_start` | Start a VirtualMachine | [NO] | [OK] | KubeVirt |
| kubevirt | `kubevirt_vm_stop` | Stop a VirtualMachine | [NO] | [OK] | KubeVirt |
| kubevirt | `kubevirt_vm_restart` | Restart a VirtualMachine | [NO] | [OK] | KubeVirt |
| kubevirt | `kubevirt_datasources_list` | List KubeVirt DataSources | [OK] | [NO] | KubeVirt |
| kubevirt | `kubevirt_instancetypes_list` | List KubeVirt InstanceTypes | [OK] | [NO] | KubeVirt |
| kiali | `kiali_mesh_graph` | Get service mesh graph | [OK] | [NO] | Kiali |
| kiali | `kiali_istio_config_get` | Get Istio configuration | [OK] | [NO] | Kiali |
| kiali | `kiali_metrics` | Get metrics | [OK] | [NO] | Kiali |
| kiali | `kiali_logs` | Get logs | [OK] | [NO] | Kiali |
| kiali | `kiali_traces` | Get traces | [OK] | [NO] | Kiali |

## Detailed Documentation

- [Core Toolset](tools/core.md) - Pods, resources, namespaces, nodes, events, metrics, scaling
- [Config Toolset](tools/config.md) - Kubeconfig inspection
- [Helm Toolset](tools/helm.md) - Chart and release management
- [GitOps Toolset](tools/gitops.md) - Flux and Argo CD application management
- [Policy Toolset](tools/policy.md) - Kyverno and Gatekeeper policy visibility
- [CAPI Toolset](tools/capi.md) - Cluster API cluster lifecycle management
- [Rollouts Toolset](tools/rollouts.md) - Progressive delivery management
- [Certs Toolset](tools/certs.md) - Cert-Manager certificate lifecycle
- [Autoscaling Toolset](tools/autoscaling.md) - HPA and KEDA autoscaling
- [Backup Toolset](tools/backup.md) - Velero backup and restore
- [Network Toolset](tools/net.md) - NetworkPolicy, Cilium, and Hubble
- [KubeVirt Toolset](tools/kubevirt.md) - VM lifecycle management
- [Kiali Toolset](tools/kiali.md) - Service mesh observability
- [Error Contract](tools/errors.md) - Error codes, shapes, and semantics

## Error Handling

All tools follow a consistent error contract. See [Error Contract](tools/errors.md) for details on error codes, shapes, and handling.

Common error types:
- `KubernetesError` - Kubernetes API errors
- `MetricsUnavailable` - Metrics server not available
- `ScalingNotSupported` - Resource does not support scaling
- `UnknownContext` - Cluster context not found
- `FeatureDisabled` - Feature disabled via configuration or server mode
- `FeatureNotInstalled` - Required CRD/API not present in cluster
- `ExternalServiceUnavailable` - Required external service unreachable
- `ValidationError` - Input validation failed
