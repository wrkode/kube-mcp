# Client Configuration Guide

This guide explains how to connect kube-mcp to various MCP clients including Cursor, Claude Code, and other applications.

## Overview

kube-mcp supports the Model Context Protocol (MCP) and can be used with any MCP-compatible client. The server supports two transport modes:

- **STDIO** (default): Uses stdin/stdout for communication - recommended for IDE integrations
- **HTTP**: Uses Streamable HTTP for web-based clients and API integrations

## Prerequisites

1. **kube-mcp installed** - See [GETTING_STARTED.md](GETTING_STARTED.md) for installation instructions
2. **Kubernetes access** - Valid kubeconfig file or in-cluster access
3. **MCP-compatible client** - Cursor, Claude Code, or other MCP clients

## Configuration File

Create a `config.toml` file for kube-mcp:

```toml
[server]
# Use stdio transport for IDE clients like Cursor
transports = ["stdio"]
log_level = "info"

# Optional: Normalize tool names for clients that don't support dots
# When enabled: "autoscaling.hpa_explain" becomes "autoscaling_hpa_explain"
# Set to true if using n8n or other clients with naming restrictions
normalize_tool_names = false

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"
# Leave context empty to allow per-call context switching
context = ""

[security]
require_rbac = true
read_only = false
non_destructive = false
```

## Cursor IDE Configuration

Cursor is a popular IDE that supports MCP servers. To connect kube-mcp:

### Step 1: Install kube-mcp

```bash
# Download binary from GitHub Releases
wget https://github.com/wrkode/kube-mcp/releases/download/v1.0.0/kube-mcp-linux-amd64
chmod +x kube-mcp-linux-amd64
sudo mv kube-mcp-linux-amd64 /usr/local/bin/kube-mcp
```

### Step 2: Create Configuration File

Create `~/.config/kube-mcp/config.toml`:

```toml
[server]
transports = ["stdio"]
log_level = "info"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"

[security]
require_rbac = true
```

### Step 3: Configure Cursor

1. Open Cursor Settings
2. Navigate to **Features** → **Model Context Protocol**
3. Add a new MCP server:

**Server Configuration:**
```json
{
  "name": "kube-mcp",
  "command": "kube-mcp",
  "args": ["--config", "~/.config/kube-mcp/config.toml"]
}
```

**Alternative (with full path):**
```json
{
  "name": "kube-mcp",
  "command": "/usr/local/bin/kube-mcp",
  "args": ["--config", "/home/username/.config/kube-mcp/config.toml"]
}
```

### Step 4: Verify Connection

1. Restart Cursor
2. Open the MCP panel or check the status indicator
3. You should see kube-mcp listed with available tools

### Step 5: Test in Cursor

Ask Cursor to use kube-mcp tools:

```
List all pods in the default namespace using kube-mcp
```

Cursor should automatically use the `pods_list` tool from kube-mcp.

## Claude Code Configuration

Claude Code (Anthropic's IDE) also supports MCP servers:

### Step 1: Install kube-mcp

Same as Cursor - download and install the binary.

### Step 2: Configure Claude Code

1. Open Claude Code Settings
2. Go to **MCP Servers** section
3. Add configuration:

```json
{
  "mcpServers": {
    "kube-mcp": {
      "command": "kube-mcp",
      "args": ["--config", "~/.config/kube-mcp/config.toml"]
    }
  }
}
```

### Step 3: Restart and Test

Restart Claude Code and verify the connection. You can then ask Claude to manage Kubernetes resources using kube-mcp tools.

## Other MCP Clients

### Generic MCP Client Configuration

For any MCP-compatible client, use this configuration pattern:

**STDIO Transport (Recommended for IDEs):**
```json
{
  "name": "kube-mcp",
  "command": "kube-mcp",
  "args": ["--config", "/path/to/config.toml"],
  "env": {}
}
```

**HTTP Transport (For web clients or APIs):**

1. Start kube-mcp in HTTP mode:
```bash
kube-mcp --transport http --config config.toml
```

2. Configure client to connect to:
```
http://localhost:8080/mcp
```

### n8n Configuration

n8n requires normalized tool names. Configure kube-mcp:

```toml
[server]
transports = ["http"]
normalize_tool_names = true  # Important for n8n!

[server.http]
address = "0.0.0.0:8080"
```

Then configure n8n to connect to `http://localhost:8080/mcp`.

## Configuration Examples

### Example 1: Local Development with Cursor

```toml
[server]
transports = ["stdio"]
log_level = "debug"  # Verbose logging for debugging

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"

[security]
require_rbac = false  # Disable RBAC checks for local dev
```

### Example 2: Production with Multiple Clusters

```toml
[server]
transports = ["stdio"]
log_level = "info"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "/etc/kube-mcp/kubeconfig"
# Empty context allows per-call context switching
context = ""

[security]
require_rbac = true
rbac_cache_ttl = 5
```

### Example 3: HTTP API Server

```toml
[server]
transports = ["http"]
log_level = "info"

[server.http]
address = "0.0.0.0:8080"

[server.http.oauth]
enabled = true
provider = "oidc"
issuer_url = "https://your-oidc-provider.com"
client_id = "kube-mcp-client"
client_secret = "your-secret"

[kubernetes]
provider = "kubeconfig"
kubeconfig_path = "~/.kube/config"

[security]
require_rbac = true
```

## Troubleshooting

### Issue: Client cannot connect to kube-mcp

**Check:**
1. kube-mcp binary is in PATH or use full path
2. Configuration file path is correct
3. kube-mcp has execute permissions
4. Configuration file is valid TOML

**Solution:**
```bash
# Test kube-mcp directly
kube-mcp --config config.toml

# Check if it starts without errors
# Should wait for stdin input (STDIO mode)
```

### Issue: Tools not appearing in client

**Check:**
1. Client supports MCP protocol
2. Connection is established (check client logs)
3. kube-mcp is running and responding

**Solution:**
```bash
# Test with a simple MCP client
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | kube-mcp --config config.toml
```

### Issue: Permission denied errors

**Check:**
1. Kubernetes credentials are valid
2. RBAC permissions are correct
3. Security configuration allows operations

**Solution:**
```toml
[security]
require_rbac = false  # Temporarily disable for testing
read_only = true       # Enable read-only mode
```

### Issue: Context not found

**Check:**
1. Context exists in kubeconfig
2. Context name is spelled correctly
3. kubeconfig is accessible

**Solution:**
```bash
# List available contexts
kubectl config get-contexts

# Verify in kube-mcp
# Use config_contexts_list tool in your client
```

## Advanced Configuration

### Multi-Transport Setup

You can enable both STDIO and HTTP transports:

```toml
[server]
transports = ["stdio", "http"]

[server.http]
address = "0.0.0.0:8080"
```

This allows:
- IDE clients to use STDIO
- Web clients/APIs to use HTTP

### Custom Tool Name Normalization

If your client has specific naming requirements, enable normalization:

```toml
[server]
normalize_tool_names = true
```

This converts tool names like:
- `autoscaling.hpa_explain` → `autoscaling_hpa_explain`
- `net.networkpolicies_list` → `net_networkpolicies_list`

### Environment Variables

You can override configuration with environment variables:

```bash
export KUBE_MCP_CONFIG=/path/to/config.toml
export KUBE_MCP_TRANSPORT=stdio
kube-mcp
```

## Security Considerations

### For Local Development

```toml
[security]
require_rbac = false
read_only = false
```

### For Production

```toml
[security]
require_rbac = true
read_only = false
non_destructive = false
rbac_cache_ttl = 5
```

### For Read-Only Access

```toml
[security]
require_rbac = true
read_only = true
```

## Testing Your Configuration

### Test 1: Verify kube-mcp starts

```bash
kube-mcp --config config.toml
# Should start and wait for input (STDIO mode)
```

### Test 2: List available tools

Use your MCP client to call:
```json
{
  "method": "tools/list"
}
```

You should see all kube-mcp tools listed.

### Test 3: Test a simple tool

```json
{
  "method": "tools/call",
  "params": {
    "name": "config_contexts_list",
    "arguments": {}
  }
}
```

## Next Steps

1. **Explore Tools**: See [TOOLS.md](TOOLS.md) for all available tools
2. **Usage Examples**: Check [USAGE_GUIDE.md](USAGE_GUIDE.md) for usage patterns
3. **Multi-Cluster**: Learn about multi-cluster management in [MULTI_CLUSTER.md](MULTI_CLUSTER.md)
4. **Security**: Review [SECURITY.md](SECURITY.md) for security best practices

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/wrkode/kube-mcp/issues)
- **Examples**: [examples/](examples/)

## Summary

Connecting kube-mcp to your MCP client:

1. **Install** kube-mcp binary
2. **Create** configuration file (`config.toml`)
3. **Configure** your client with kube-mcp command
4. **Test** the connection
5. **Start using** Kubernetes management tools!

For IDE clients like Cursor, use STDIO transport. For web clients or APIs, use HTTP transport.
