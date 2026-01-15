package autoscaling

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
)

// Tools returns all tools in this toolset.
func (t *Toolset) Tools() []*mcp.Tool {
	tools := []*mcp.Tool{}

	// autoscaling.hpa_list
	tools = append(tools, mcpHelpers.NewTool("autoscaling.hpa_list", "List HorizontalPodAutoscalers").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
		WithParameter("label_selector", "string", "Label selector", false).
		WithParameter("limit", "integer", "Maximum number of items to return", false).
		WithParameter("continue", "string", "Token from previous paginated request", false).
		WithReadOnly().
		Build())

	// autoscaling.hpa_explain
	tools = append(tools, mcpHelpers.NewTool("autoscaling.hpa_explain", "Explain HPA status and metrics").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("name", "string", "HPA name", true).
		WithParameter("namespace", "string", "Namespace name", true).
		WithReadOnly().
		Build())

	// KEDA tools (only if KEDA CRDs exist)
	if t.hasKEDA {
		tools = append(tools, mcpHelpers.NewTool("autoscaling.keda_scaledobjects_list", "List KEDA ScaledObjects").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
			WithParameter("label_selector", "string", "Label selector", false).
			WithParameter("limit", "integer", "Maximum number of items to return", false).
			WithParameter("continue", "string", "Token from previous paginated request", false).
			WithReadOnly().
			Build())

		tools = append(tools, mcpHelpers.NewTool("autoscaling.keda_scaledobject_get", "Get KEDA ScaledObject details").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("name", "string", "ScaledObject name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("raw", "boolean", "Return raw object if true", false).
			WithReadOnly().
			Build())

		tools = append(tools, mcpHelpers.NewTool("autoscaling.keda_triggers_explain", "Explain KEDA triggers").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("name", "string", "ScaledObject name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithReadOnly().
			Build())

		tools = append(tools, mcpHelpers.NewTool("autoscaling.keda_pause", "Pause KEDA autoscaling").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("name", "string", "ScaledObject name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("confirm", "boolean", "Must be true to pause", true).
			WithDestructive().
			Build())

		tools = append(tools, mcpHelpers.NewTool("autoscaling.keda_resume", "Resume KEDA autoscaling").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("name", "string", "ScaledObject name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("confirm", "boolean", "Must be true to resume", true).
			WithDestructive().
			Build())
	}

	return tools
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	// Register autoscaling.hpa_list
	type HPAListArgs struct {
		Context       string `json:"context"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[HPAListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleHPAList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("autoscaling.hpa_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[HPAListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "autoscaling.hpa_list",
		Description: "List HorizontalPodAutoscalers",
	}, wrappedHandler)

	// Register autoscaling.hpa_explain
	type HPAExplainArgs struct {
		Context   string `json:"context"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[HPAExplainArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleHPAExplain(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("autoscaling.hpa_explain", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[HPAExplainArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "autoscaling.hpa_explain",
		Description: "Explain HPA status and metrics",
	}, wrappedHandler)

	// Register KEDA tools if available
	if t.hasKEDA {
		// autoscaling.keda_scaledobjects_list
		type KEDAScaledObjectsListArgs struct {
			Context       string `json:"context"`
			Namespace     string `json:"namespace"`
			LabelSelector string `json:"label_selector"`
			Limit         int    `json:"limit"`
			Continue      string `json:"continue"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[KEDAScaledObjectsListArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleKEDAScaledObjectsList(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("autoscaling.keda_scaledobjects_list", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[KEDAScaledObjectsListArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "autoscaling.keda_scaledobjects_list",
			Description: "List KEDA ScaledObjects",
		}, wrappedHandler)

		// autoscaling.keda_scaledobject_get
		type KEDAScaledObjectGetArgs struct {
			Context   string `json:"context"`
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Raw       bool   `json:"raw"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[KEDAScaledObjectGetArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleKEDAScaledObjectGet(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("autoscaling.keda_scaledobject_get", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[KEDAScaledObjectGetArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "autoscaling.keda_scaledobject_get",
			Description: "Get KEDA ScaledObject details",
		}, wrappedHandler)

		// autoscaling.keda_triggers_explain
		type KEDATriggersExplainArgs struct {
			Context   string `json:"context"`
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[KEDATriggersExplainArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleKEDATriggersExplain(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("autoscaling.keda_triggers_explain", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[KEDATriggersExplainArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "autoscaling.keda_triggers_explain",
			Description: "Explain KEDA triggers",
		}, wrappedHandler)

		// autoscaling.keda_pause
		type KEDAPauseArgs struct {
			Context   string `json:"context"`
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Confirm   bool   `json:"confirm"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[KEDAPauseArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleKEDAPause(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("autoscaling.keda_pause", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[KEDAPauseArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "autoscaling.keda_pause",
			Description: "Pause KEDA autoscaling",
		}, wrappedHandler)

		// autoscaling.keda_resume
		type KEDAResumeArgs struct {
			Context   string `json:"context"`
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Confirm   bool   `json:"confirm"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[KEDAResumeArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleKEDAResume(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("autoscaling.keda_resume", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[KEDAResumeArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "autoscaling.keda_resume",
			Description: "Resume KEDA autoscaling",
		}, wrappedHandler)
	}

	return nil
}
