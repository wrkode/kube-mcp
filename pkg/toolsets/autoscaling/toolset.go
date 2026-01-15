package autoscaling

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

// Toolset implements the Autoscaling toolset for HPA and KEDA.
type Toolset struct {
	provider        kubernetes.ClientProvider
	discovery       *kubernetes.CRDDiscovery
	logger          *observability.Logger
	metrics         *observability.Metrics
	rbacAuthorizer  kubernetes.RBACAuthorizer
	requireRBAC     bool
	enabled         bool // Always true (HPA is native)
	hasKEDA         bool
	scaledObjectGVR schema.GroupVersionResource
	scaledJobGVR    schema.GroupVersionResource
}

// NewToolset creates a new Autoscaling toolset with CRD detection.
func NewToolset(provider kubernetes.ClientProvider, discovery *kubernetes.CRDDiscovery) *Toolset {
	enabled := true // HPA is always available
	hasKEDA := false
	var scaledObjectGVR, scaledJobGVR schema.GroupVersionResource

	if discovery != nil {
		// Check for KEDA ScaledObject
		if gvr, ok := discovery.GetGVR(ScaledObjectGVK); ok {
			hasKEDA = true
			scaledObjectGVR = gvr
		}
		// Check for KEDA ScaledJob (optional)
		if gvr, ok := discovery.GetGVR(ScaledJobGVK); ok {
			scaledJobGVR = gvr
		}
	}

	return &Toolset{
		provider:        provider,
		discovery:       discovery,
		enabled:         enabled,
		hasKEDA:         hasKEDA,
		scaledObjectGVR: scaledObjectGVR,
		scaledJobGVR:    scaledJobGVR,
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

// IsEnabled returns whether the Autoscaling toolset is enabled.
func (t *Toolset) IsEnabled() bool {
	return t.enabled
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "autoscaling"
}

// unmarshalArgs unmarshals args from map[string]interface{} to the target struct type.
func unmarshalArgs[T any](args any) (T, error) {
	var result T
	if args == nil {
		return result, nil
	}

	if typed, ok := args.(T); ok {
		return typed, nil
	}

	if argsMap, ok := args.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(argsMap)
		if err != nil {
			return result, err
		}
		err = json.Unmarshal(jsonData, &result)
		return result, err
	}

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
