# Configuration Guide

## Overview

kube-mcp uses TOML format configuration files with support for:

> **New to kube-mcp?** Start with the [Getting Started Guide](GETTING_STARTED.md) for installation and basic setup.
> 
> **Need help with multi-cluster?** See the [Multi-Cluster Guide](MULTI_CLUSTER.md) for detailed multi-cluster configuration and usage.
- Base configuration file
- Drop-in directory (`conf.d/*.toml`)
- Deep merging of configurations
- Hot reload on SIGHUP (not available on Windows)

## Configuration File Structure

```toml
[server]
transports = ["stdio"]
log_level = "info"

[server.http]
address = "0.0.0.0:8080"

[server.http.oauth]
enabled = false
provider = "oidc"
issuer_url = ""
client_id = ""
client_secret = ""

[server.http.cors]
enabled = false
allowed_origins = ["*"]
allowed_methods = ["GET", "POST", "OPTIONS"]
allowed_headers = ["Content-Type", "Authorization"]


[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"
context = ""
default_namespace = ""
qps = 100
burst = 200
timeout = "30s"

[security]
read_only = false
non_destructive = false
denied_gvks = []
require_rbac = true

[helm]
storage_driver = "secret"
default_namespace = "default"

[kubevirt]
enabled = true
vm_group_version = "kubevirt.io/v1"

[kiali]
enabled = false
url = ""
token = ""
timeout = "30s"

[kiali.tls]
enabled = false
ca_file = ""
cert_file = ""
key_file = ""
insecure_skip_verify = false

[toolsets.gitops]
enabled = true

[toolsets.policy]
enabled = true

[toolsets.capi]
enabled = true

[toolsets.rollouts]
enabled = true

[toolsets.certs]
enabled = true

[toolsets.autoscaling]
enabled = true

[toolsets.backup]
enabled = true

[toolsets.net]
enabled = true
hubble_api_url = ""
hubble_insecure = false
hubble_ca_file = ""
hubble_timeout = "10s"
```

## Configuration Sections

### `[server]`
Server-level configuration:
- `transports`: List of enabled transports (`stdio`, `http`)
- `log_level`: Logging level (`debug`, `info`, `warn`, `error`)

### `[server.http]`
HTTP transport configuration (uses Streamable HTTP):
- `address`: Bind address for the server
- `oauth`: OAuth2/OIDC configuration
- `cors`: CORS configuration

### `[kubernetes]`
Kubernetes client configuration:
- `provider`: Provider type (`kubeconfig`, `in-cluster`, `single`)
- `kubeconfig_path`: Path to kubeconfig file
- `context`: Context name (for single-cluster mode)
- `qps`: Queries per second limit
- `burst`: Burst limit
- `timeout`: Request timeout

### `[security]`
Security settings:
- `read_only`: Enable read-only mode
- `non_destructive`: Enable non-destructive mode
- `denied_gvks`: List of denied GroupVersionKinds
- `require_rbac`: Require RBAC checks

### `[helm]`
Helm configuration:
- `storage_driver`: Storage driver (`secret`, `configmap`, `memory`)
- `default_namespace`: Default namespace for Helm operations

### `[kubevirt]`
KubeVirt configuration:
- `enabled`: Enable KubeVirt toolset (auto-detected if CRDs exist)
- `vm_group_version`: VirtualMachine CRD group version

### `[kiali]`
Kiali configuration:
- `enabled`: Enable Kiali integration
- `url`: Kiali server URL
- `token`: Authentication token
- `timeout`: Request timeout
- `tls`: TLS configuration

### `[toolsets.gitops]`
GitOps toolset configuration:
- `enabled`: Enable GitOps toolset (auto-detected if CRDs exist)

### `[toolsets.policy]`
Policy toolset configuration:
- `enabled`: Enable Policy toolset (auto-detected if CRDs exist)

### `[toolsets.capi]`
CAPI toolset configuration:
- `enabled`: Enable CAPI toolset (auto-detected if CRDs exist)

### `[toolsets.rollouts]`
Progressive Delivery toolset configuration:
- `enabled`: Enable Rollouts toolset (auto-detected if CRDs exist)

### `[toolsets.certs]`
Cert-Manager toolset configuration:
- `enabled`: Enable Certs toolset (auto-detected if CRDs exist)

### `[toolsets.autoscaling]`
Autoscaling toolset configuration:
- `enabled`: Enable Autoscaling toolset (HPA always available, KEDA auto-detected)

### `[toolsets.backup]`
Backup/Restore toolset configuration:
- `enabled`: Enable Backup toolset (auto-detected if CRDs exist)

### `[toolsets.net]`
Network toolset configuration:
- `enabled`: Enable Network toolset (NetworkPolicy always available, Cilium auto-detected)
- `hubble_api_url`: Hubble API URL (optional, for flow queries)
- `hubble_insecure`: Skip TLS verification for Hubble (default: false)
- `hubble_ca_file`: Path to CA certificate file for Hubble TLS
- `hubble_timeout`: Request timeout for Hubble API (default: "10s")

## Example Configurations

### Minimal Configuration (STDIO only, Local Dev)
```toml
[server]
transports = ["stdio"]
log_level = "debug"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"

[security]
require_rbac = false  # Disable RBAC for local development
```

### Multi-Cluster via Kubeconfig
```toml
[server]
transports = ["stdio", "http"]
log_level = "info"

[server.http]
address = "0.0.0.0:8080"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"
# context is empty - uses default context or can be specified per tool call
qps = 100
burst = 200
timeout = "30s"

[security]
require_rbac = true
rbac_cache_ttl = 5
```

### HTTP Transport with OAuth + TokenReview
```toml
[server]
transports = ["http"]
log_level = "info"

[server.http]
address = "0.0.0.0:8080"

[server.http.oauth]
enabled = true
provider = "oidc"
issuer_url = "https://auth.example.com"
client_id = "kube-mcp-client"
client_secret = "${OAUTH_CLIENT_SECRET}"  # Can use env var
scopes = ["openid", "profile", "email"]
redirect_url = "http://localhost:8080/oauth/callback"

[server.http.cors]
enabled = true
allowed_origins = ["https://app.example.com"]
allowed_methods = ["GET", "POST", "OPTIONS"]
allowed_headers = ["Content-Type", "Authorization"]

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"

[security]
require_rbac = true
validate_token = true  # Validate Bearer tokens using TokenReview
rbac_cache_ttl = 5
```

### Kiali-Enabled Configuration
```toml
[server]
transports = ["stdio"]
log_level = "info"

[kubernetes]
provider = "kubeconfig"

[kiali]
enabled = true
url = "https://kiali.example.com"
token = "${KIALI_TOKEN}"  # Can use env var
timeout = "30s"

[kiali.tls]
enabled = true
ca_file = "/etc/kube-mcp/certs/kiali-ca.crt"
cert_file = "/etc/kube-mcp/certs/kiali-client.crt"
key_file = "/etc/kube-mcp/certs/kiali-client.key"
insecure_skip_verify = false
```

### KubeVirt-Enabled Configuration
```toml
[server]
transports = ["stdio"]
log_level = "info"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"

[kubevirt]
enabled = true
# KubeVirt toolset will auto-detect CRDs and enable tools if available
```

### In-Cluster Configuration (Running as Pod)
```toml
[server]
transports = ["stdio"]
log_level = "info"

[kubernetes]
provider = "in-cluster"
# Uses service account credentials automatically
qps = 50
burst = 100
timeout = "30s"

[security]
require_rbac = true
rbac_cache_ttl = 10
```

### Single-Cluster Mode (Fixed Context)
```toml
[server]
transports = ["stdio"]

[kubernetes]
provider = "single"
kubeconfig_path = "~/.kube/config"
context = "production-cluster"  # Fixed context
```

### Production Configuration with All Features
```toml
[server]
transports = ["http"]
log_level = "info"

[server.http]
address = "0.0.0.0:8080"

[server.http.oauth]
enabled = true
provider = "oidc"
issuer_url = "https://auth.example.com"
client_id = "kube-mcp"
client_secret = "${OAUTH_CLIENT_SECRET}"

[server.http.cors]
enabled = true
allowed_origins = ["https://app.example.com"]
issuer_url = "https://auth.example.com"
client_id = "kube-mcp"
client_secret = "${OAUTH_CLIENT_SECRET}"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "/etc/kube-mcp/kubeconfig"
qps = 100
burst = 200
timeout = "30s"

[security]
read_only = false
non_destructive = false
require_rbac = true
rbac_cache_ttl = 5
validate_token = true

[helm]
storage_driver = "secret"
default_namespace = "default"

[kubevirt]
enabled = true

[kiali]
enabled = true
url = "https://kiali.example.com"
token = "${KIALI_TOKEN}"
timeout = "30s"

[kiali.tls]
enabled = true
ca_file = "/etc/kube-mcp/certs/kiali-ca.crt"
insecure_skip_verify = false

[toolsets.gitops]
enabled = true

[toolsets.policy]
enabled = true

[toolsets.capi]
enabled = true

[toolsets.rollouts]
enabled = true

[toolsets.certs]
enabled = true

[toolsets.autoscaling]
enabled = true

[toolsets.backup]
enabled = true

[toolsets.net]
enabled = true
hubble_api_url = ""
hubble_insecure = false
hubble_ca_file = ""
hubble_timeout = "10s"
```

### Read-Only Mode (Audit/Inspection Only)
```toml
[server]
transports = ["stdio"]
log_level = "info"

[kubernetes]
provider = "kubeconfig"

[security]
read_only = true
require_rbac = true
rbac_cache_ttl = 5
```

## Drop-In Configuration

Place additional `.toml` files in the `conf.d/` directory. They will be loaded and merged with the base configuration:

```
/etc/kube-mcp/
├── config.toml          # Base configuration
└── conf.d/
    ├── 10-security.toml # Security overrides
    └── 20-kiali.toml    # Kiali configuration
```

## Hot Reload

Send SIGHUP to reload configuration:
```bash
kill -HUP <pid>
```

Note: Hot reload is not available on Windows.

### Runtime-Reloadable Settings

The following settings can be reloaded at runtime without restarting the server:

- **Logging**: `log_level`, `log_format`
- **Toolset enabling**: Toolset enable/disable flags
- **Kiali settings**: `kiali.enabled`, `kiali.url`, `kiali.token`, `kiali.timeout`
- **KubeVirt settings**: `kubevirt.enabled`
- **OAuth settings**: `oauth.validate_token`, `oauth.propagate_token`
- **Rate limiting**: `rate_limit.enabled`, `rate_limit.rps`
- **Metrics**: `metrics.enabled`

### Restart-Required Settings

The following settings require a server restart:

- **Transports**: `server.transports` (stdio/http)
- **Ports and addresses**: `server.http.address`
- **Kubernetes provider**: `kubernetes.provider`, `kubernetes.kubeconfig_path`, `kubernetes.context`
- **TLS certificates**: `server.tls.cert_file`, `server.tls.key_file`
- **OAuth provider URLs**: `server.http.oauth.issuer_url`, `server.http.oauth.client_id`, `server.http.oauth.client_secret`

When reloading configuration, only runtime-reloadable settings are applied. Changes to restart-required settings are ignored until the server is restarted.

