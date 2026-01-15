# Getting Started with kube-mcp

## Quick Start

This guide will help you get kube-mcp up and running in minutes.

## Prerequisites

- Kubernetes cluster access (kubeconfig or in-cluster)
- Go 1.24+ (for building from source)
- Docker (optional, for containerized deployment)

## Installation

### Option 1: Docker (Recommended for Quick Testing)

```bash
# Pull the image
docker pull ghcr.io/wrkode/kube-mcp:1.0.0

# Run with default config
docker run -v ~/.kube/config:/etc/kube-mcp/kubeconfig:ro \
           -p 8080:8080 \
           ghcr.io/wrkode/kube-mcp:1.0.0 \
           --transport http \
           --config /etc/kube-mcp/config.toml
```

### Option 2: Binary Release

```bash
# Download from GitHub Releases
wget https://github.com/wrkode/kube-mcp/releases/download/v1.0.0/kube-mcp-linux-amd64
chmod +x kube-mcp-linux-amd64

# Run
./kube-mcp-linux-amd64 --config config.toml
```

### Option 3: From Source

```bash
# Clone repository
git clone https://github.com/wrkode/kube-mcp.git
cd kube-mcp

# Build
make build

# Run
./dist/kube-mcp-linux-amd64 --config examples/config.toml
```

### Option 4: Helm Chart (Kubernetes Deployment)

```bash
# Add Helm repository
helm repo add kube-mcp https://wrkode.github.io/kube-mcp/charts
helm repo update

# Install
helm install kube-mcp kube-mcp/kube-mcp

# Or install from local chart
helm install kube-mcp ./charts/kube-mcp
```

## Basic Configuration

Create a minimal `config.toml`:

```toml
[server]
transports = ["stdio"]
log_level = "info"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"

[security]
require_rbac = false  # Disable for local development
```

### Configuration File Locations

kube-mcp looks for configuration in this order:
1. `--config` flag path
2. `./config.toml`
3. `~/.config/kube-mcp/config.toml`
4. `/etc/kube-mcp/config.toml`

## Running kube-mcp

### STDIO Mode (Default)

```bash
kube-mcp --config config.toml
```

This starts kube-mcp in STDIO mode, ready to accept MCP protocol messages via stdin/stdout.

### HTTP Mode

```bash
kube-mcp --transport http --config config.toml
```

Access the HTTP endpoint:
```bash
curl http://localhost:8080/mcp
```


## First Steps

### 1. Verify Installation

**List available contexts:**
```json
{
  "tool": "config_contexts_list",
  "params": {}
}
```

**Expected response:**
```json
{
  "contexts": [
    {
      "name": "dev-cluster",
      "cluster": "dev-cluster",
      "user": "dev-user"
    }
  ],
  "current_context": "dev-cluster"
}
```

### 2. List Pods

**List all pods:**
```json
{
  "tool": "pods_list",
  "params": {
    "namespace": ""
  }
}
```

**List pods in default namespace:**
```json
{
  "tool": "pods_list",
  "params": {
    "namespace": "default"
  }
}
```

### 3. Get Pod Details

```json
{
  "tool": "pods_get",
  "params": {
    "name": "nginx-pod-abc123",
    "namespace": "default"
  }
}
```

### 4. List Namespaces

```json
{
  "tool": "namespaces_list",
  "params": {}
}
```

## Common Use Cases

### Use Case 1: Local Development

**Configuration:**
```toml
[server]
transports = ["stdio"]
log_level = "debug"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"

[security]
require_rbac = false  # Disable RBAC for local dev
```

**Usage:**
- Connect your IDE or MCP client
- Use STDIO transport for direct integration
- Debug with verbose logging

### Use Case 2: HTTP API Server

**Configuration:**
```toml
[server]
transports = ["http"]
log_level = "info"

[server.http]
address = "0.0.0.0:8080"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"

[security]
require_rbac = true
```

**Usage:**
```bash
# Start server
kube-mcp --transport http --config config.toml

# Make API calls
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "pods_list",
    "params": {"namespace": "default"}
  }'
```

### Use Case 3: In-Cluster Deployment

**Configuration:**
```toml
[server]
transports = ["stdio"]
log_level = "info"

[kubernetes]
provider = "in-cluster"
qps = 50
burst = 100

[security]
require_rbac = true
rbac_cache_ttl = 10
```

**Deployment:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-mcp
spec:
  replicas: 1
  template:
    spec:
      serviceAccountName: kube-mcp
      containers:
      - name: kube-mcp
        image: ghcr.io/wrkode/kube-mcp:1.0.0
        args: ["--config", "/etc/kube-mcp/config.toml"]
        volumeMounts:
        - name: config
          mountPath: /etc/kube-mcp/config.toml
          subPath: config.toml
      volumes:
      - name: config
        configMap:
          name: kube-mcp-config
```

### Use Case 4: Multi-Cluster Management

**Configuration:**
```toml
[server]
transports = ["stdio", "http"]
log_level = "info"

[server.http]
address = "0.0.0.0:8080"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"
# Empty context allows per-call switching

[security]
require_rbac = true
rbac_cache_ttl = 5
```

**Usage:**
```json
// List pods in dev cluster
{
  "tool": "pods_list",
  "params": {
    "context": "dev-cluster",
    "namespace": "default"
  }
}

// List pods in prod cluster
{
  "tool": "pods_list",
  "params": {
    "context": "prod-cluster",
    "namespace": "production"
  }
}
```

## Testing Your Setup

### Test 1: Connectivity

```bash
# HTTP mode
curl http://localhost:8080/health

# Expected: {"status":"ok"}
```

### Test 2: List Tools

```json
{
  "tool": "config_contexts_list",
  "params": {}
}
```

### Test 3: Basic Operations

```json
// List namespaces
{
  "tool": "namespaces_list",
  "params": {}
}

// List pods
{
  "tool": "pods_list",
  "params": {
    "namespace": "default"
  }
}
```

## Next Steps

1. **Connect Your Client**: [CLIENT_CONFIGURATION.md](CLIENT_CONFIGURATION.md) - Connect to Cursor, Claude Code, and other MCP clients
2. **Read the Configuration Guide**: [CONFIGURATION.md](CONFIGURATION.md)
3. **Explore Tools**: [TOOLS.md](TOOLS.md)
4. **Learn Multi-Cluster**: [MULTI_CLUSTER.md](MULTI_CLUSTER.md)
5. **Usage Patterns**: [USAGE_GUIDE.md](USAGE_GUIDE.md)
6. **Security Setup**: [SECURITY.md](SECURITY.md)

## Troubleshooting

### Issue: Cannot connect to cluster

**Check:**
1. Kubeconfig path is correct
2. Kubeconfig is readable
3. Cluster API server is accessible
4. Credentials are valid

**Solution:**
```bash
# Test kubeconfig
kubectl cluster-info

# Verify path in config.toml
kubeconfig_path = "~/.kube/config"  # or absolute path
```

### Issue: Permission denied

**Check:**
1. RBAC permissions
2. Service account permissions (in-cluster)
3. Security configuration

**Solution:**
```toml
[security]
require_rbac = false  # Disable for testing
```

### Issue: Port already in use

**Check:**
1. Port 8080 (HTTP) is available
2. No other service using the port

**Solution:**
```toml
[server.http]
address = "0.0.0.0:9090"  # Use different port
```

### Issue: Context not found

**Check:**
1. Context exists in kubeconfig
2. Context name is spelled correctly
3. Kubeconfig is loaded correctly

**Solution:**
```bash
# List contexts
kubectl config get-contexts

# Verify in kube-mcp
{
  "tool": "config_contexts_list",
  "params": {}
}
```

## Examples

See the [examples/](examples/) directory for:
- Sample configuration files
- Example tool calls
- Integration patterns

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/wrkode/kube-mcp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/wrkode/kube-mcp/discussions)

## Summary

You should now have kube-mcp running and ready to use. Key points:

1. **Installation**: Choose Docker, binary, source, or Helm
2. **Configuration**: Create a minimal config.toml
3. **Running**: Start with STDIO or HTTP transport
4. **Testing**: Verify with basic tool calls
5. **Next Steps**: Explore documentation and examples

Happy Kubernetes management with kube-mcp!

