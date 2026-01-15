package rollouts

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// normalizeRolloutSummary normalizes an unstructured rollout into a summary.
func (t *Toolset) normalizeRolloutSummary(obj *unstructured.Unstructured, kind RolloutKind) RolloutSummary {
	summary := RolloutSummary{
		Kind:      string(kind),
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Phase:     RolloutPhaseUnknown,
		Ready:     false,
	}

	switch kind {
	case RolloutKindRollout:
		t.normalizeArgoRollout(obj, &summary)
	case RolloutKindCanary:
		t.normalizeFlaggerCanary(obj, &summary)
	}

	return summary
}

// normalizeArgoRollout normalizes an Argo Rollouts Rollout.
func (t *Toolset) normalizeArgoRollout(obj *unstructured.Unstructured, summary *RolloutSummary) {
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if !found {
		return
	}

	// Get phase
	if phase, found, _ := unstructured.NestedString(status, "phase"); found {
		switch phase {
		case "Progressing":
			summary.Phase = RolloutPhaseProgressing
		case "Paused":
			summary.Phase = RolloutPhasePaused
		case "Degraded":
			summary.Phase = RolloutPhaseDegraded
		case "Healthy":
			summary.Phase = RolloutPhaseHealthy
			summary.Ready = true
		default:
			summary.Phase = RolloutPhaseUnknown
		}
	}

	// Get message from conditions
	if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				if condType, _ := condMap["type"].(string); condType == "Progressing" {
					if msg, _ := condMap["message"].(string); msg != "" {
						summary.Message = msg
					}
					if lastTransitionTime, _ := condMap["lastTransitionTime"].(string); lastTransitionTime != "" {
						summary.LastUpdated = &lastTransitionTime
					}
					break
				}
			}
		}
	}

	// Get current revision
	if currentPodHash, found, _ := unstructured.NestedString(status, "currentPodHash"); found {
		summary.Revision = currentPodHash
	}
}

// normalizeFlaggerCanary normalizes a Flagger Canary.
func (t *Toolset) normalizeFlaggerCanary(obj *unstructured.Unstructured, summary *RolloutSummary) {
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if !found {
		return
	}

	// Get phase
	if phase, found, _ := unstructured.NestedString(status, "phase"); found {
		switch phase {
		case "Progressing", "Initializing":
			summary.Phase = RolloutPhaseProgressing
		case "Waiting":
			summary.Phase = RolloutPhasePaused
		case "Failed":
			summary.Phase = RolloutPhaseDegraded
		case "Succeeded":
			summary.Phase = RolloutPhaseHealthy
			summary.Ready = true
		default:
			summary.Phase = RolloutPhaseUnknown
		}
	}

	// Get message
	if canaryStatus, found, _ := unstructured.NestedMap(status, "canaryStatus"); found {
		if msg, found, _ := unstructured.NestedString(canaryStatus, "message"); found {
			summary.Message = msg
		}
	}

	// Get last update time
	if lastTransitionTime, found, _ := unstructured.NestedString(status, "lastTransitionTime"); found {
		summary.LastUpdated = &lastTransitionTime
	}
}

// handleRolloutsList handles the rollouts.list tool.
func (t *Toolset) handleRolloutsList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var rollouts []RolloutSummary

	// List Argo Rollouts if available
	if t.hasRollout {
		listOptions := metav1.ListOptions{
			LabelSelector: args.LabelSelector,
		}
		if args.Limit > 0 {
			listOptions.Limit = int64(args.Limit)
		}
		if args.Continue != "" {
			listOptions.Continue = args.Continue
		}

		var list *unstructured.UnstructuredList
		var err error
		if args.Namespace != "" {
			list, err = clientSet.Dynamic.Resource(t.rolloutGVR).Namespace(args.Namespace).List(ctx, listOptions)
		} else {
			list, err = clientSet.Dynamic.Resource(t.rolloutGVR).List(ctx, listOptions)
		}

		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Rollouts: %w", err)), nil
		}

		for _, item := range list.Items {
			rollouts = append(rollouts, t.normalizeRolloutSummary(&item, RolloutKindRollout))
		}

		// Handle continue token
		if list.GetContinue() != "" {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"items":                rollouts,
				"continue":             list.GetContinue(),
				"remaining_item_count": list.GetRemainingItemCount(),
			})
			return result, err
		}
	}

	// List Flagger Canaries if available
	if t.hasCanary {
		listOptions := metav1.ListOptions{
			LabelSelector: args.LabelSelector,
		}
		if args.Limit > 0 && len(rollouts) < args.Limit {
			listOptions.Limit = int64(args.Limit - len(rollouts))
		}
		if args.Continue != "" && len(rollouts) == 0 {
			listOptions.Continue = args.Continue
		}

		var list *unstructured.UnstructuredList
		var err error
		if args.Namespace != "" {
			list, err = clientSet.Dynamic.Resource(t.canaryGVR).Namespace(args.Namespace).List(ctx, listOptions)
		} else {
			list, err = clientSet.Dynamic.Resource(t.canaryGVR).List(ctx, listOptions)
		}

		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Canaries: %w", err)), nil
		}

		for _, item := range list.Items {
			rollouts = append(rollouts, t.normalizeRolloutSummary(&item, RolloutKindCanary))
		}

		// Handle continue token
		if list.GetContinue() != "" {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"items":                rollouts,
				"continue":             list.GetContinue(),
				"remaining_item_count": list.GetRemainingItemCount(),
			})
			return result, err
		}
	}

	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"items": rollouts,
	})
	return result, err
}
