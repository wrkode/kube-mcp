# Release Notes - kube-mcp v1.0.0

**Release Date:** December 2025  
**Status:** Stable Release  

---

## Welcome to kube-mcp v1.0.0

We're excited to announce the first stable release of **kube-mcp**, a production-grade Model Context Protocol (MCP) server for Kubernetes management. This release represents the completion of all planned features from our initial roadmap and marks kube-mcp as ready for production use.

---

## What's New

### Transport Updates
- **SSE Transport Removed**: The deprecated SSE transport has been removed. HTTP transport now uses Streamable HTTP (via `mcp.NewStreamableHTTPHandler`) for efficient bidirectional communication.
- **Streamable HTTP**: The HTTP transport provides full streaming capabilities, replacing the need for a separate SSE transport.

### Complete Feature Set

kube-mcp v1.0.0 includes **60+ MCP tools** across **13 toolsets**, providing comprehensive Kubernetes management capabilities:

#### Core Kubernetes Operations
- Full CRUD operations for any Kubernetes resource (including CRDs)
- Server-side apply with field manager support
- Resource patching (merge, JSON, strategic merge)
- Resource diffing (unified, JSON, YAML formats)
- Resource validation before applying
- Resource watching with event collection
- Resource scaling with dry-run support
- Resource relationships (owners and dependents)

#### Pod Management
- List, get, delete pods
- Fetch logs with time filtering and follow support
- Execute commands in pods
- Port forwarding setup
- Resource usage metrics

#### ConfigMap & Secret Management
- Get and set ConfigMap data
- Get and set Secret data with base64 encoding/decoding
- Merge or replace operations

#### Observability
- kubectl-style resource describe
- Event listing
- Node and pod metrics
- Prometheus metrics endpoint

#### Additional Toolsets
- **Helm** - Chart and release management
- **GitOps** - Flux and Argo CD application management (auto-enabled when CRDs detected)
- **Policy** - Kyverno and Gatekeeper policy visibility (auto-enabled when CRDs detected)
- **CAPI** - Cluster API cluster lifecycle management (auto-enabled when CRDs detected)
- **Rollouts** - Progressive delivery management (Argo Rollouts and Flagger) (auto-enabled when CRDs detected)
- **Certs** - Cert-Manager certificate lifecycle management (auto-enabled when CRDs detected)
- **Autoscaling** - HPA and KEDA autoscaling management (HPA always available, KEDA auto-enabled when CRDs detected)
- **Backup** - Velero backup and restore management (auto-enabled when CRDs detected)
- **Network** - NetworkPolicy, Cilium, and Hubble network observability (NetworkPolicy always available)
- **KubeVirt** - VM lifecycle operations (auto-enabled when CRDs detected)
- **Kiali** - Service mesh observability (optional, when configured)
- **Config** - Kubeconfig inspection and context management

---

## Key Features

### Multi-Cluster Support
Manage multiple Kubernetes clusters from a single kube-mcp instance. Switch between clusters using the `context` parameter in any tool.

### Security First
- **RBAC Integration**: All destructive operations check permissions before execution
- **RBAC Caching**: Configurable TTL-based caching for performance
- **TokenReview**: Bearer token validation via Kubernetes API
- **Distroless Container**: Minimal attack surface
- **Non-root User**: Container runs as non-root

### Observability
- **Structured Logging**: JSON and text formats
- **Prometheus Metrics**: Comprehensive metrics for monitoring
- **Health Endpoints**: `/health` and `/healthz` for health checks
- **Panic Recovery**: Automatic error recovery and reporting

### Multiple Transports
- **STDIO** (default): Standard input/output for MCP clients
- **HTTP**: Streamable HTTP transport with OAuth2/OIDC support (replaces deprecated SSE transport)

### Client Compatibility
- **Tool Name Normalization**: Optional normalization of tool names for clients that don't support dots in function names (e.g., n8n)
  - Enable with `normalize_tool_names = true` in `[server]` configuration
  - Automatically replaces dots with underscores (e.g., `autoscaling.hpa_explain` → `autoscaling_hpa_explain`)
  - Maintains backward compatibility - disabled by default

### Dynamic CRD Discovery
Automatically detects and supports Custom Resource Definitions without configuration. Multiple toolsets auto-enable when their respective CRDs are detected:
- **GitOps**: Flux (Kustomization, HelmRelease) or Argo CD (Application)
- **Policy**: Kyverno (ClusterPolicy, Policy) or Gatekeeper (ConstraintTemplate)
- **CAPI**: Cluster API (Cluster, Machine, MachineDeployment)
- **Rollouts**: Argo Rollouts (Rollout) or Flagger (Canary)
- **Certs**: Cert-Manager (Certificate, Issuer, ClusterIssuer)
- **Autoscaling**: KEDA (ScaledObject) - HPA always available
- **Backup**: Velero (Backup, Restore, BackupStorageLocation)
- **Network**: Cilium (CiliumNetworkPolicy) - NetworkPolicy always available
- **KubeVirt**: KubeVirt (VirtualMachine)

---

## Phase 1: Foundation Features

### Dry-Run Support
Validate operations without making changes:
- `resources_apply` with `dry_run: true`
- `resources_delete` with `dry_run: true`
- `resources_scale` with `dry_run: true`
- `resources_patch` with `dry_run: true`

### Resource Patch Tool
Partial resource updates with three patch types:
- **Merge Patch** (RFC 7396): Default, merges changes
- **JSON Patch** (RFC 6902): Array of operations
- **Strategic Merge Patch**: Kubernetes-aware merging

### Resource Diff Tool
Compare current vs desired state:
- **Unified Format**: Shows current, desired, and patch
- **JSON Format**: JSON patch that would be applied
- **YAML Format**: YAML comparison

---

## Phase 2: Streaming Features

### Log Follow
Stream logs in real-time:
```bash
# Follow logs for a pod
Follow logs for pod my-pod in namespace default
```

### Port Forwarding
Set up port forwarding:
```bash
# Forward local port 8080 to pod port 80
Forward port 8080 to pod my-pod port 80
```

### Resource Watch
Monitor resources for changes:
```bash
# Watch deployments for 30 seconds
Watch deployments in namespace default for 30 seconds
```

---

## Phase 3: Polish Features

### Resource Validation
Validate manifests before applying:
```bash
# Validate a deployment manifest
Validate this deployment manifest before applying it
```

### ConfigMap/Secret Tools
Specialized operations:
```bash
# Get ConfigMap data
Get all data from ConfigMap my-config

# Update Secret
Update Secret my-secret with password=secret123
```

### Resource Relationships
Find resource dependencies:
```bash
# Find what owns a pod
Show me what owns pod my-pod

# Find what depends on a deployment
What pods are owned by deployment my-app?
```

---

## Phase 4: Domain-Specific Toolsets

### GitOps Toolset
Manage Flux and Argo CD applications:
- **List Applications**: `gitops.apps_list` - List Kustomizations, HelmReleases, or Applications
- **Get Details**: `gitops.app_get` - Get application status and details
- **Trigger Reconciliation**: `gitops.app_reconcile` - Force reconciliation for Flux apps

**Example:**
```json
{
  "tool": "gitops.apps_list",
  "params": {
    "namespace": "flux-system",
    "kinds": ["Kustomization", "HelmRelease"]
  }
}
```

### Policy Toolset
Visibility into policy engines:
- **List Policies**: `policy.policies_list` - List Kyverno or Gatekeeper policies
- **Get Policy**: `policy.policy_get` - Get policy details
- **List Violations**: `policy.violations_list` - List policy violations
- **Explain Denial**: `policy.explain_denial` - Heuristic explanation of admission denials

**Example:**
```json
{
  "tool": "policy.violations_list",
  "params": {
    "namespace": "default",
    "policy_name": "require-labels"
  }
}
```

### CAPI Toolset
Cluster API cluster management:
- **List Clusters**: `capi.clusters_list` - List CAPI clusters
- **Get Cluster**: `capi.cluster_get` - Get cluster details
- **List Machines**: `capi.machines_list` - List machines for a cluster
- **List Machine Deployments**: `capi.machinedeployments_list` - List machine deployments
- **Rollout Status**: `capi.rollout_status` - Get cluster rollout status
- **Scale Machine Deployment**: `capi.scale_machinedeployment` - Scale worker nodes

**Example:**
```json
{
  "tool": "capi.clusters_list",
  "params": {
    "context": "management-cluster"
  }
}
```

### Rollouts Toolset
Progressive delivery management:
- **List Rollouts**: `rollouts.list` - List Argo Rollouts or Flagger Canaries
- **Get Status**: `rollouts.get_status` - Get detailed rollout status
- **Promote**: `rollouts.promote` - Promote rollout to next step (Argo Rollouts)
- **Abort**: `rollouts.abort` - Abort a rollout (Argo Rollouts)
- **Retry**: `rollouts.retry` - Retry rollout analysis (Argo Rollouts)

**Example:**
```json
{
  "tool": "rollouts.promote",
  "params": {
    "kind": "Rollout",
    "name": "my-app",
    "namespace": "default",
    "confirm": true
  }
}
```

### Certs Toolset
Cert-Manager certificate management:
- **List Certificates**: `certs.certificates_list` - List certificates
- **Get Certificate**: `certs.certificate_get` - Get certificate details
- **List Issuers**: `certs.issuers_list` - List issuers and cluster issuers
- **Status Explain**: `certs.status_explain` - Explain certificate status with diagnosis hints
- **Renew Certificate**: `certs.renew` - Trigger certificate renewal
- **List ACME Challenges**: `certs.acme_challenges_list` - List ACME challenges/orders

**Example:**
```json
{
  "tool": "certs.status_explain",
  "params": {
    "name": "my-cert",
    "namespace": "cert-manager"
  }
}
```

### Autoscaling Toolset
HPA and KEDA autoscaling:
- **List HPAs**: `autoscaling.hpa_list` - List HorizontalPodAutoscalers (always available)
- **Explain HPA**: `autoscaling.hpa_explain` - Explain HPA status and metrics
- **List KEDA ScaledObjects**: `autoscaling.keda_scaledobjects_list` - List KEDA ScaledObjects
- **Get ScaledObject**: `autoscaling.keda_scaledobject_get` - Get ScaledObject details
- **Explain Triggers**: `autoscaling.keda_triggers_explain` - Explain KEDA triggers
- **Pause/Resume**: `autoscaling.keda_pause`, `autoscaling.keda_resume` - Control autoscaling

**Example:**
```json
{
  "tool": "autoscaling.hpa_explain",
  "params": {
    "name": "my-hpa",
    "namespace": "default"
  }
}
```

### Backup Toolset
Velero backup and restore:
- **List Backups**: `backup.backups_list` - List Velero backups
- **Get Backup**: `backup.backup_get` - Get backup details
- **Create Backup**: `backup.backup_create` - Create a new backup
- **List Restores**: `backup.restores_list` - List Velero restores
- **Create Restore**: `backup.restore_create` - Create a restore from backup
- **List Storage Locations**: `backup.locations_list` - List backup storage locations

**Example:**
```json
{
  "tool": "backup.backup_create",
  "params": {
    "namespace": "velero",
    "included_namespaces": ["default", "production"],
    "snapshot_volumes": true,
    "confirm": true
  }
}
```

### Network Toolset
Network policy and observability:
- **List NetworkPolicies**: `net.networkpolicies_list` - List NetworkPolicies (always available)
- **Explain NetworkPolicy**: `net.networkpolicy_explain` - Explain policy rules in normalized format
- **Connectivity Hint**: `net.connectivity_hint` - Analyze connectivity between pods (best-effort)
- **List Cilium Policies**: `net.cilium_policies_list` - List Cilium policies (CRD-gated)
- **Get Cilium Policy**: `net.cilium_policy_get` - Get Cilium policy details
- **Query Hubble Flows**: `net.hubble_flows_query` - Query Hubble flow data (requires Hubble API)

**Example:**
```json
{
  "tool": "net.connectivity_hint",
  "params": {
    "src_namespace": "default",
    "src_labels": {"app": "frontend"},
    "dst_namespace": "default",
    "dst_labels": {"app": "backend"},
    "port": "8080",
    "protocol": "TCP"
  }
}
```

---

## Installation

### Helm Chart (Recommended for Kubernetes)

The easiest way to deploy kube-mcp on Kubernetes is using the included Helm chart:

```bash
# Install from local chart
helm install kube-mcp ./charts/kube-mcp

# Or install with custom values
helm install kube-mcp ./charts/kube-mcp \
  --set server.transports[0]=http \
  --set kubernetes.provider=in-cluster
```

The Helm chart includes:
- Deployment with configurable replicas
- Service for HTTP transport
- ConfigMap for configuration
- ServiceAccount and RBAC (ClusterRole/ClusterRoleBinding)
- Optional Ingress, HPA, PDB, and ServiceMonitor support
- Comprehensive values.yaml with all configuration options

See [charts/kube-mcp/README.md](charts/kube-mcp/README.md) for detailed documentation.

### From Source
```bash
go install github.com/wrkode/kube-mcp/cmd/kube-mcp@latest
```

### Docker
```bash
docker pull ghcr.io/wrkode/kube-mcp:1.0.0
docker run -v /path/to/config.toml:/etc/kube-mcp/config.toml:ro \
           -p 8080:8080 \
           ghcr.io/wrkode/kube-mcp:1.0.0
```

### Binary Releases
Download pre-built binaries from [GitHub Releases](https://github.com/wrkode/kube-mcp/releases/tag/v1.0.0).

---

## Quick Start

### STDIO Mode (Default)
```bash
kube-mcp --config /path/to/config.toml
```

### HTTP Mode
```bash
kube-mcp --transport http --config /path/to/config.toml
```

### Configuration
See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for detailed configuration options.

**New in v1.0.0:**
- **Tool Name Normalization**: Add `normalize_tool_names = true` to `[server]` section to enable compatibility with clients like n8n that don't support dots in function names. When enabled, tool names like `autoscaling.hpa_explain` are automatically normalized to `autoscaling_hpa_explain`.

---

## Documentation

- **[Architecture](docs/ARCHITECTURE.md)** - System architecture overview
- **[Configuration](docs/CONFIGURATION.md)** - Configuration guide
- **[Tools](docs/TOOLS.md)** - Complete tool reference
- **[Security](docs/SECURITY.md)** - Security and RBAC documentation
- **[Field Testing Guides](field_testing/)** - Comprehensive feature guides

---

## Breaking Changes

**None** - This release maintains full backward compatibility with v0.1.0.

---

## Migration Guide

### From v0.1.0

No migration required! All existing configurations and tool calls remain compatible.

### New Features

To take advantage of new features:

1. **Dry-Run**: Add `dry_run: true` to apply, delete, scale, or patch operations
2. **Patch Tool**: Use `resources_patch` for partial updates instead of full apply
3. **Diff Tool**: Use `resources_diff` to preview changes before applying
4. **Watch**: Use `resources_watch` for real-time monitoring
5. **Log Follow**: Add `follow: true` to `pods_logs` for streaming
6. **ConfigMap/Secret Tools**: Use specialized tools for data operations
7. **Tool Name Normalization**: Enable `normalize_tool_names = true` in `[server]` section for n8n and other clients that don't support dots in function names

---

## Known Limitations

1. **Port Forwarding**: Setup validation only; full forwarding requires HTTP transport or kubectl
3. **Watch Events**: Limited to 1000 events per watch to prevent memory issues
4. **Metrics**: Requires metrics-server to be installed in cluster

---

## Performance

- **RBAC Caching**: Reduces Kubernetes API calls with configurable TTL (default: 5s)
- **Connection Pooling**: Efficient client connection management
- **Pagination**: Handles large result sets efficiently
- **Graceful Degradation**: Handles missing dependencies gracefully

---

## Security

- **RBAC Checks**: All destructive operations verify permissions
- **Token Validation**: Bearer tokens validated via Kubernetes TokenReview API
- **Distroless Image**: Minimal attack surface
- **Non-root User**: Container runs as non-root (UID 65532)

---

## Testing

All features are covered by comprehensive integration tests:
- Core resource operations
- Pod operations
- ConfigMap/Secret operations
- Dry-run operations
- Patch operations
- Diff operations
- Watch operations
- Validation operations
- Relationship operations
- GitOps toolset (Flux and Argo CD)
- Policy toolset (Kyverno and Gatekeeper)
- CAPI toolset (Cluster API)
- Rollouts toolset (Argo Rollouts and Flagger)
- Certs toolset (Cert-Manager)
- Autoscaling toolset (HPA and KEDA)
- Backup toolset (Velero)
- Network toolset (NetworkPolicy, Cilium, Hubble)

Run tests:
```bash
make test-integration
```

---

## Contributors

Thank you to all contributors who made this release possible!

---

## Helm Chart

### New in v1.0.0

A comprehensive Helm chart is now included for easy Kubernetes deployment:

- **Location**: `charts/kube-mcp/`
- **Chart Version**: 1.0.0
- **Features**:
  - Production-ready deployment configuration
  - Configurable RBAC with ClusterRole
  - Support for STDIO and HTTP (Streamable HTTP) transports
  - Optional Ingress, HPA, and ServiceMonitor
  - Comprehensive values.yaml with all options
  - Full documentation in chart README

### Quick Start

```bash
# Install with default values
helm install kube-mcp ./charts/kube-mcp

# Install with custom configuration
helm install kube-mcp ./charts/kube-mcp \
  --set server.transports[0]=http \
  --set kubernetes.provider=in-cluster \
  --set security.readOnly=false
```

### Chart Release

The Helm chart is automatically packaged and included in GitHub releases. Download `kube-mcp-1.0.0.tgz` from the release assets.

---

## What's Next

With v1.0.0 complete, future releases will focus on:
- Additional toolsets (based on community feedback)
- Performance optimizations
- Extended documentation
- Helm chart repository setup
- Enhanced error messages and diagnostics
- Additional mutation operations for progressive delivery
- Extended backup/restore capabilities

---

## Support

- **Issues**: [GitHub Issues](https://github.com/wrkode/kube-mcp/issues)
- **Documentation**: [docs/](docs/)
- **Discussions**: [GitHub Discussions](https://github.com/wrkode/kube-mcp/discussions)

---

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

---

**Full Changelog**: [CHANGELOG.md](./CHANGELOG.md)

