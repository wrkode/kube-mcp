package gitops

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
)

// Tools returns all tools in this toolset (only if enabled).
func (t *Toolset) Tools() []*mcp.Tool {
	if !t.enabled {
		return []*mcp.Tool{}
	}

	tools := []*mcp.Tool{}

	// gitops.apps_list
	tools = append(tools, mcpHelpers.NewTool("gitops.apps_list", "List GitOps applications (Flux Kustomization/HelmRelease or Argo CD Application)").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
		WithParameter("label_selector", "string", "Label selector", false).
		WithParameter("kinds", "array", "Array of kinds to filter: 'Kustomization', 'HelmRelease', 'Application' (default: all available)", false).
		WithParameter("limit", "integer", "Maximum number of items to return", false).
		WithParameter("continue", "string", "Token from previous paginated request", false).
		WithReadOnly().
		Build())

	// gitops.app_get
	tools = append(tools, mcpHelpers.NewTool("gitops.app_get", "Get GitOps application details").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("kind", "string", "Application kind: 'Kustomization', 'HelmRelease', or 'Application'", true).
		WithParameter("name", "string", "Application name", true).
		WithParameter("namespace", "string", "Namespace name (required for namespaced kinds)", true).
		WithParameter("raw", "boolean", "Return raw object if true", false).
		WithReadOnly().
		Build())

	// gitops.app_reconcile (only for Flux)
	tools = append(tools, mcpHelpers.NewTool("gitops.app_reconcile", "Trigger reconciliation for a Flux Kustomization or HelmRelease").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("kind", "string", "Application kind: 'Kustomization' or 'HelmRelease'", true).
		WithParameter("name", "string", "Application name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("confirm", "boolean", "Must be true to reconcile", true).
		WithDestructive().
		Build())

	return tools
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	if !t.enabled {
		return nil
	}

	// Register gitops.apps_list
	type AppsListArgs struct {
		Context       string   `json:"context"`
		Namespace     string   `json:"namespace"`
		LabelSelector string   `json:"label_selector"`
		Kinds         []string `json:"kinds"`
		Limit         int      `json:"limit"`
		Continue      string   `json:"continue"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[AppsListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleAppsList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("gitops.apps_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[AppsListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "gitops.apps_list",
		Description: "List GitOps applications",
	}, wrappedHandler)

	// Register gitops.app_get
	type AppGetArgs struct {
		Context   string `json:"context"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Raw       bool   `json:"raw"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[AppGetArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleAppGet(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("gitops.app_get", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[AppGetArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "gitops.app_get",
		Description: "Get GitOps application details",
	}, wrappedHandler)

	// Register gitops.app_reconcile
	type AppReconcileArgs struct {
		Context   string `json:"context"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Confirm   bool   `json:"confirm"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[AppReconcileArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleAppReconcile(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("gitops.app_reconcile", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[AppReconcileArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "gitops.app_reconcile",
		Description: "Trigger reconciliation for a Flux Kustomization or HelmRelease",
	}, wrappedHandler)

	return nil
}
