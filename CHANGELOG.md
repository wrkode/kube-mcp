# Changelog

All notable changes to kube-mcp will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-XX

### First Stable Release

This is the first stable release of kube-mcp, marking the completion of all planned features from the initial roadmap. kube-mcp is now production-ready with comprehensive Kubernetes management capabilities through the Model Context Protocol.

### Added

#### Core Features
- **Multi-cluster support** - Manage multiple Kubernetes clusters from a single server instance
- **Multiple transports** - STDIO (default), HTTP, and SSE transport support
- **Dynamic CRD discovery** - Automatic detection and support for Custom Resource Definitions
- **Server-side apply** - Safe resource reconciliation using Kubernetes field managers
- **RBAC integration** - Comprehensive RBAC checks with caching for performance
- **Observability** - Structured logging, Prometheus metrics, and panic recovery

#### Core Toolset (27+ tools)
- **Resource Management**
  - `resources_list` - List any Kubernetes resource with label/field selectors and pagination
  - `resources_get` - Get detailed resource information
  - `resources_apply` - Apply resources using server-side apply with dry-run support
  - `resources_delete` - Delete resources with dry-run support
  - `resources_scale` - Scale resources (Deployments, StatefulSets, etc.) with dry-run support
  - `resources_patch` - Patch resources using merge, JSON, or strategic merge patches
  - `resources_diff` - Compare current vs desired state with unified/JSON/YAML formats
  - `resources_describe` - kubectl-style formatted output with events and conditions
  - `resources_validate` - Validate manifests before applying
  - `resources_watch` - Watch resources for changes (ADDED, MODIFIED, DELETED events)
  - `resources_relationships` - Find resource owners and dependents

- **Pod Operations**
  - `pods_list` - List pods with label/field selectors and pagination
  - `pods_get` - Get pod details
  - `pods_delete` - Delete pods
  - `pods_logs` - Fetch pod logs with time filtering, previous logs, and follow support
  - `pods_exec` - Execute commands in pods
  - `pods_top` - Get pod resource usage metrics
  - `pods_port_forward` - Set up port forwarding from local to pod ports

- **Namespace & Node Operations**
  - `namespaces_list` - List all namespaces
  - `nodes_top` - Get node resource usage metrics
  - `nodes_summary` - Get node summary statistics

- **ConfigMap & Secret Operations**
  - `configmaps_get_data` - Get ConfigMap data (all or specific keys)
  - `configmaps_set_data` - Update ConfigMap data (merge or replace)
  - `secrets_get_data` - Get Secret data with base64 decode option
  - `secrets_set_data` - Update Secret data with automatic base64 encoding

- **Events**
  - `events_list` - List events in namespace or cluster-wide

#### Helm Toolset
- `helm_releases_list` - List Helm releases
- `helm_charts_list` - List available Helm charts
- `helm_releases_get` - Get Helm release details

#### KubeVirt Toolset (auto-enabled when CRDs detected)
- `kubevirt_vms_list` - List VirtualMachines
- `kubevirt_vms_get` - Get VirtualMachine details
- `kubevirt_vms_create` - Create VirtualMachines
- `kubevirt_vms_delete` - Delete VirtualMachines
- `kubevirt_vms_start` - Start VirtualMachines
- `kubevirt_vms_stop` - Stop VirtualMachines

#### Kiali Toolset (optional, when configured)
- `kiali_services_list` - List services in service mesh
- `kiali_services_get` - Get service details
- `kiali_graph` - Get service mesh graph

#### Config Toolset
- `config_contexts_list` - List available Kubernetes contexts
- `config_kubeconfig_view` - View kubeconfig contents

### Enhanced Features

#### Phase 0: Foundation Enhancements
- **Label/Field Selector Support** - Filter resources efficiently at the API level
- **Pagination Support** - Handle large result sets with limit and continue tokens
- **Enhanced Log Fetching** - Time-based filtering (`since`, `since_time`) and previous logs
- **Resources Describe Tool** - kubectl-style formatted output with events

#### Phase 1: Foundation Features
- **Dry-Run Support** - Validate operations without making changes (apply, delete, scale, patch)
- **Resource Patch Tool** - Partial resource updates with multiple patch types
- **Resource Diff Tool** - Compare current vs desired state with multiple formats

#### Phase 2: Streaming Features
- **Log Follow** - Real-time log streaming with `follow` parameter
- **Port Forwarding** - Set up port forwarding between local and pod ports
- **Resource Watch** - Monitor resources for changes with event collection

#### Phase 3: Polish Features
- **Resource Validation** - Validate manifests before applying
- **ConfigMap/Secret Tools** - Specialized operations for ConfigMaps and Secrets
- **Resource Relationships** - Find resource owners and dependents

### Security

- **RBAC Integration** - All destructive operations check permissions before execution
- **RBAC Caching** - Configurable TTL-based caching for performance
- **TokenReview Support** - Bearer token validation via Kubernetes TokenReview API
- **Credential Selection** - Support for multiple authentication methods
- **Distroless Container** - Minimal attack surface with distroless base image
- **Non-root User** - Container runs as non-root user (UID 65532)

### Observability

- **Structured Logging** - JSON and text formats with configurable levels
- **Prometheus Metrics** - Comprehensive metrics for tool calls, latency, errors
- **Panic Recovery** - Automatic panic recovery with error reporting
- **HTTP Metrics** - Request/response metrics for HTTP transport
- **Health Endpoints** - `/health` and `/healthz` endpoints for monitoring

### Documentation

- **Architecture Documentation** - System architecture overview
- **Configuration Guide** - Comprehensive configuration documentation
- **Tool Reference** - Complete tool documentation with examples
- **Development Guide** - Guide for developing new toolsets
- **Security Documentation** - Security and RBAC documentation
- **Field Testing Guides** - Comprehensive guides for new features
- **Quick Reference** - Quick reference for all features

### Testing

- **Integration Tests** - Comprehensive integration test suite using envtest
- **Unit Tests** - Unit tests for core functionality
- **Test Helpers** - Reusable test helpers for all toolsets
- **CI/CD Pipeline** - Automated testing and building

### Infrastructure

- **Docker Support** - Multi-architecture Docker images (amd64, arm64)
- **Helm Chart** - Production-ready Helm chart for Kubernetes deployment
- **GitHub Actions** - CI/CD pipeline with automated releases and chart packaging
- **Makefile** - Comprehensive build and test targets
- **Go Modules** - Modern Go dependency management

#### Helm Chart Features
- Deployment with configurable replicas and resources
- Service for HTTP and optional SSE transport
- ConfigMap for TOML configuration
- ServiceAccount and RBAC (ClusterRole/ClusterRoleBinding)
- Optional Ingress, HPA, PodDisruptionBudget, and ServiceMonitor support
- Comprehensive values.yaml with all configuration options
- Full documentation in chart README
- Automatic chart packaging in release workflow

### Changed

- **Version Format** - Updated from 0.1.0 to 1.0.0 to reflect stable release
- **API Stability** - All APIs are now considered stable for 1.x releases
- **Release Workflow** - Added Helm chart packaging to automated releases

### Fixed

- Fixed RBAC check for read-only scale operations (replicas: 0)
- Fixed error handling in scale operations
- Fixed TOML duration parsing for SSE heartbeat interval
- Fixed argument unmarshaling for all tool handlers
- Fixed diff tool fallback when dry-run apply fails
- Fixed port forwarding validation and error handling

### Performance

- **RBAC Caching** - Reduces Kubernetes API calls with configurable TTL
- **Connection Pooling** - Efficient Kubernetes client connection management
- **Pagination Support** - Reduces memory usage for large result sets
- **Graceful Degradation** - Handles missing dependencies (metrics server, etc.)

### Migration Notes

- No breaking changes from 0.1.0 to 1.0.0
- All existing configurations remain compatible
- New features are additive and backward compatible

---

## [0.1.0] - 2024-XX-XX

### Initial Release

Initial development release with core functionality:
- Basic resource operations
- Pod operations
- Multi-cluster support
- STDIO transport
- Helm toolset
- KubeVirt toolset

---

## Version History

- **1.0.0** - First stable release with all roadmap features completed
- **0.1.0** - Initial development release

---

## Release Notes

For detailed release notes, see [RELEASE_NOTES.md](./RELEASE_NOTES.md).

