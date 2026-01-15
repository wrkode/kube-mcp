package net

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
)

// Tools returns all tools in this toolset.
func (t *Toolset) Tools() []*mcp.Tool {
	tools := []*mcp.Tool{}

	// net.networkpolicies_list
	tools = append(tools, mcpHelpers.NewTool("net.networkpolicies_list", "List NetworkPolicies").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
		WithParameter("label_selector", "string", "Label selector", false).
		WithParameter("limit", "integer", "Maximum number of items to return", false).
		WithParameter("continue", "string", "Token from previous paginated request", false).
		WithReadOnly().
		Build())

	// net.networkpolicy_explain
	tools = append(tools, mcpHelpers.NewTool("net.networkpolicy_explain", "Explain NetworkPolicy rules").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("name", "string", "NetworkPolicy name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithReadOnly().
		Build())

	// net.connectivity_hint
	tools = append(tools, mcpHelpers.NewTool("net.connectivity_hint", "Analyze connectivity between pods").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("src_namespace", "string", "Source namespace", true).
		WithParameter("src_labels", "object", "Source pod labels", true).
		WithParameter("dst_namespace", "string", "Destination namespace", true).
		WithParameter("dst_labels", "object", "Destination pod labels", true).
		WithParameter("port", "string", "Port number", true).
		WithParameter("protocol", "string", "Protocol (TCP, UDP, SCTP)", true).
		WithReadOnly().
		Build())

	// Cilium tools (only if CRDs exist)
	if t.hasCilium {
		tools = append(tools, mcpHelpers.NewTool("net.cilium_policies_list", "List Cilium network policies").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
			WithParameter("label_selector", "string", "Label selector", false).
			WithParameter("limit", "integer", "Maximum number of items to return", false).
			WithParameter("continue", "string", "Token from previous paginated request", false).
			WithReadOnly().
			Build())

		tools = append(tools, mcpHelpers.NewTool("net.cilium_policy_get", "Get Cilium policy details").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("kind", "string", "Policy kind: 'CiliumNetworkPolicy' or 'CiliumClusterwideNetworkPolicy'", true).
			WithParameter("name", "string", "Policy name", true).
			WithParameter("namespace", "string", "Namespace name (ignored for Clusterwide)", false).
			WithParameter("raw", "boolean", "Return raw object if true", false).
			WithReadOnly().
			Build())
	}

	// Hubble tool (only if configured)
	if t.hasHubble {
		tools = append(tools, mcpHelpers.NewTool("net.hubble_flows_query", "Query Hubble flows").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("namespace", "string", "Filter by namespace", false).
			WithParameter("pod", "string", "Filter by pod name", false).
			WithParameter("verdict", "string", "Filter by verdict (FORWARDED, DROPPED, etc.)", false).
			WithParameter("protocol", "string", "Filter by protocol", false).
			WithParameter("since_seconds", "integer", "Query flows since N seconds ago", false).
			WithParameter("limit", "integer", "Maximum number of flows to return", false).
			WithReadOnly().
			Build())
	}

	return tools
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	// Register net.networkpolicies_list
	type NetworkPoliciesListArgs struct {
		Context       string `json:"context"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[NetworkPoliciesListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleNetworkPoliciesList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("net.networkpolicies_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[NetworkPoliciesListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "net.networkpolicies_list",
		Description: "List NetworkPolicies",
	}, wrappedHandler)

	// Register net.networkpolicy_explain
	type NetworkPolicyExplainArgs struct {
		Context   string `json:"context"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[NetworkPolicyExplainArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleNetworkPolicyExplain(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("net.networkpolicy_explain", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[NetworkPolicyExplainArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "net.networkpolicy_explain",
		Description: "Explain NetworkPolicy rules",
	}, wrappedHandler)

	// Register net.connectivity_hint
	type ConnectivityHintArgs struct {
		Context      string            `json:"context"`
		SrcNamespace string            `json:"src_namespace"`
		SrcLabels    map[string]string `json:"src_labels"`
		DstNamespace string            `json:"dst_namespace"`
		DstLabels    map[string]string `json:"dst_labels"`
		Port         string            `json:"port"`
		Protocol     string            `json:"protocol"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ConnectivityHintArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleConnectivityHint(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("net.connectivity_hint", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ConnectivityHintArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "net.connectivity_hint",
		Description: "Analyze connectivity between pods",
	}, wrappedHandler)

	// Register Cilium tools if available
	if t.hasCilium {
		// net.cilium_policies_list
		type CiliumPoliciesListArgs struct {
			Context       string `json:"context"`
			Namespace     string `json:"namespace"`
			LabelSelector string `json:"label_selector"`
			Limit         int    `json:"limit"`
			Continue      string `json:"continue"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[CiliumPoliciesListArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleCiliumPoliciesList(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("net.cilium_policies_list", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[CiliumPoliciesListArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "net.cilium_policies_list",
			Description: "List Cilium network policies",
		}, wrappedHandler)

		// net.cilium_policy_get
		type CiliumPolicyGetArgs struct {
			Context   string `json:"context"`
			Kind      string `json:"kind"`
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Raw       bool   `json:"raw"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[CiliumPolicyGetArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleCiliumPolicyGet(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("net.cilium_policy_get", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[CiliumPolicyGetArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "net.cilium_policy_get",
			Description: "Get Cilium policy details",
		}, wrappedHandler)
	}

	// Register Hubble tool if available
	if t.hasHubble {
		type HubbleFlowsQueryArgs struct {
			Context      string `json:"context"`
			Namespace    string `json:"namespace,omitempty"`
			Pod          string `json:"pod,omitempty"`
			Verdict      string `json:"verdict,omitempty"`
			Protocol     string `json:"protocol,omitempty"`
			SinceSeconds int    `json:"since_seconds,omitempty"`
			Limit        int    `json:"limit,omitempty"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[HubbleFlowsQueryArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleHubbleFlowsQuery(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("net.hubble_flows_query", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[HubbleFlowsQueryArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "net.hubble_flows_query",
			Description: "Query Hubble flows",
		}, wrappedHandler)
	}

	return nil
}
