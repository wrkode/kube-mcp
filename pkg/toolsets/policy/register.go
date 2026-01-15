package policy

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

	return []*mcp.Tool{
		mcpHelpers.NewTool("policy.policies_list", "List policy policies (Kyverno or Gatekeeper)").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("namespace", "string", "Namespace name (for Kyverno namespaced Policy)", false).
			WithParameter("engine", "string", "Policy engine: 'kyverno', 'gatekeeper', or 'all' (default: all available)", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("policy.policy_get", "Get policy details").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("engine", "string", "Policy engine: 'kyverno' or 'gatekeeper'", true).
			WithParameter("kind", "string", "Policy kind (e.g., 'ClusterPolicy', 'Policy', 'ConstraintTemplate')", true).
			WithParameter("name", "string", "Policy name", true).
			WithParameter("namespace", "string", "Namespace name (required for namespaced policies)", false).
			WithParameter("raw", "boolean", "Return raw object if true", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("policy.violations_list", "List policy violations").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
			WithParameter("engine", "string", "Policy engine: 'kyverno', 'gatekeeper', or 'all' (default: all available)", false).
			WithParameter("limit", "integer", "Maximum number of items to return", false).
			WithParameter("continue", "string", "Token from previous paginated request", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("policy.explain_denial", "Explain an admission denial message (heuristic)").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("message", "string", "Admission denial message or event text", true).
			WithReadOnly().
			Build(),
	}
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	if !t.enabled {
		return nil
	}

	// Register policy.policies_list
	type PoliciesListArgs struct {
		Context   string `json:"context"`
		Namespace string `json:"namespace"`
		Engine    string `json:"engine"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[PoliciesListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handlePoliciesList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("policy.policies_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[PoliciesListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "policy.policies_list",
		Description: "List policy policies",
	}, wrappedHandler)

	// Register policy.policy_get
	type PolicyGetArgs struct {
		Context   string `json:"context"`
		Engine    string `json:"engine"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Raw       bool   `json:"raw"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[PolicyGetArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handlePolicyGet(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("policy.policy_get", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[PolicyGetArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "policy.policy_get",
		Description: "Get policy details",
	}, wrappedHandler)

	// Register policy.violations_list
	type ViolationsListArgs struct {
		Context   string `json:"context"`
		Namespace string `json:"namespace"`
		Engine    string `json:"engine"`
		Limit     int    `json:"limit"`
		Continue  string `json:"continue"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ViolationsListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleViolationsList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("policy.violations_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ViolationsListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "policy.violations_list",
		Description: "List policy violations",
	}, wrappedHandler)

	// Register policy.explain_denial
	type ExplainDenialArgs struct {
		Context string `json:"context"`
		Message string `json:"message"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ExplainDenialArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleExplainDenial(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("policy.explain_denial", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ExplainDenialArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "policy.explain_denial",
		Description: "Explain an admission denial message (heuristic)",
	}, wrappedHandler)

	return nil
}
