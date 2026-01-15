package rollouts

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handleRolloutPromote handles the rollouts.promote tool.
func (t *Toolset) handleRolloutPromote(ctx context.Context, args struct {
	Context   string `json:"context"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	if !args.Confirm {
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to promote")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var gvr schema.GroupVersionResource
	var annotationKey string

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(RolloutGVK); ok && !t.hasRollout {
				t.hasRollout = true
				t.rolloutGVR = gvr
			}
			if gvr, ok := t.discovery.GetGVR(CanaryGVK); ok && !t.hasCanary {
				t.hasCanary = true
				t.canaryGVR = gvr
			}
		}
	}

	switch args.Kind {
	case "Rollout":
		if !t.hasRollout {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"error": map[string]any{
					"type":    "FeatureNotInstalled",
					"message": "Argo Rollouts CRD not available",
				},
			})
			return result, err
		}
		gvr = t.rolloutGVR
		annotationKey = ArgoRolloutsPromoteAnnotation
	case "Canary":
		// Flagger promote: return FeatureDisabled for now (can be implemented if safe mechanism exists)
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureDisabled",
				"message": "Promote is not yet supported for Flagger Canary",
				"details": "Flagger Canary promotion requires controller-specific mechanisms. Use Argo Rollouts for promote operations.",
			},
		})
		return result, err
	default:
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid kind: %s (must be 'Rollout' or 'Canary')", args.Kind)), nil
	}

	// RBAC check
	if errResult, err := t.checkRBAC(ctx, clientSet, "update", gvr, args.Namespace); errResult != nil || err != nil {
		return errResult, err
	}

	// Get current object
	obj, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get %s: %w", args.Kind, err)), nil
	}

	// Add promote annotation
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[annotationKey] = time.Now().Format(time.RFC3339)
	obj.SetAnnotations(annotations)

	// Update the object
	patched, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to promote %s: %w", args.Kind, err)), nil
	}

	// Return updated status
	status := t.normalizeRolloutSummary(patched, RolloutKindRollout)
	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"result": map[string]any{
			"annotation_applied": annotationKey,
			"timestamp":          time.Now().Format(time.RFC3339),
		},
		"summary": status,
	})
	return result, err
}

// handleRolloutAbort handles the rollouts.abort tool.
func (t *Toolset) handleRolloutAbort(ctx context.Context, args struct {
	Context   string `json:"context"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	if !args.Confirm {
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to abort")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var gvr schema.GroupVersionResource
	var annotationKey string

	switch args.Kind {
	case "Rollout":
		if !t.hasRollout {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"error": map[string]any{
					"type":    "FeatureNotInstalled",
					"message": "Argo Rollouts CRD not available",
				},
			})
			return result, err
		}
		gvr = t.rolloutGVR
		annotationKey = ArgoRolloutsAbortAnnotation
	case "Canary":
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureDisabled",
				"message": "Abort is not yet supported for Flagger Canary",
			},
		})
		return result, err
	default:
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid kind: %s", args.Kind)), nil
	}

	// RBAC check
	if errResult, err := t.checkRBAC(ctx, clientSet, "update", gvr, args.Namespace); errResult != nil || err != nil {
		return errResult, err
	}

	// Get and update object
	obj, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get %s: %w", args.Kind, err)), nil
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[annotationKey] = time.Now().Format(time.RFC3339)
	obj.SetAnnotations(annotations)

	patched, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to abort %s: %w", args.Kind, err)), nil
	}

	status := t.normalizeRolloutSummary(patched, RolloutKindRollout)
	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"result": map[string]any{
			"annotation_applied": annotationKey,
			"timestamp":          time.Now().Format(time.RFC3339),
		},
		"summary": status,
	})
	return result, err
}

// handleRolloutRetry handles the rollouts.retry tool.
func (t *Toolset) handleRolloutRetry(ctx context.Context, args struct {
	Context   string `json:"context"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	if !args.Confirm {
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to retry")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var gvr schema.GroupVersionResource
	var annotationKey string

	switch args.Kind {
	case "Rollout":
		if !t.hasRollout {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"error": map[string]any{
					"type":    "FeatureNotInstalled",
					"message": "Argo Rollouts CRD not available",
				},
			})
			return result, err
		}
		gvr = t.rolloutGVR
		annotationKey = ArgoRolloutsRetryAnnotation
	case "Canary":
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureDisabled",
				"message": "Retry is not yet supported for Flagger Canary",
			},
		})
		return result, err
	default:
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid kind: %s", args.Kind)), nil
	}

	// RBAC check
	if errResult, err := t.checkRBAC(ctx, clientSet, "update", gvr, args.Namespace); errResult != nil || err != nil {
		return errResult, err
	}

	// Get and update object
	obj, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get %s: %w", args.Kind, err)), nil
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[annotationKey] = time.Now().Format(time.RFC3339)
	obj.SetAnnotations(annotations)

	patched, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to retry %s: %w", args.Kind, err)), nil
	}

	status := t.normalizeRolloutSummary(patched, RolloutKindRollout)
	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"result": map[string]any{
			"annotation_applied": annotationKey,
			"timestamp":          time.Now().Format(time.RFC3339),
		},
		"summary": status,
	})
	return result, err
}
