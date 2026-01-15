package rollouts

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

	// rollouts.list
	tools = append(tools, mcpHelpers.NewTool("rollouts.list", "List progressive delivery resources (Argo Rollouts Rollout or Flagger Canary)").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
		WithParameter("label_selector", "string", "Label selector", false).
		WithParameter("limit", "integer", "Maximum number of items to return", false).
		WithParameter("continue", "string", "Token from previous paginated request", false).
		WithReadOnly().
		Build())

	// rollouts.get_status
	tools = append(tools, mcpHelpers.NewTool("rollouts.get_status", "Get detailed status of a progressive delivery resource").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("kind", "string", "Resource kind: 'Rollout' or 'Canary'", true).
		WithParameter("name", "string", "Resource name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("raw", "boolean", "Return raw object if true", false).
		WithReadOnly().
		Build())

	// rollouts.promote
	tools = append(tools, mcpHelpers.NewTool("rollouts.promote", "Promote a rollout to the next step (Argo Rollouts only)").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("kind", "string", "Resource kind: 'Rollout'", true).
		WithParameter("name", "string", "Resource name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("confirm", "boolean", "Must be true to promote", true).
		WithDestructive().
		Build())

	// rollouts.abort
	tools = append(tools, mcpHelpers.NewTool("rollouts.abort", "Abort a rollout (Argo Rollouts only)").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("kind", "string", "Resource kind: 'Rollout'", true).
		WithParameter("name", "string", "Resource name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("confirm", "boolean", "Must be true to abort", true).
		WithDestructive().
		Build())

	// rollouts.retry
	tools = append(tools, mcpHelpers.NewTool("rollouts.retry", "Retry a rollout analysis or progression (Argo Rollouts only)").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("kind", "string", "Resource kind: 'Rollout'", true).
		WithParameter("name", "string", "Resource name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("confirm", "boolean", "Must be true to retry", true).
		WithDestructive().
		Build())

	return tools
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	if !t.enabled {
		return nil
	}

	// Register rollouts.list
	type RolloutsListArgs struct {
		Context       string `json:"context"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[RolloutsListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleRolloutsList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("rollouts.list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[RolloutsListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "rollouts.list",
		Description: "List progressive delivery resources",
	}, wrappedHandler)

	// Register rollouts.get_status
	type RolloutGetStatusArgs struct {
		Context   string `json:"context"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Raw       bool   `json:"raw"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[RolloutGetStatusArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleRolloutGetStatus(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("rollouts.get_status", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[RolloutGetStatusArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "rollouts.get_status",
		Description: "Get detailed status of a progressive delivery resource",
	}, wrappedHandler)

	// Register rollouts.promote
	type RolloutPromoteArgs struct {
		Context   string `json:"context"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Confirm   bool   `json:"confirm"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[RolloutPromoteArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleRolloutPromote(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("rollouts.promote", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[RolloutPromoteArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "rollouts.promote",
		Description: "Promote a rollout to the next step",
	}, wrappedHandler)

	// Register rollouts.abort
	type RolloutAbortArgs struct {
		Context   string `json:"context"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Confirm   bool   `json:"confirm"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[RolloutAbortArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleRolloutAbort(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("rollouts.abort", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[RolloutAbortArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "rollouts.abort",
		Description: "Abort a rollout",
	}, wrappedHandler)

	// Register rollouts.retry
	type RolloutRetryArgs struct {
		Context   string `json:"context"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Confirm   bool   `json:"confirm"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[RolloutRetryArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleRolloutRetry(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("rollouts.retry", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[RolloutRetryArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "rollouts.retry",
		Description: "Retry a rollout analysis or progression",
	}, wrappedHandler)

	return nil
}
