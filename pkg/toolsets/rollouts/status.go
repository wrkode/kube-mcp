package rollouts

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handleRolloutGetStatus handles the rollouts.get_status tool.
func (t *Toolset) handleRolloutGetStatus(ctx context.Context, args struct {
	Context   string `json:"context"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Raw       bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var gvr schema.GroupVersionResource
	var hasResource bool

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
					"message": "Argo Rollouts CRD not detected",
				},
			})
			return result, err
		}
		gvr = t.rolloutGVR
		hasResource = true
	case "Canary":
		if !t.hasCanary {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"error": map[string]any{
					"type":    "FeatureNotInstalled",
					"message": "Flagger Canary CRD not detected",
				},
			})
			return result, err
		}
		gvr = t.canaryGVR
		hasResource = true
	default:
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid kind: %s (must be 'Rollout' or 'Canary')", args.Kind)), nil
	}

	if !hasResource {
		return mcpHelpers.NewErrorResult(fmt.Errorf("kind %s not available", args.Kind)), nil
	}

	obj, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get %s: %w", args.Kind, err)), nil
	}

	if args.Raw {
		result, err := mcpHelpers.NewJSONResult(obj.Object)
		return result, err
	}

	// Build status summary
	status := RolloutStatus{
		RolloutSummary: t.normalizeRolloutSummary(obj, RolloutKind(args.Kind)),
		Paused:         false,
	}

	statusObj, found, _ := unstructured.NestedMap(obj.Object, "status")
	if found {
		// Extract detailed status based on kind
		if args.Kind == "Rollout" {
			t.extractArgoRolloutStatus(obj, statusObj, &status)
		} else if args.Kind == "Canary" {
			t.extractFlaggerCanaryStatus(obj, statusObj, &status)
		}

		// Extract conditions
		if conditions, found, _ := unstructured.NestedSlice(statusObj, "conditions"); found {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]interface{}); ok {
					status.Conditions = append(status.Conditions, condMap)
				}
			}
		}
	}

	result, err := mcpHelpers.NewJSONResult(status)
	return result, err
}

// extractArgoRolloutStatus extracts detailed status from Argo Rollout.
func (t *Toolset) extractArgoRolloutStatus(obj *unstructured.Unstructured, statusObj map[string]interface{}, status *RolloutStatus) {
	// Check if paused
	if paused, found, _ := unstructured.NestedBool(obj.Object, "spec", "paused"); found {
		status.Paused = paused
	}

	// Get current step
	if currentStepNumber, found, _ := unstructured.NestedInt64(statusObj, "currentStepIndex"); found {
		step := int(currentStepNumber)
		status.CurrentStep = &step
	}

	// Get total steps from spec
	if steps, found, _ := unstructured.NestedSlice(obj.Object, "spec", "strategy", "canary", "steps"); found {
		total := len(steps)
		status.TotalSteps = &total
	}

	// Get traffic weight
	if weight, found, _ := unstructured.NestedInt64(statusObj, "canary", "weight"); found {
		w := int(weight)
		status.TrafficWeight = &w
	}

	// Get current revision
	if currentPodHash, found, _ := unstructured.NestedString(statusObj, "currentPodHash"); found {
		status.Revision = currentPodHash
	}
}

// extractFlaggerCanaryStatus extracts detailed status from Flagger Canary.
func (t *Toolset) extractFlaggerCanaryStatus(obj *unstructured.Unstructured, statusObj map[string]interface{}, status *RolloutStatus) {
	// Get canary status
	if canaryStatus, found, _ := unstructured.NestedMap(statusObj, "canaryStatus"); found {
		status.AnalysisStatus = canaryStatus
	}

	// Get traffic weight
	if weight, found, _ := unstructured.NestedInt64(statusObj, "canaryWeight"); found {
		w := int(weight)
		status.TrafficWeight = &w
	}
}
