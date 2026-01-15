package core

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
)

// Toolset implements the Core toolset for Kubernetes operations.
type Toolset struct {
	provider       kubernetes.ClientProvider
	logger         *observability.Logger
	metrics        *observability.Metrics
	rbacAuthorizer kubernetes.RBACAuthorizer
	requireRBAC    bool
}

// NewToolset creates a new Core toolset.
func NewToolset(provider kubernetes.ClientProvider) *Toolset {
	return &Toolset{
		provider: provider,
	}
}

// SetObservability sets the observability components for the toolset.
func (t *Toolset) SetObservability(logger *observability.Logger, metrics *observability.Metrics) {
	t.logger = logger
	t.metrics = metrics
}

// SetRBACAuthorizer sets the RBAC authorizer for the toolset.
func (t *Toolset) SetRBACAuthorizer(authorizer kubernetes.RBACAuthorizer, requireRBAC bool) {
	t.rbacAuthorizer = authorizer
	t.requireRBAC = requireRBAC
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "core"
}

// Tools returns all tools in this toolset.
func (t *Toolset) Tools() []*mcp.Tool {
	return []*mcp.Tool{
		// Pod tools
		mcpHelpers.NewTool("pods_list", "List pods in a namespace or all namespaces").
			WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
			WithParameter("label_selector", "string", "Label selector (e.g., 'app=frontend' or 'app in (frontend,backend)')", false).
			WithParameter("field_selector", "string", "Field selector (e.g., 'status.phase=Running')", false).
			WithParameter("limit", "integer", "Maximum number of items to return", false).
			WithParameter("continue", "string", "Token from previous paginated request", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("pods_get", "Get pod details").
			WithParameter("name", "string", "Pod name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("pods_delete", "Delete a pod").
			WithParameter("name", "string", "Pod name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("pods_logs", "Fetch pod logs").
			WithParameter("name", "string", "Pod name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("container", "string", "Container name (optional)", false).
			WithParameter("tail_lines", "integer", "Number of lines to tail", false).
			WithParameter("since", "string", "Duration string (e.g., '5m', '1h') to fetch logs since", false).
			WithParameter("since_time", "string", "RFC3339 timestamp to fetch logs since", false).
			WithParameter("previous", "boolean", "Fetch logs from previous container instance", false).
			WithParameter("follow", "boolean", "Follow log stream (for real-time streaming, use HTTP transport)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("pods_exec", "Execute command in pod").
			WithParameter("name", "string", "Pod name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("container", "string", "Container name (optional)", false).
			WithParameter("command", "array", "Command to execute", true).
			WithParameter("context", "string", "Kubernetes context name", false).
			Build(),
		mcpHelpers.NewTool("pods_top", "Get pod resource usage metrics").
			WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("pods_port_forward", "Set up port forwarding from local port to pod port").
			WithParameter("name", "string", "Pod name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("local_port", "integer", "Local port to forward from", true).
			WithParameter("pod_port", "integer", "Pod port to forward to", true).
			WithParameter("container", "string", "Container name (optional)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			Build(),
		// Resource tools
		mcpHelpers.NewTool("resources_list", "List resources by GroupVersionKind").
			WithParameter("group", "string", "API group", false).
			WithParameter("version", "string", "API version", true).
			WithParameter("kind", "string", "Resource kind", true).
			WithParameter("namespace", "string", "Namespace name (empty for cluster-scoped)", false).
			WithParameter("label_selector", "string", "Label selector (e.g., 'app=frontend' or 'app in (frontend,backend)')", false).
			WithParameter("field_selector", "string", "Field selector (e.g., 'status.phase=Running')", false).
			WithParameter("limit", "integer", "Maximum number of items to return", false).
			WithParameter("continue", "string", "Token from previous paginated request", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("resources_get", "Get a resource").
			WithParameter("group", "string", "API group", false).
			WithParameter("version", "string", "API version", true).
			WithParameter("kind", "string", "Resource kind", true).
			WithParameter("name", "string", "Resource name", true).
			WithParameter("namespace", "string", "Namespace name (empty for cluster-scoped)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("resources_describe", "Describe a resource in kubectl-style format").
			WithParameter("group", "string", "API group", false).
			WithParameter("version", "string", "API version", true).
			WithParameter("kind", "string", "Resource kind", true).
			WithParameter("name", "string", "Resource name", true).
			WithParameter("namespace", "string", "Namespace name (empty for cluster-scoped)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("resources_diff", "Compare current resource state with desired manifest and show differences").
			WithParameter("group", "string", "API group", false).
			WithParameter("version", "string", "API version", true).
			WithParameter("kind", "string", "Resource kind", true).
			WithParameter("name", "string", "Resource name", true).
			WithParameter("namespace", "string", "Namespace name (empty for cluster-scoped)", false).
			WithParameter("manifest", "object", "Desired resource manifest (YAML or JSON)", true).
			WithParameter("diff_format", "string", "Diff format: 'unified' (default), 'json', or 'yaml'", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("resources_validate", "Validate a resource manifest without applying it").
			WithParameter("manifest", "object", "Resource manifest to validate (YAML or JSON)", true).
			WithParameter("schema_version", "string", "Schema version for validation (optional)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("resources_relationships", "Find resource owners and/or dependents").
			WithParameter("group", "string", "API group", false).
			WithParameter("version", "string", "API version", true).
			WithParameter("kind", "string", "Resource kind", true).
			WithParameter("name", "string", "Resource name", true).
			WithParameter("namespace", "string", "Namespace name (empty for cluster-scoped)", false).
			WithParameter("direction", "string", "Direction: 'owners', 'dependents', or 'both' (default: 'both')", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("configmaps_get_data", "Get ConfigMap data").
			WithParameter("name", "string", "ConfigMap name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("keys", "array", "Specific keys to retrieve (optional, returns all if omitted)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("configmaps_set_data", "Update ConfigMap data").
			WithParameter("name", "string", "ConfigMap name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("data", "object", "Data to set (map of string keys to string values)", true).
			WithParameter("merge", "boolean", "If true, merge with existing data; if false, replace (default: false)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("secrets_get_data", "Get Secret data").
			WithParameter("name", "string", "Secret name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("keys", "array", "Specific keys to retrieve (optional, returns all if omitted)", false).
			WithParameter("decode", "boolean", "If true, base64 decode values; if false, return base64 encoded (default: false)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("secrets_set_data", "Update Secret data").
			WithParameter("name", "string", "Secret name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("data", "object", "Data to set (map of string keys to string values)", true).
			WithParameter("merge", "boolean", "If true, merge with existing data; if false, replace (default: false)", false).
			WithParameter("encode", "boolean", "If true, base64 encode provided values; if false, assume already encoded (default: true)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("resources_apply", "Create or update a resource using server-side apply").
			WithParameter("manifest", "object", "Resource manifest (YAML or JSON)", true).
			WithParameter("field_manager", "string", "Field manager name", false).
			WithParameter("dry_run", "boolean", "If true, validate without applying changes", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("resources_patch", "Partially update a resource using JSON Patch, Merge Patch, or Strategic Merge Patch").
			WithParameter("group", "string", "API group", false).
			WithParameter("version", "string", "API version", true).
			WithParameter("kind", "string", "Resource kind", true).
			WithParameter("name", "string", "Resource name", true).
			WithParameter("namespace", "string", "Namespace name (empty for cluster-scoped)", false).
			WithParameter("patch_type", "string", "Patch type: 'merge' (default), 'json', or 'strategic'", false).
			WithParameter("patch_data", "object", "Patch data (object for merge/strategic, array or object with 'op' field for json patch)", true).
			WithParameter("field_manager", "string", "Field manager name", false).
			WithParameter("dry_run", "boolean", "If true, validate without applying changes", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("resources_delete", "Delete a resource").
			WithParameter("group", "string", "API group", false).
			WithParameter("version", "string", "API version", true).
			WithParameter("kind", "string", "Resource kind", true).
			WithParameter("name", "string", "Resource name", true).
			WithParameter("namespace", "string", "Namespace name (empty for cluster-scoped)", false).
			WithParameter("dry_run", "boolean", "If true, validate without deleting", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("resources_scale", "Scale a resource. Omit replicas or set to null for get-only operation").
			WithParameter("group", "string", "API group", false).
			WithParameter("version", "string", "API version", true).
			WithParameter("kind", "string", "Resource kind", true).
			WithParameter("name", "string", "Resource name", true).
			WithParameter("namespace", "string", "Namespace name", true).
			WithParameter("replicas", "integer", "Number of replicas (omit or null for get-only, 0 to scale to zero, >0 to scale to that number)", false).
			WithParameter("dry_run", "boolean", "If true, validate without scaling", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithDestructive().
			Build(),
		mcpHelpers.NewTool("resources_watch", "Watch resources for changes (returns events within timeout)").
			WithParameter("group", "string", "API group", false).
			WithParameter("version", "string", "API version", true).
			WithParameter("kind", "string", "Resource kind", true).
			WithParameter("namespace", "string", "Namespace name (empty for cluster-scoped)", false).
			WithParameter("label_selector", "string", "Label selector", false).
			WithParameter("field_selector", "string", "Field selector", false).
			WithParameter("timeout", "integer", "Timeout in seconds (0 = default 30 seconds)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		// Namespace tools
		mcpHelpers.NewTool("namespaces_list", "List all namespaces").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		// Node tools
		mcpHelpers.NewTool("nodes_top", "Get node resource usage metrics").
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		mcpHelpers.NewTool("nodes_summary", "Get node summary statistics").
			WithParameter("name", "string", "Node name (optional)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
		// Event tools
		mcpHelpers.NewTool("events_list", "List events").
			WithParameter("namespace", "string", "Namespace name (empty for all namespaces)", false).
			WithParameter("context", "string", "Kubernetes context name", false).
			WithReadOnly().
			Build(),
	}
}
