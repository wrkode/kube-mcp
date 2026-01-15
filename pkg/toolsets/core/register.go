package core

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
)

// registerPodTools registers pod-related tools with observability.
func (t *Toolset) registerPodTools(server *mcp.Server) {
	// pods_list
	type PodsListArgs struct {
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		FieldSelector string `json:"field_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
		Context       string `json:"context"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[PodsListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handlePodsList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("pods_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[PodsListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "pods_list",
		Description: "List pods in a namespace or all namespaces",
	}, wrappedHandler)

	// pods_get
	type PodsGetArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[PodsGetArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handlePodsGet(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("pods_get", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[PodsGetArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "pods_get",
		Description: "Get pod details",
	}, wrappedHandler)

	// pods_delete
	type PodsDeleteArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[PodsDeleteArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handlePodsDelete(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("pods_delete", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[PodsDeleteArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "pods_delete",
		Description: "Delete a pod",
	}, wrappedHandler)

	// pods_logs
	type PodsLogsArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Container string `json:"container"`
		TailLines *int   `json:"tail_lines"`
		Since     string `json:"since"`
		SinceTime string `json:"since_time"`
		Previous  bool   `json:"previous"`
		Follow    bool   `json:"follow"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[PodsLogsArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handlePodsLogs(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("pods_logs", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[PodsLogsArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "pods_logs",
		Description: "Fetch pod logs",
	}, wrappedHandler)

	// pods_exec
	type PodsExecArgs struct {
		Name      string   `json:"name"`
		Namespace string   `json:"namespace"`
		Container string   `json:"container"`
		Command   []string `json:"command"`
		Context   string   `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[PodsExecArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handlePodsExec(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("pods_exec", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[PodsExecArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "pods_exec",
		Description: "Execute command in pod",
	}, wrappedHandler)

	// pods_top
	type PodsTopArgs struct {
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[PodsTopArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handlePodsTop(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("pods_top", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[PodsTopArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "pods_top",
		Description: "Get pod resource usage metrics from metrics.k8s.io API",
	}, wrappedHandler)

	// pods_port_forward
	type PodsPortForwardArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		LocalPort int    `json:"local_port"`
		PodPort   int    `json:"pod_port"`
		Container string `json:"container"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[PodsPortForwardArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handlePodsPortForward(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("pods_port_forward", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[PodsPortForwardArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "pods_port_forward",
		Description: "Set up port forwarding from local port to pod port",
	}, wrappedHandler)
}

// registerResourceTools registers resource-related tools with observability.
func (t *Toolset) registerResourceTools(server *mcp.Server) {
	// resources_list
	type ResourcesListArgs struct {
		Group         string `json:"group"`
		Version       string `json:"version"`
		Kind          string `json:"kind"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		FieldSelector string `json:"field_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
		Context       string `json:"context"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("resources_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_list",
		Description: "List resources by GroupVersionKind",
	}, wrappedHandler)

	// resources_get
	type ResourcesGetArgs struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesGetArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesGet(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_get", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesGetArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_get",
		Description: "Get a resource",
	}, wrappedHandler)

	// resources_describe
	type ResourcesDescribeArgs struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesDescribeArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesDescribe(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_describe", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesDescribeArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_describe",
		Description: "Describe a resource in kubectl-style format",
	}, wrappedHandler)

	// resources_diff
	type ResourcesDiffArgs struct {
		Group      string                 `json:"group"`
		Version    string                 `json:"version"`
		Kind       string                 `json:"kind"`
		Name       string                 `json:"name"`
		Namespace  string                 `json:"namespace"`
		Manifest   map[string]interface{} `json:"manifest"`
		DiffFormat string                 `json:"diff_format"`
		Context    string                 `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesDiffArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesDiff(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_diff", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesDiffArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_diff",
		Description: "Compare current resource state with desired manifest and show differences",
	}, wrappedHandler)

	// resources_validate
	type ResourcesValidateArgs struct {
		Manifest      map[string]interface{} `json:"manifest"`
		SchemaVersion string                 `json:"schema_version"`
		Context       string                 `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesValidateArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesValidate(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_validate", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesValidateArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_validate",
		Description: "Validate a resource manifest without applying it",
	}, wrappedHandler)

	// resources_relationships
	type ResourcesRelationshipsArgs struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Direction string `json:"direction"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesRelationshipsArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesRelationships(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_relationships", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesRelationshipsArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_relationships",
		Description: "Find resource owners and/or dependents",
	}, wrappedHandler)

	// configmaps_get_data
	type ConfigMapsGetDataArgs struct {
		Name      string   `json:"name"`
		Namespace string   `json:"namespace"`
		Keys      []string `json:"keys"`
		Context   string   `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ConfigMapsGetDataArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleConfigMapsGetData(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("configmaps_get_data", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ConfigMapsGetDataArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "configmaps_get_data",
		Description: "Get ConfigMap data",
	}, wrappedHandler)

	// configmaps_set_data
	type ConfigMapsSetDataArgs struct {
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Data      map[string]string `json:"data"`
		Merge     bool              `json:"merge"`
		Context   string            `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ConfigMapsSetDataArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleConfigMapsSetData(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("configmaps_set_data", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ConfigMapsSetDataArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "configmaps_set_data",
		Description: "Update ConfigMap data",
	}, wrappedHandler)

	// secrets_get_data
	type SecretsGetDataArgs struct {
		Name      string   `json:"name"`
		Namespace string   `json:"namespace"`
		Keys      []string `json:"keys"`
		Decode    bool     `json:"decode"`
		Context   string   `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[SecretsGetDataArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleSecretsGetData(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("secrets_get_data", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[SecretsGetDataArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "secrets_get_data",
		Description: "Get Secret data",
	}, wrappedHandler)

	// secrets_set_data
	type SecretsSetDataArgs struct {
		Name      string            `json:"name"`
		Namespace string            `json:"namespace"`
		Data      map[string]string `json:"data"`
		Merge     bool              `json:"merge"`
		Encode    bool              `json:"encode"`
		Context   string            `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[SecretsSetDataArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleSecretsSetData(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("secrets_set_data", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[SecretsSetDataArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "secrets_set_data",
		Description: "Update Secret data",
	}, wrappedHandler)

	// resources_apply
	type ResourcesApplyArgs struct {
		Manifest     map[string]any `json:"manifest"`
		FieldManager string         `json:"field_manager"`
		DryRun       bool           `json:"dry_run"`
		Context      string         `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesApplyArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesApply(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_apply", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesApplyArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_apply",
		Description: "Create or update a resource using server-side apply",
	}, wrappedHandler)

	// resources_patch
	type ResourcesPatchArgs struct {
		Group        string      `json:"group"`
		Version      string      `json:"version"`
		Kind         string      `json:"kind"`
		Name         string      `json:"name"`
		Namespace    string      `json:"namespace"`
		PatchType    string      `json:"patch_type"`
		PatchData    interface{} `json:"patch_data"` // object for merge/strategic, array for json patch
		FieldManager string      `json:"field_manager"`
		DryRun       bool        `json:"dry_run"`
		Context      string      `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesPatchArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesPatch(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_patch", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesPatchArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_patch",
		Description: "Partially update a resource using JSON Patch, Merge Patch, or Strategic Merge Patch",
	}, wrappedHandler)

	// resources_delete
	type ResourcesDeleteArgs struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		DryRun    bool   `json:"dry_run"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesDeleteArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesDelete(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_delete", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesDeleteArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_delete",
		Description: "Delete a resource",
	}, wrappedHandler)

	// resources_scale
	type ResourcesScaleArgs struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Replicas  *int   `json:"replicas"` // nil = get-only, 0 = scale to zero, >0 = scale to that number
		DryRun    bool   `json:"dry_run"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesScaleArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesScale(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_scale", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesScaleArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_scale",
		Description: "Scale a resource",
	}, wrappedHandler)

	// resources_watch
	type ResourcesWatchArgs struct {
		Group         string `json:"group"`
		Version       string `json:"version"`
		Kind          string `json:"kind"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		FieldSelector string `json:"field_selector"`
		Timeout       int    `json:"timeout"`
		Context       string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[ResourcesWatchArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleResourcesWatch(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("resources_watch", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[ResourcesWatchArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "resources_watch",
		Description: "Watch resources for changes (returns events within timeout)",
	}, wrappedHandler)
}

// registerNamespaceTools registers namespace-related tools with observability.
func (t *Toolset) registerNamespaceTools(server *mcp.Server) {
	type NamespacesListArgs struct {
		Context string `json:"context"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[NamespacesListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleNamespacesList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("namespaces_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[NamespacesListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "namespaces_list",
		Description: "List all namespaces",
	}, wrappedHandler)
}

// registerNodeTools registers node-related tools with observability.
func (t *Toolset) registerNodeTools(server *mcp.Server) {
	type NodesTopArgs struct {
		Context string `json:"context"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[NodesTopArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleNodesTop(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("nodes_top", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[NodesTopArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "nodes_top",
		Description: "Get node resource usage metrics from metrics.k8s.io API",
	}, wrappedHandler)

	type NodesSummaryArgs struct {
		Name    string `json:"name"`
		Context string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[NodesSummaryArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleNodesSummary(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("nodes_summary", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[NodesSummaryArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "nodes_summary",
		Description: "Get node summary statistics",
	}, wrappedHandler)
}

// registerEventTools registers event-related tools with observability.
func (t *Toolset) registerEventTools(server *mcp.Server) {
	type EventsListArgs struct {
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[EventsListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleEventsList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler := t.wrapToolHandler("events_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[EventsListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "events_list",
		Description: "List events",
	}, wrappedHandler)
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	// Register pod tools
	t.registerPodTools(server)
	// Register resource tools
	t.registerResourceTools(server)
	// Register namespace tools
	t.registerNamespaceTools(server)
	// Register node tools
	t.registerNodeTools(server)
	// Register event tools
	t.registerEventTools(server)
	return nil
}
