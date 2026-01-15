package kubevirt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Toolset implements the KubeVirt toolset for VM lifecycle management.
type Toolset struct {
	provider        kubernetes.ClientProvider
	discovery       *kubernetes.CRDDiscovery
	vmGVR           schema.GroupVersionResource
	enabled         bool
	hasDataSource   bool
	hasInstanceType bool
	logger          *observability.Logger
	metrics         *observability.Metrics
}

// NewToolset creates a new KubeVirt toolset with improved CRD detection.
func NewToolset(provider kubernetes.ClientProvider, discovery *kubernetes.CRDDiscovery) *Toolset {
	// Check for VirtualMachine CRD (kubevirt.io/v1)
	vmGVK := schema.GroupVersionKind{
		Group:   "kubevirt.io",
		Version: "v1",
		Kind:    "VirtualMachine",
	}

	enabled := false
	var vmGVR schema.GroupVersionResource
	hasDataSource := false
	hasInstanceTypes := false

	if discovery != nil {
		// Ensure CRDs are discovered before checking
		// Use background context since this is initialization
		ctx := context.Background()
		if err := discovery.DiscoverCRDs(ctx); err != nil {
			// If discovery fails, we'll still try GetGVR in case cache is valid
		}

		// Check for VirtualMachine CRD
		if gvr, ok := discovery.GetGVR(vmGVK); ok {
			enabled = true
			vmGVR = gvr
		}

		// Also check for DataSource (cdi.kubevirt.io/v1beta1)
		dsGVK := schema.GroupVersionKind{
			Group:   "cdi.kubevirt.io",
			Version: "v1beta1",
			Kind:    "DataSource",
		}
		if _, ok := discovery.GetGVR(dsGVK); ok {
			hasDataSource = true
		}

		// Check for InstanceTypes (instancetypes.kubevirt.io)
		itGVK := schema.GroupVersionKind{
			Group:   "instancetypes.kubevirt.io",
			Version: "v1beta1",
			Kind:    "VirtualMachineInstancetype",
		}
		if _, ok := discovery.GetGVR(itGVK); ok {
			hasInstanceTypes = true
		}

		// Only enable if at least VirtualMachine is available
		// DataSource and InstanceTypes are optional but enhance functionality
		if !enabled {
			// Log that KubeVirt CRDs are not available
			// This will be handled by IsEnabled()
		}
	}

	return &Toolset{
		provider:        provider,
		discovery:       discovery,
		vmGVR:           vmGVR,
		enabled:         enabled,
		hasDataSource:   hasDataSource,
		hasInstanceType: hasInstanceTypes,
	}
}

// SetObservability sets the observability components for the toolset.
func (t *Toolset) SetObservability(logger *observability.Logger, metrics *observability.Metrics) {
	t.logger = logger
	t.metrics = metrics
}

// unmarshalArgs unmarshals args from map[string]interface{} to the target struct type.
func unmarshalArgs[T any](args any) (T, error) {
	var result T
	if args == nil {
		return result, nil
	}

	// If args is already the correct type, return it
	if typed, ok := args.(T); ok {
		return typed, nil
	}

	// If args is a map, unmarshal it
	if argsMap, ok := args.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(argsMap)
		if err != nil {
			return result, err
		}
		err = json.Unmarshal(jsonData, &result)
		return result, err
	}

	// Try to marshal and unmarshal as a fallback
	jsonData, err := json.Marshal(args)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(jsonData, &result)
	return result, err
}

// IsEnabled returns whether the KubeVirt toolset is enabled.
func (t *Toolset) IsEnabled() bool {
	return t.enabled
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "kubevirt"
}

// Tools returns all tools in this toolset (only if enabled).
func (t *Toolset) Tools() []*mcp.Tool {
	if !t.enabled {
		return []*mcp.Tool{}
	}

	return []*mcp.Tool{
		mcpHelpers.NewTool("kubevirt_vm_create", "Create a VirtualMachine").
			WithParameter("manifest", "object", "VirtualMachine manifest", true).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("kubevirt_vm_start", "Start a VirtualMachine").
			WithParameter("name", "string", "VM name", true).
			WithParameter("namespace", "string", "Namespace", true).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("kubevirt_vm_stop", "Stop a VirtualMachine").
			WithParameter("name", "string", "VM name", true).
			WithParameter("namespace", "string", "Namespace", true).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("kubevirt_vm_restart", "Restart a VirtualMachine").
			WithParameter("name", "string", "VM name", true).
			WithParameter("namespace", "string", "Namespace", true).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("kubevirt_datasources_list", "List KubeVirt DataSources").
			WithParameter("namespace", "string", "Namespace (empty for all)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("kubevirt_instancetypes_list", "List KubeVirt InstanceTypes").
			WithParameter("namespace", "string", "Namespace (empty for all)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
	}
}

// RegisterTools registers all tools from this toolset with the MCP server.
func (t *Toolset) RegisterTools(server *mcp.Server) error {
	if !t.enabled {
		return nil
	}

	var handler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error)
	var wrappedHandler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error)

	// Register kubevirt_vm_create
	type VMCreateArgs struct {
		Manifest map[string]any `json:"manifest"`
		Context  string         `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[VMCreateArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleVMCreate(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kubevirt_vm_create", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[VMCreateArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kubevirt_vm_create",
		Description: "Create a VirtualMachine",
	}, wrappedHandler)

	// Register kubevirt_vm_start
	type VMStartArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[VMStartArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleVMStart(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kubevirt_vm_start", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[VMStartArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kubevirt_vm_start",
		Description: "Start a VirtualMachine",
	}, wrappedHandler)

	// Register kubevirt_vm_stop
	type VMStopArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[VMStopArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleVMStop(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kubevirt_vm_stop", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[VMStopArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kubevirt_vm_stop",
		Description: "Stop a VirtualMachine",
	}, wrappedHandler)

	// Register kubevirt_vm_restart
	type VMRestartArgs struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[VMRestartArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleVMRestart(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kubevirt_vm_restart", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[VMRestartArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kubevirt_vm_restart",
		Description: "Restart a VirtualMachine",
	}, wrappedHandler)

	// Register kubevirt_datasources_list
	type DataSourcesListArgs struct {
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[DataSourcesListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleDataSourcesList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kubevirt_datasources_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[DataSourcesListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kubevirt_datasources_list",
		Description: "List KubeVirt DataSources",
	}, wrappedHandler)

	// Register kubevirt_instancetypes_list
	type InstanceTypesListArgs struct {
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}
	handler = func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		typedArgs, err := unmarshalArgs[InstanceTypesListArgs](args)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse arguments: %w", err)), nil, nil
		}
		result, err := t.handleInstanceTypesList(ctx, typedArgs)
		if err != nil {
			return mcpHelpers.NewErrorResult(err), nil, nil
		}
		return result, nil, nil
	}
	wrappedHandler = t.wrapToolHandler("kubevirt_instancetypes_list", handler, func(args any) string {
		typedArgs, _ := unmarshalArgs[InstanceTypesListArgs](args)
		return typedArgs.Context
	})
	mcpHelpers.AddTool(server, &mcp.Tool{
		Name:        "kubevirt_instancetypes_list",
		Description: "List KubeVirt InstanceTypes",
	}, wrappedHandler)

	return nil
}

// handleVMCreate handles the kubevirt_vm_create tool.
func (t *Toolset) handleVMCreate(ctx context.Context, args struct {
	Manifest map[string]any `json:"manifest"`
	Context  string         `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	obj := &unstructured.Unstructured{Object: args.Manifest}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kubevirt.io",
		Version: "v1",
		Kind:    "VirtualMachine",
	})

	created, err := clientSet.Dynamic.Resource(t.vmGVR).Namespace(obj.GetNamespace()).
		Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create VM: %w", err)), nil
	}

	return mcpHelpers.NewJSONResult(map[string]any{
		"name":      created.GetName(),
		"namespace": created.GetNamespace(),
		"status":    "created",
	})
}

// handleVMStart handles the kubevirt_vm_start tool.
func (t *Toolset) handleVMStart(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleVMAction(ctx, args, "start")
}

// handleVMStop handles the kubevirt_vm_stop tool.
func (t *Toolset) handleVMStop(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleVMAction(ctx, args, "stop")
}

// handleVMRestart handles the kubevirt_vm_restart tool.
func (t *Toolset) handleVMRestart(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleVMAction(ctx, args, "restart")
}

// handleVMAction handles VM lifecycle actions.
func (t *Toolset) handleVMAction(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}, action string) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	running := action == "start" || action == "restart"
	patchData := fmt.Sprintf(`{"spec":{"running":%v}}`, running)

	patched, err := clientSet.Dynamic.Resource(t.vmGVR).Namespace(args.Namespace).
		Patch(ctx, args.Name, "merge", []byte(patchData), metav1.PatchOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to %s VM: %w", action, err)), nil
	}

	return mcpHelpers.NewJSONResult(map[string]any{
		"name":   patched.GetName(),
		"action": action,
		"status": "success",
	})
}

// handleDataSourcesList handles the kubevirt_datasources_list tool.
func (t *Toolset) handleDataSourcesList(ctx context.Context, args struct {
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	dsGVK := schema.GroupVersionKind{
		Group:   "cdi.kubevirt.io",
		Version: "v1beta1",
		Kind:    "DataSource",
	}

	var dsGVR schema.GroupVersionResource
	if t.discovery != nil {
		// Ensure discovery is up to date
		if err := t.discovery.DiscoverCRDs(ctx); err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to discover CRDs: %w", err)), nil
		}
		if gvr, ok := t.discovery.GetGVR(dsGVK); ok {
			dsGVR = gvr
		} else {
			return mcpHelpers.NewErrorResult(fmt.Errorf("DataSource CRD not installed. Install CDI (Containerized Data Importer) to use this feature")), nil
		}
	} else {
		dsGVR = schema.GroupVersionResource{
			Group:    "cdi.kubevirt.io",
			Version:  "v1beta1",
			Resource: "datasources",
		}
	}

	list, err := clientSet.Dynamic.Resource(dsGVR).Namespace(args.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list DataSources: %w", err)), nil
	}

	datasources := make([]map[string]any, 0, len(list.Items))
	for _, ds := range list.Items {
		datasources = append(datasources, map[string]any{
			"name":      ds.GetName(),
			"namespace": ds.GetNamespace(),
		})
	}

	return mcpHelpers.NewJSONResult(map[string]any{"datasources": datasources})
}

// handleInstanceTypesList handles the kubevirt_instancetypes_list tool.
func (t *Toolset) handleInstanceTypesList(ctx context.Context, args struct {
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	itGVK := schema.GroupVersionKind{
		Group:   "instancetypes.kubevirt.io",
		Version: "v1beta1",
		Kind:    "VirtualMachineInstancetype",
	}

	var itGVR schema.GroupVersionResource
	if t.discovery != nil {
		// Ensure discovery is up to date
		if err := t.discovery.DiscoverCRDs(ctx); err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to discover CRDs: %w", err)), nil
		}
		if gvr, ok := t.discovery.GetGVR(itGVK); ok {
			itGVR = gvr
		} else {
			return mcpHelpers.NewErrorResult(fmt.Errorf("InstanceType CRD not installed. Install KubeVirt InstanceTypes to use this feature")), nil
		}
	} else {
		itGVR = schema.GroupVersionResource{
			Group:    "instancetypes.kubevirt.io",
			Version:  "v1beta1",
			Resource: "virtualmachineinstancetypes",
		}
	}

	list, err := clientSet.Dynamic.Resource(itGVR).Namespace(args.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list InstanceTypes: %w", err)), nil
	}

	instancetypes := make([]map[string]any, 0, len(list.Items))
	for _, it := range list.Items {
		instancetypes = append(instancetypes, map[string]any{
			"name":      it.GetName(),
			"namespace": it.GetNamespace(),
		})
	}

	return mcpHelpers.NewJSONResult(map[string]any{"instancetypes": instancetypes})
}
