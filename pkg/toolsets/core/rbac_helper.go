package core

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// checkRBAC performs an RBAC check before an operation.
// Returns nil if allowed, or an error result if denied.
func (t *Toolset) checkRBAC(ctx context.Context, clientSet *kubernetes.ClientSet, verb string, gvr schema.GroupVersionResource, namespace string) (*mcp.CallToolResult, error) {
	// If RBAC is not required or authorizer is not set, skip check
	if !t.requireRBAC || t.rbacAuthorizer == nil {
		return nil, nil
	}

	// Get current user (for now, use empty string - SelfSubjectAccessReview uses current context)
	// In the future, we could extract user from the request context if available
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
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create RBAC error result: %w", err)), nil
		}
		return result, nil
	}

	return nil, nil
}

// checkRBACGVK performs an RBAC check using a GVK.
func (t *Toolset) checkRBACGVK(ctx context.Context, clientSet *kubernetes.ClientSet, verb string, gvk schema.GroupVersionKind, namespace string) (*mcp.CallToolResult, error) {
	// Map GVK to GVR
	mapping, err := clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		// If mapping fails, skip RBAC check (resource might not exist)
		return nil, nil
	}

	return t.checkRBAC(ctx, clientSet, verb, mapping.Resource, namespace)
}
