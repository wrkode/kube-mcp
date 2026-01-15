package autoscaling

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// normalizeKEDAScaledObjectSummary normalizes an unstructured ScaledObject into a summary.
func (t *Toolset) normalizeKEDAScaledObjectSummary(obj *unstructured.Unstructured) KEDAScaledObjectSummary {
	summary := KEDAScaledObjectSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Paused:    false,
	}

	// Get target reference
	if scaleTargetRef, found, _ := unstructured.NestedMap(obj.Object, "spec", "scaleTargetRef"); found {
		if kind, _ := scaleTargetRef["kind"].(string); kind != "" {
			summary.TargetKind = kind
		}
		if name, _ := scaleTargetRef["name"].(string); name != "" {
			summary.TargetName = name
		}
	}

	// Get min/max replicas
	if minReplicas, found, _ := unstructured.NestedInt64(obj.Object, "spec", "minReplicaCount"); found {
		min := int32(minReplicas)
		summary.MinReplicas = &min
	}
	if maxReplicas, found, _ := unstructured.NestedInt64(obj.Object, "spec", "maxReplicaCount"); found {
		max := int32(maxReplicas)
		summary.MaxReplicas = &max
	}

	// Check if paused (annotation)
	if annotations := obj.GetAnnotations(); annotations != nil {
		if paused, found := annotations[KEDAPauseAnnotation]; found && paused == "true" {
			summary.Paused = true
		}
	}

	// Get triggers
	if triggers, found, _ := unstructured.NestedSlice(obj.Object, "spec", "triggers"); found {
		for _, trigger := range triggers {
			if triggerMap, ok := trigger.(map[string]interface{}); ok {
				if triggerType, _ := triggerMap["type"].(string); triggerType != "" {
					summary.Triggers = append(summary.Triggers, triggerType)
				}
			}
		}
	}

	// Get status
	if status, found, _ := unstructured.NestedMap(obj.Object, "status"); found {
		if replicas, found, _ := unstructured.NestedInt64(status, "replicas"); found {
			current := int32(replicas)
			summary.CurrentReplicas = &current
		}
	}

	return summary
}

// handleKEDAScaledObjectsList handles the autoscaling.keda_scaledobjects_list tool.
func (t *Toolset) handleKEDAScaledObjectsList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(ScaledObjectGVK); ok && !t.hasKEDA {
				t.hasKEDA = true
				t.scaledObjectGVR = gvr
			}
		}
	}

	if !t.hasKEDA {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "KEDA ScaledObject CRD not available",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

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
	if args.Namespace != "" {
		list, err = clientSet.Dynamic.Resource(t.scaledObjectGVR).Namespace(args.Namespace).List(ctx, listOptions)
	} else {
		list, err = clientSet.Dynamic.Resource(t.scaledObjectGVR).List(ctx, listOptions)
	}

	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list ScaledObjects: %w", err)), nil
	}

	var scaledObjects []KEDAScaledObjectSummary
	for _, item := range list.Items {
		scaledObjects = append(scaledObjects, t.normalizeKEDAScaledObjectSummary(&item))
	}

	resultData := map[string]any{
		"items": scaledObjects,
	}
	if list.GetContinue() != "" {
		resultData["continue"] = list.GetContinue()
		if list.GetRemainingItemCount() != nil {
			resultData["remaining_item_count"] = *list.GetRemainingItemCount()
		}
	}

	result, err := mcpHelpers.NewJSONResult(resultData)
	return result, err
}

// handleKEDAScaledObjectGet handles the autoscaling.keda_scaledobject_get tool.
func (t *Toolset) handleKEDAScaledObjectGet(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Raw       bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(ScaledObjectGVK); ok && !t.hasKEDA {
				t.hasKEDA = true
				t.scaledObjectGVR = gvr
			}
		}
	}

	if !t.hasKEDA {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "KEDA ScaledObject CRD not available",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	obj, err := clientSet.Dynamic.Resource(t.scaledObjectGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get ScaledObject: %w", err)), nil
	}

	if args.Raw {
		result, err := mcpHelpers.NewJSONResult(obj.Object)
		return result, err
	}

	summary := t.normalizeKEDAScaledObjectSummary(obj)
	result, err := mcpHelpers.NewJSONResult(summary)
	return result, err
}

// handleKEDATriggersExplain handles the autoscaling.keda_triggers_explain tool.
func (t *Toolset) handleKEDATriggersExplain(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}) (*mcp.CallToolResult, error) {
	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(ScaledObjectGVK); ok && !t.hasKEDA {
				t.hasKEDA = true
				t.scaledObjectGVR = gvr
			}
		}
	}

	if !t.hasKEDA {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "KEDA ScaledObject CRD not available",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	obj, err := clientSet.Dynamic.Resource(t.scaledObjectGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get ScaledObject: %w", err)), nil
	}

	var triggers []KEDATriggerSummary

	if triggerSpecs, found, _ := unstructured.NestedSlice(obj.Object, "spec", "triggers"); found {
		for _, triggerSpec := range triggerSpecs {
			if triggerMap, ok := triggerSpec.(map[string]interface{}); ok {
				trigger := KEDATriggerSummary{}
				if triggerType, _ := triggerMap["type"].(string); triggerType != "" {
					trigger.Type = triggerType
				}
				if metadata, found, _ := unstructured.NestedMap(triggerMap, "metadata"); found {
					trigger.Metadata = metadata
				}
				triggers = append(triggers, trigger)
			}
		}
	}

	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"triggers": triggers,
	})
	return result, err
}

// handleKEDAPause handles the autoscaling.keda_pause tool.
func (t *Toolset) handleKEDAPause(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(ScaledObjectGVK); ok && !t.hasKEDA {
				t.hasKEDA = true
				t.scaledObjectGVR = gvr
			}
		}
	}

	if !t.hasKEDA {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "KEDA ScaledObject CRD not available",
			},
		})
		return result, err
	}

	if !args.Confirm {
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to pause")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// RBAC check
	if errResult, err := t.checkRBAC(ctx, clientSet, "update", t.scaledObjectGVR, args.Namespace); errResult != nil || err != nil {
		return errResult, err
	}

	// Get current object
	obj, err := clientSet.Dynamic.Resource(t.scaledObjectGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get ScaledObject: %w", err)), nil
	}

	// Add pause annotation
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[KEDAPauseAnnotation] = "true"
	obj.SetAnnotations(annotations)

	// Update the object
	patched, err := clientSet.Dynamic.Resource(t.scaledObjectGVR).Namespace(args.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to pause ScaledObject: %w", err)), nil
	}

	summary := t.normalizeKEDAScaledObjectSummary(patched)
	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"result": map[string]any{
			"annotation_applied": KEDAPauseAnnotation,
		},
		"summary": summary,
	})
	return result, err
}

// handleKEDAResume handles the autoscaling.keda_resume tool.
func (t *Toolset) handleKEDAResume(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(ScaledObjectGVK); ok && !t.hasKEDA {
				t.hasKEDA = true
				t.scaledObjectGVR = gvr
			}
		}
	}

	if !t.hasKEDA {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "KEDA ScaledObject CRD not available",
			},
		})
		return result, err
	}

	if !args.Confirm {
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to resume")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// RBAC check
	if errResult, err := t.checkRBAC(ctx, clientSet, "update", t.scaledObjectGVR, args.Namespace); errResult != nil || err != nil {
		return errResult, err
	}

	// Get current object
	obj, err := clientSet.Dynamic.Resource(t.scaledObjectGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get ScaledObject: %w", err)), nil
	}

	// Remove pause annotation
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	delete(annotations, KEDAPauseAnnotation)
	obj.SetAnnotations(annotations)

	// Update the object
	patched, err := clientSet.Dynamic.Resource(t.scaledObjectGVR).Namespace(args.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to resume ScaledObject: %w", err)), nil
	}

	summary := t.normalizeKEDAScaledObjectSummary(patched)
	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"result": map[string]any{
			"annotation_removed": KEDAPauseAnnotation,
		},
		"summary": summary,
	})
	return result, err
}
