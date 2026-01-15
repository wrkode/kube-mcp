package capi

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

	// capi.clusters_list
	tools = append(tools, mcpHelpers.NewTool("capi.clusters_list", "List Cluster API clusters").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
		WithParameter("label_selector", "string", "Label selector", false).
		WithReadOnly().
		Build())

	// capi.cluster_get
	tools = append(tools, mcpHelpers.NewTool("capi.cluster_get", "Get Cluster API cluster details").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("namespace", "string", "Namespace name", true).
		WithParameter("name", "string", "Cluster name", true).
		WithParameter("raw", "boolean", "Return raw object if true", false).
		WithReadOnly().
		Build())

	// capi.machines_list (if Machine CRD exists)
	if t.hasMachine {
		tools = append(tools, mcpHelpers.NewTool("capi.machines_list", "List machines for a cluster").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("cluster_namespace", "string", "Cluster namespace", true).
			WithParameter("cluster_name", "string", "Cluster name", true).
			WithParameter("limit", "integer", "Maximum number of items to return", false).
			WithParameter("continue", "string", "Token from previous paginated request", false).
			WithReadOnly().
			Build())
	}

	// capi.machinedeployments_list (if MachineDeployment CRD exists)
	if t.hasMachineDeployment {
		tools = append(tools, mcpHelpers.NewTool("capi.machinedeployments_list", "List machine deployments for a cluster").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("cluster_namespace", "string", "Cluster namespace", true).
			WithParameter("cluster_name", "string", "Cluster name", true).
			WithReadOnly().
			Build())
	}

	// capi.rollout_status
	tools = append(tools, mcpHelpers.NewTool("capi.rollout_status", "Get cluster rollout status").
		WithParameter("context", "string", "Kubernetes context name", false).
		WithParameter("cluster_namespace", "string", "Cluster namespace", true).
		WithParameter("cluster_name", "string", "Cluster name", true).
		WithReadOnly().
		Build())

	// capi.scale_machinedeployment (if MachineDeployment CRD exists)
	if t.hasMachineDeployment {
		tools = append(tools, mcpHelpers.NewTool("capi.scale_machinedeployment", "Scale a machine deployment").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("name", "string", "MachineDeployment name", true).
			WithParameter("replicas", "integer", "Number of replicas", true).
			WithParameter("confirm", "boolean", "Must be true to scale", true).
			WithDestructive().
			Build())
	}

	return tools
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	if !t.enabled {
		return nil
	}

	// Register capi.clusters_list
	type ClustersListArgs struct {
		Context       string `json:"context"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ClustersListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleClustersList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("capi.clusters_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ClustersListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "capi.clusters_list",
		Description: "List Cluster API clusters",
	}, wrappedHandler)

	// Register capi.cluster_get
	type ClusterGetArgs struct {
		Context   string `json:"context"`
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Raw       bool   `json:"raw"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ClusterGetArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleClusterGet(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("capi.cluster_get", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ClusterGetArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "capi.cluster_get",
		Description: "Get Cluster API cluster details",
	}, wrappedHandler)

	// Register capi.machines_list
	if t.hasMachine {
		type MachinesListArgs struct {
			Context          string `json:"context"`
			ClusterNamespace string `json:"cluster_namespace"`
			ClusterName      string `json:"cluster_name"`
			Limit            int    `json:"limit"`
			Continue         string `json:"continue"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[MachinesListArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleMachinesList(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("capi.machines_list", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[MachinesListArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "capi.machines_list",
			Description: "List machines for a cluster",
		}, wrappedHandler)
	}

	// Register capi.machinedeployments_list
	if t.hasMachineDeployment {
		type MachineDeploymentsListArgs struct {
			Context          string `json:"context"`
			ClusterNamespace string `json:"cluster_namespace"`
			ClusterName      string `json:"cluster_name"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[MachineDeploymentsListArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleMachineDeploymentsList(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("capi.machinedeployments_list", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[MachineDeploymentsListArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "capi.machinedeployments_list",
			Description: "List machine deployments for a cluster",
		}, wrappedHandler)
	}

	// Register capi.rollout_status
	type RolloutStatusArgs struct {
		Context          string `json:"context"`
		ClusterNamespace string `json:"cluster_namespace"`
		ClusterName      string `json:"cluster_name"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[RolloutStatusArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleRolloutStatus(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("capi.rollout_status", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[RolloutStatusArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "capi.rollout_status",
		Description: "Get cluster rollout status",
	}, wrappedHandler)

	// Register capi.scale_machinedeployment
	if t.hasMachineDeployment {
		type ScaleMachineDeploymentArgs struct {
			Context   string `json:"context"`
			Namespace string `json:"namespace"`
			Name      string `json:"name"`
			Replicas  int    `json:"replicas"`
			Confirm   bool   `json:"confirm"`
		}
		handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
			typedArgs, err := unmarshalArgs[ScaleMachineDeploymentArgs](args)
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
			}
			result, err := t.handleScaleMachineDeployment(ctx, typedArgs)
			if err != nil {
				return mcpHelpers.NewErrorResult(err), nil, nil
			}
			return result, nil, nil
		}
		wrappedHandler = t.wrapToolHandler("capi.scale_machinedeployment", handler, func(args any) string {
			typedArgs, _ := unmarshalArgs[ScaleMachineDeploymentArgs](args)
			return typedArgs.Context
		})
		mcpHelpers.AddTool(server, &mcp.Tool{
			Name:        "capi.scale_machinedeployment",
			Description: "Scale a machine deployment",
		}, wrappedHandler)
	}

	return nil
}
