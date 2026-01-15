package config

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
)

// Toolset implements the Config toolset for kubeconfig operations.
type Toolset struct {
	provider kubernetes.ClientProvider
}

// NewToolset creates a new Config toolset.
func NewToolset(provider kubernetes.ClientProvider) *Toolset {
	return &Toolset{
		provider: provider,
	}
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "config"
}

// Tools returns all tools in this toolset.
func (t *Toolset) Tools() []*mcp.Tool {
	return []*mcp.Tool{
		mcpHelpers.NewTool("config_contexts_list", "List all available Kubernetes contexts from kubeconfig").
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("config_kubeconfig_view", "View kubeconfig file contents (full or minified)").
			WithParameter("minified", "boolean", "If true, return minified version", false).
			WithReadOnly().
			Build(),
	}
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	// Register config_contexts_list
	type ContextsListArgs struct{}
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "config_contexts_list",
		Description: "List all available Kubernetes contexts from kubeconfig",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ContextsListArgs) (*mcp.CallToolResult, any, error) {
		contexts, err := t.provider.ListContexts()
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list contexts: %w", err)), nil, nil
		}

		current, err := t.provider.GetCurrentContext()
		if err != nil {
			current = ""
		}

		result := map[string]any{
			"contexts":        contexts,
			"current_context": current,
		}

		res, err := mcpHelpers.NewJSONResult(result)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return res, nil, nil
	})

	// Register config_kubeconfig_view
	type KubeconfigViewArgs struct {
		Minified bool `json:"minified"`
	}
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "config_kubeconfig_view",
		Description: "View kubeconfig file contents (full or minified)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args KubeconfigViewArgs) (*mcp.CallToolResult, any, error) {
		result := map[string]any{
			"message":  "Kubeconfig viewing not yet implemented - requires kubeconfig path access",
			"minified": args.Minified,
		}

		res, err := mcpHelpers.NewJSONResult(result)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return res, nil, nil
	})

	return nil
}
