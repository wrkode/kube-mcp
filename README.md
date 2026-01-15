# kube-mcp

A production-grade Model Context Protocol (MCP) server for Kubernetes management, written in Go.

## Overview

kube-mcp provides comprehensive Kubernetes cluster management through the Model Context Protocol. It exposes **60+ MCP tools** across **13 toolsets**, supporting multiple transports (STDIO, HTTP) and multiple Kubernetes clusters.

## Features

### Core Capabilities
- **Multi-cluster support** - Manage multiple Kubernetes clusters from a single server instance
- **Multiple transports** - STDIO (default) and HTTP (Streamable HTTP) with OAuth2/OIDC support
- **Dynamic CRD discovery** - Automatically detects and supports Custom Resource Definitions
- **Server-side apply** - Safe resource reconciliation using Kubernetes field managers
- **RBAC integration** - Comprehensive permission checks with configurable caching
- **Observability** - Structured logging, Prometheus metrics, and health endpoints

### Kubernetes Operations (Core Toolset)
- **Resource Management**: List, get, apply, delete, patch, diff, validate, watch, scale, describe, relationships
- **Pod Operations**: List, get, delete, logs (with follow), exec, port-forward, metrics
- **ConfigMap & Secrets**: Get and set data with merge/replace options
- **Namespaces & Nodes**: List namespaces, node metrics and summaries
- **Events**: List and filter Kubernetes events

### Additional Toolsets
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

### Advanced Features
- **Dry-run support** - Validate operations without making changes
- **Resource patching** - Merge, JSON, and strategic merge patches
- **Resource diffing** - Compare current vs desired state (unified/JSON/YAML)
- **Resource watching** - Real-time monitoring with event collection
- **Label/field selectors** - Efficient resource filtering
- **Pagination** - Handle large result sets efficiently
- **Log streaming** - Follow pod logs in real-time

## Installation

### Helm Chart (Recommended for Kubernetes)

```bash
# From Helm repository (requires gh-pages branch setup)
# See docs/CHARTS_REPOSITORY_SETUP.md for setup instructions
helm repo add kube-mcp https://wrkode.github.io/kube-mcp
helm repo update
helm install kube-mcp kube-mcp/kube-mcp

# Or from local chart
helm install kube-mcp ./charts/kube-mcp
```

### Docker

```bash
docker pull ghcr.io/wrkode/kube-mcp:1.0.0
docker run -v /path/to/config.toml:/etc/kube-mcp/config.toml:ro \
           -p 8080:8080 \
           ghcr.io/wrkode/kube-mcp:1.0.0
```

### From Source

```bash
go install github.com/wrkode/kube-mcp/cmd/kube-mcp@latest
```

### Binary Releases

Download pre-built binaries from [GitHub Releases](https://github.com/wrkode/kube-mcp/releases).

## Quick Start

### STDIO Mode (Default)

```bash
kube-mcp --config /path/to/config.toml
```

### HTTP Mode

```bash
kube-mcp --transport http --config /path/to/config.toml
```

See [examples/config.toml](examples/config.toml) for a sample configuration file.

## Documentation

### Client Configuration
- **[Client Configuration Guide](docs/CLIENT_CONFIGURATION.md)** - Connect kube-mcp to Cursor, Claude Code, and other MCP clients

### Getting Started
- **[Getting Started](docs/GETTING_STARTED.md)** - Quick start guide and installation
- **[Usage Guide](docs/USAGE_GUIDE.md)** - Practical examples and workflows
- **[Best Practices](docs/BEST_PRACTICES.md)** - Production recommendations and patterns

### Core Documentation
- **[Architecture](docs/ARCHITECTURE.md)** - System architecture overview
- **[Configuration](docs/CONFIGURATION.md)** - Complete configuration guide
- **[Multi-Cluster Guide](docs/MULTI_CLUSTER.md)** - Multi-cluster setup and usage
- **[Tools](docs/TOOLS.md)** - Full tool reference with examples
- **[Security](docs/SECURITY.md)** - Security and RBAC documentation

### Deployment
- **[Helm Chart](charts/kube-mcp/README.md)** - Helm chart documentation
- **[Field Testing Guides](field_testing/)** - Feature guides and quick reference

## Security

- **RBAC checks** - All destructive operations verify permissions before execution
- **RBAC caching** - Configurable TTL-based caching for performance
- **Token validation** - Bearer token validation via Kubernetes TokenReview API
- **Read-only mode** - Optional read-only mode for restricted deployments
- **Distroless container** - Minimal attack surface with non-root user

## Status

**Version**: 1.0.0 (Stable)  
**License**: Apache License 2.0

## Contributing

Contributions are welcome! Please see our [development guide](docs/developing-new-toolsets.md) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/wrkode/kube-mcp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/wrkode/kube-mcp/discussions)
