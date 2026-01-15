# Developing New Toolsets

This guide explains how to create a new toolset for kube-mcp, following the established patterns and best practices.

## Overview

A toolset is a collection of related MCP tools that share common functionality. kube-mcp currently includes **13 toolsets**:

**Always Available:**
- **Config toolset**: Kubeconfig inspection and context management
- **Core toolset**: Kubernetes operations (pods, resources, namespaces, etc.)
- **Helm toolset**: Helm chart management

**CRD-Gated (Auto-enabled when CRDs detected):**
- **GitOps toolset**: Flux and Argo CD application management
- **Policy toolset**: Kyverno and Gatekeeper policy visibility
- **CAPI toolset**: Cluster API cluster lifecycle management
- **Rollouts toolset**: Progressive delivery management (Argo Rollouts, Flagger)
- **Certs toolset**: Cert-Manager certificate lifecycle management
- **Autoscaling toolset**: HPA and KEDA autoscaling management
- **Backup toolset**: Velero backup and restore operations
- **Network toolset**: NetworkPolicy, Cilium, and Hubble observability
- **KubeVirt toolset**: Virtual machine lifecycle management

**Configuration-Gated:**
- **Kiali toolset**: Service mesh observability

## Step-by-Step Guide

### 1. Create Directory Structure

Create a new directory under `pkg/toolsets/`:

```bash
mkdir -p pkg/toolsets/mynewtoolset
```

### 2. Define the Toolset Struct

Create `pkg/toolsets/mynewtoolset/toolset.go`:

```go
package mynewtoolset

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
)

// Toolset implements your new toolset.
type Toolset struct {
	provider kubernetes.ClientProvider
	logger   *observability.Logger
	metrics  *observability.Metrics
	// Add your toolset-specific fields here
}

// NewToolset creates a new toolset instance.
func NewToolset(provider kubernetes.ClientProvider) *Toolset {
	return &Toolset{
		provider: provider,
	}
}

// SetObservability sets the observability components.
func (t *Toolset) SetObservability(logger *observability.Logger, metrics *observability.Metrics) {
	t.logger = logger
	t.metrics = metrics
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "mynewtoolset"
}

// Tools returns all tools in this toolset.
func (t *Toolset) Tools() []*mcp.Tool {
	return []*mcp.Tool{
		mcpHelpers.NewTool("mynewtoolset_operation", "Description of operation").
			WithParameter("param1", "string", "Parameter description", true).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().  // or WithDestructive() if it modifies state
			Build(),
		// Add more tools...
	}
}
```

### 3. Implement Tool Handlers

Create handler functions for each tool. Use `ClusterTarget` for cluster-aware operations:

```go
// handleOperation handles the mynewtoolset_operation tool.
func (t *Toolset) handleOperation(ctx context.Context, args struct {
	Param1  string `json:"param1"`
	Context string `json:"context"`  // ClusterTarget embedded
}) (*mcp.CallToolResult, error) {
	// Get client set for the specified context
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Perform your operation
	// ...

	// Return result
	return mcpHelpers.NewJSONResult(map[string]any{
		"result": "success",
	}), nil
}
```

### 4. Create Observability Helper

Create `pkg/toolsets/mynewtoolset/observability_helper.go`:

```go
package mynewtoolset

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// wrapToolHandler wraps a tool handler with observability.
func (t *Toolset) wrapToolHandler(
	toolName string,
	handler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error),
	getCluster func(args any) string,
) func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
	if t.logger == nil || t.metrics == nil {
		return handler
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		start := time.Now()
		cluster := getCluster(args)
		if cluster == "" {
			cluster = "default"
		}

		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				t.logger.Error(ctx, "Panic in tool handler",
					"tool", toolName,
					"panic", r,
					"cluster", cluster,
				)
			}
		}()

		result, out, err := handler(ctx, req, args)

		// Log and record metrics
		duration := time.Since(start)
		t.logger.LogToolInvocation(ctx, toolName, cluster, duration, err)
		success := err == nil && (result == nil || !result.IsError)
		t.metrics.RecordToolCall(toolName, cluster, success, duration.Seconds())

		return result, out, err
	}
}
```

### 5. Register Tools with Observability

Create `pkg/toolsets/mynewtoolset/register.go`:

```go
package mynewtoolset

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
)

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	var handler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error)
	var wrappedHandler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error)

	// Register mynewtoolset_operation
	type OperationArgs struct {
		Param1  string `json:"param1"`
		Context string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs := args.(OperationArgs)
		result, err := t.handleOperation(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("mynewtoolset_operation", handler, func(args any) string {
		return args.(OperationArgs).Context
	})
	mcp.AddTool(server, &mcp.Tool{
		Name:        "mynewtoolset_operation",
		Description: "Description of operation",
	}, wrappedHandler)

	return nil
}
```

### 6. Add RBAC Checks (Optional but Recommended)

If your toolset performs destructive operations, add RBAC checks:

```go
// In your handler, before performing the operation:
gvr := schema.GroupVersionResource{
	Group:    "your.group",
	Version:  "v1",
	Resource: "yourresource",
}

if rbacResult, rbacErr := t.checkRBAC(ctx, clientSet, "create", gvr, namespace); rbacErr != nil || rbacResult != nil {
	if rbacResult != nil {
		return rbacResult, nil
	}
	return mcpHelpers.NewErrorResult(rbacErr), nil
}
```

You'll need to add RBAC authorizer support to your toolset struct (similar to Core toolset).

### 7. Register Toolset in Main Application

Add your toolset to `cmd/kube-mcp/main.go`:

```go
// In registerToolsets function:
mynewToolset := mynewtoolset.NewToolset(provider)
mynewToolset.SetObservability(logger, metrics)
if err := mcpServer.RegisterToolset(mynewToolset); err != nil {
	return fmt.Errorf("failed to register mynewtoolset: %w", err)
}
```

### 8. Add Documentation

Create `docs/tools/mynewtoolset.md` following the pattern of existing toolset documentation:

- Tool descriptions
- Input/output schemas
- Example calls and responses
- Error scenarios
- Feature gating (if applicable)

### 9. Add Tests

Create tests following the existing patterns:

- Unit tests: `pkg/toolsets/mynewtoolset/toolset_test.go`
- Integration tests: `test/integration/mynewtoolset_test.go` (if applicable)

## Best Practices

### 1. Use ClusterTarget Consistently

Always embed `ClusterTarget` (via `context` field) in tool arguments for cluster-aware operations:

```go
type MyArgs struct {
	Param1  string `json:"param1"`
	Context string `json:"context"`  // ClusterTarget
}
```

### 2. Feature Gating

If your toolset depends on external services or CRDs, implement feature gating:

```go
func (t *Toolset) IsEnabled() bool {
	return t.enabled
}

func (t *Toolset) Tools() []*mcp.Tool {
	if !t.enabled {
		return []*mcp.Tool{}
	}
	// Return tools...
}
```

### 3. Error Handling

Always return structured errors using `mcpHelpers.NewErrorResult()`:

```go
if err != nil {
	return mcpHelpers.NewErrorResult(fmt.Errorf("operation failed: %w", err)), nil
}
```

### 4. Observability

Always wrap tool handlers with `wrapToolHandler()` to ensure:
- Logging of tool invocations
- Metrics recording
- Panic recovery

### 5. Read-Only vs Destructive

Mark tools appropriately:
- `WithReadOnly()`: Tool only reads data
- `WithDestructive()`: Tool modifies or deletes resources

### 6. Input Validation

Validate inputs early in handlers:

```go
if args.Param1 == "" {
	return mcpHelpers.NewErrorResult(fmt.Errorf("param1 is required")), nil
}
```

### 7. Context Handling

Always use `context.Context` for cancellation and timeouts:

```go
func (t *Toolset) handleOperation(ctx context.Context, args MyArgs) (*mcp.CallToolResult, error) {
	// Use ctx for all operations
	clientSet.Typed.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}
```

## Example: Complete Toolset Structure

```
pkg/toolsets/mynewtoolset/
├── toolset.go              # Toolset definition and metadata
├── register.go              # Tool registration with observability wrappers
├── handlers.go              # Tool handler implementations
├── observability_helper.go  # Observability wrapper function
└── helpers.go               # Helper functions (optional)
```

## Integration Checklist

- [ ] Toolset struct defined with observability fields
- [ ] `SetObservability()` method implemented
- [ ] `Name()` method returns toolset name
- [ ] `Tools()` method returns tool definitions
- [ ] `RegisterTools()` method registers all tools with wrappers
- [ ] All handlers use `ClusterTarget` for cluster-aware operations
- [ ] Observability wrappers applied to all tools
- [ ] RBAC checks added for destructive operations (if applicable)
- [ ] Toolset registered in `main.go`
- [ ] Documentation created in `docs/tools/mynewtoolset.md`
- [ ] Tests added (unit and/or integration)
- [ ] Code compiles without errors
- [ ] `go vet` passes

## Testing Your Toolset

### Unit Tests

```go
func TestMyToolset(t *testing.T) {
	provider := &mockProvider{}
	toolset := NewToolset(provider)
	
	// Test toolset methods
	assert.Equal(t, "mynewtoolset", toolset.Name())
	assert.NotNil(t, toolset.Tools())
}
```

### Integration Tests

Use the envtest suite for integration tests:

```go
func TestMyToolsetIntegration(t *testing.T) {
	suite.Run(t, &MyToolsetTestSuite{EnvtestSuite: EnvtestSuite{}})
}
```

## Common Patterns

### Pattern: Conditional Toolset (like KubeVirt)

```go
type Toolset struct {
	enabled bool
	// ...
}

func NewToolset(provider kubernetes.ClientProvider, discovery *kubernetes.CRDDiscovery) *Toolset {
	enabled := checkIfCRDsExist(discovery)
	return &Toolset{
		enabled: enabled,
		// ...
	}
}
```

### Pattern: External Service Integration (like Kiali)

```go
type Toolset struct {
	client  *ExternalClient
	enabled bool
	// ...
}

func NewToolset(cfg *config.ExternalConfig) (*Toolset, error) {
	if !cfg.Enabled {
		return &Toolset{enabled: false}, nil
	}
	
	client, err := NewExternalClient(cfg)
	if err != nil {
		return nil, err
	}
	
	return &Toolset{
		client:  client,
		enabled: true,
	}, nil
}
```

## Next Steps

1. Review existing toolsets (`core`, `helm`) for reference
2. Follow the patterns established in this guide
3. Test thoroughly before submitting
4. Update `docs/TOOLS.md` to include your new toolset
5. Add examples to documentation

## Questions?

Refer to:
- Existing toolset implementations in `pkg/toolsets/`
- MCP SDK documentation: https://github.com/modelcontextprotocol/go-sdk
- Kubernetes client-go documentation: https://github.com/kubernetes/client-go

