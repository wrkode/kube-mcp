package capi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Toolset implements the CAPI toolset for Cluster API.
type Toolset struct {
	provider             kubernetes.ClientProvider
	discovery            *kubernetes.CRDDiscovery
	logger               *observability.Logger
	metrics              *observability.Metrics
	rbacAuthorizer       kubernetes.RBACAuthorizer
	requireRBAC          bool
	enabled              bool
	hasCluster           bool
	hasMachine           bool
	hasMachineDeployment bool
	clusterGVR           schema.GroupVersionResource
	machineGVR           schema.GroupVersionResource
	machineDeploymentGVR schema.GroupVersionResource
}

// NewToolset creates a new CAPI toolset with CRD detection.
func NewToolset(provider kubernetes.ClientProvider, discovery *kubernetes.CRDDiscovery) *Toolset {
	enabled := false
	hasCluster := false
	hasMachine := false
	hasMachineDeployment := false
	var clusterGVR, machineGVR, machineDeploymentGVR schema.GroupVersionResource

	if discovery != nil {
		// Check for Cluster CRD (required)
		if gvr, ok := discovery.GetGVR(ClusterGVK); ok {
			enabled = true
			hasCluster = true
			clusterGVR = gvr
		}

		// Check for Machine CRD (optional but enhances functionality)
		if gvr, ok := discovery.GetGVR(MachineGVK); ok {
			hasMachine = true
			machineGVR = gvr
		}

		// Check for MachineDeployment CRD (optional but enhances functionality)
		if gvr, ok := discovery.GetGVR(MachineDeploymentGVK); ok {
			hasMachineDeployment = true
			machineDeploymentGVR = gvr
		}
	}

	return &Toolset{
		provider:             provider,
		discovery:            discovery,
		enabled:              enabled,
		hasCluster:           hasCluster,
		hasMachine:           hasMachine,
		hasMachineDeployment: hasMachineDeployment,
		clusterGVR:           clusterGVR,
		machineGVR:           machineGVR,
		machineDeploymentGVR: machineDeploymentGVR,
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

// IsEnabled returns whether the CAPI toolset is enabled.
func (t *Toolset) IsEnabled() bool {
	return t.enabled
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "capi"
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

// wrapToolHandler wraps a tool handler with observability.
func (t *Toolset) wrapToolHandler(
	toolName string,
	handler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error),
	getCluster func(args any) string,
) func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
	if t.logger == nil || t.metrics == nil {
		return handler
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		start := time.Now()
		cluster := getCluster(args)
		if cluster == "" {
			cluster = "default"
		}

		defer func() {
			if r := recover(); r != nil {
				t.logger.Error(ctx, "Panic in tool handler",
					"tool", toolName,
					"panic", r,
					"cluster", cluster,
				)
			}
		}()

		result, out, err := handler(ctx, req, args)

		duration := time.Since(start)
		t.logger.LogToolInvocation(ctx, toolName, cluster, duration, err)
		success := err == nil && (result == nil || !result.IsError)
		t.metrics.RecordToolCall(toolName, cluster, success, duration.Seconds())

		return result, out, err
	}
}

// checkFeatureEnabled checks if the toolset is enabled and returns an error if not.
func (t *Toolset) checkFeatureEnabled() (*mcp.CallToolResult, error) {
	if !t.enabled {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "CAPI toolset is not enabled",
				"details": "Required CAPI CRD (cluster.x-k8s.io/v1beta1/Cluster) not detected in cluster",
			},
		})
		return result, err
	}
	return nil, nil
}

// checkRBAC performs an RBAC check before an operation.
func (t *Toolset) checkRBAC(ctx context.Context, clientSet *kubernetes.ClientSet, verb string, gvr schema.GroupVersionResource, namespace string) (*mcp.CallToolResult, error) {
	if !t.requireRBAC || t.rbacAuthorizer == nil {
		return nil, nil
	}

	user := ""
	allowed, err := t.rbacAuthorizer.Allowed(ctx, user, verb, gvr, namespace)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to check RBAC: %w", err)), nil
	}

	if !allowed {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"code":    "KubernetesError",
				"message": fmt.Sprintf("Forbidden: user does not have permission to %s %s/%s in namespace %s", verb, gvr.Group, gvr.Resource, namespace),
				"details": map[string]any{
					"verb":      verb,
					"group":     gvr.Group,
					"resource":  gvr.Resource,
					"namespace": namespace,
					"reason":    "Forbidden",
				},
			},
		})
		return result, err
	}

	return nil, nil
}
