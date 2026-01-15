package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// handleResourcesScale handles the resources_scale tool using the scale subresource API.
// If replicas is nil, returns current scale without modifying (get-only operation).
// If replicas is 0, scales the resource to zero replicas.
// If replicas > 0, scales the resource to that number of replicas.
func (t *Toolset) handleResourcesScale(ctx context.Context, args struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  *int   `json:"replicas"` // nil = get-only, 0 = scale to zero, >0 = scale to that number
	DryRun    bool   `json:"dry_run"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	gvk := schema.GroupVersionKind{
		Group:   args.Group,
		Version: args.Version,
		Kind:    args.Kind,
	}

	// Map GVK to GVR
	mapping, err := clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to map GVK to GVR: %w", err)), nil
	}

	gvr := mapping.Resource

	// Check if resource supports scaling by attempting to get the scale subresource
	scaleGVR := schema.GroupVersionResource{
		Group:    gvr.Group,
		Version:  gvr.Version,
		Resource: gvr.Resource,
	}

	// Get current scale
	scaleObj, err := clientSet.Dynamic.Resource(scaleGVR).Namespace(args.Namespace).
		Get(ctx, args.Name, metav1.GetOptions{}, "scale")
	if err != nil {
		// Return an error result (IsError: true) for non-scalable resources
		return mcpHelpers.NewErrorResult(fmt.Errorf("resource %s/%s/%s does not support scaling: %w", args.Group, args.Version, args.Kind, err)), nil
	}

	// Parse current scale from unstructured object
	scaleJSON, err := json.Marshal(scaleObj.Object)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to marshal scale object: %w", err)), nil
	}

	var currentScale autoscalingv1.Scale
	if err := json.Unmarshal(scaleJSON, &currentScale); err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse current scale: %w", err)), nil
	}

	currentReplicas := int(currentScale.Spec.Replicas)

	// If replicas is nil, just return current scale (get-only operation)
	if args.Replicas == nil {
		result, jsonErr := mcpHelpers.NewJSONResult(map[string]any{
			"name":             args.Name,
			"namespace":        args.Namespace,
			"kind":             args.Kind,
			"replicas":         currentReplicas,
			"current_replicas": currentReplicas,
			"desired_replicas": currentReplicas,
			"generation":       currentScale.ObjectMeta.Generation,
			"resource_version": currentScale.ObjectMeta.ResourceVersion,
			"status":           "current",
		})
		if jsonErr != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", jsonErr)), nil
		}
		return result, nil
	}

	// Apply the scale update using PATCH (replicas can be 0 or >0)
	desiredReplicas := *args.Replicas
	scaleBytes, err := json.Marshal(map[string]any{
		"spec": map[string]any{
			"replicas": desiredReplicas,
		},
	})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to marshal scale patch: %w", err)), nil
	}

	// Build patch options with dry-run support
	patchOptions := metav1.PatchOptions{}
	if args.DryRun {
		patchOptions.DryRun = []string{metav1.DryRunAll}
	}

	updatedScale, err := clientSet.Dynamic.Resource(scaleGVR).Namespace(args.Namespace).
		Patch(ctx, args.Name, types.MergePatchType, scaleBytes, patchOptions, "scale")
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to update scale: %w", err)), nil
	}

	// Parse updated scale
	updatedScaleJSON, err := json.Marshal(updatedScale.Object)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to marshal updated scale object: %w", err)), nil
	}

	var newScale autoscalingv1.Scale
	if err := json.Unmarshal(updatedScaleJSON, &newScale); err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse updated scale: %w", err)), nil
	}

	// Use the actual current replicas from the API server response (not the stale pre-patch value)
	actualCurrentReplicas := int(newScale.Spec.Replicas)

	status := "scaled"
	if args.DryRun {
		status = "dry-run-scaled"
	}

	result, jsonErr := mcpHelpers.NewJSONResult(map[string]any{
		"name":             args.Name,
		"namespace":        args.Namespace,
		"kind":             args.Kind,
		"replicas":         desiredReplicas,
		"current_replicas": actualCurrentReplicas,
		"desired_replicas": desiredReplicas,
		"generation":       newScale.ObjectMeta.Generation,
		"resource_version": newScale.ObjectMeta.ResourceVersion,
		"status":           status,
		"dry_run":          args.DryRun,
	})
	if jsonErr != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", jsonErr)), nil
	}
	return result, nil
}
