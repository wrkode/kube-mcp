package gitops

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// normalizeAppSummary normalizes an unstructured GitOps app into a summary.
func (t *Toolset) normalizeAppSummary(obj *unstructured.Unstructured, kind AppKind) AppSummary {
	summary := AppSummary{
		Kind:      string(kind),
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Ready:     false,
		Status:    AppStatusUnknown,
	}

	// Extract status based on kind
	switch kind {
	case AppKindKustomization:
		t.normalizeFluxKustomization(obj, &summary)
	case AppKindHelmRelease:
		t.normalizeFluxHelmRelease(obj, &summary)
	case AppKindApplication:
		t.normalizeArgoApplication(obj, &summary)
	}

	return summary
}

// normalizeFluxKustomization normalizes a Flux Kustomization.
func (t *Toolset) normalizeFluxKustomization(obj *unstructured.Unstructured, summary *AppSummary) {
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if !found {
		return
	}

	// Check Ready condition
	if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				condType, _ := condMap["type"].(string)
				condStatus, _ := condMap["status"].(string)
				condMessage, _ := condMap["message"].(string)

				if condType == "Ready" {
					summary.Ready = condStatus == "True"
					if condStatus == "True" {
						summary.Status = AppStatusReady
					} else {
						summary.Status = AppStatusDegraded
					}
					summary.Message = condMessage
				} else if condType == "ReconciliationSucceeded" && condStatus == "True" {
					if summary.Status == AppStatusUnknown {
						summary.Status = AppStatusProgressing
					}
				}
			}
		}
	}

	// Get last applied revision
	if lastAppliedRevision, found, _ := unstructured.NestedString(status, "lastAppliedRevision"); found {
		summary.Revision = lastAppliedRevision
	}

	// Get last update time from conditions
	if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				if lastTransitionTime, found, _ := unstructured.NestedString(condMap, "lastTransitionTime"); found {
					summary.LastUpdated = &lastTransitionTime
					break
				}
			}
		}
	}
}

// normalizeFluxHelmRelease normalizes a Flux HelmRelease.
func (t *Toolset) normalizeFluxHelmRelease(obj *unstructured.Unstructured, summary *AppSummary) {
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if !found {
		return
	}

	// Check Ready condition
	if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				condType, _ := condMap["type"].(string)
				condStatus, _ := condMap["status"].(string)
				condMessage, _ := condMap["message"].(string)

				if condType == "Ready" {
					summary.Ready = condStatus == "True"
					if condStatus == "True" {
						summary.Status = AppStatusReady
					} else {
						summary.Status = AppStatusDegraded
					}
					summary.Message = condMessage
				}
			}
		}
	}

	// Get last applied revision
	if lastAppliedRevision, found, _ := unstructured.NestedString(status, "lastAppliedRevision"); found {
		summary.Revision = lastAppliedRevision
	}

	// Get artifact revision
	if artifact, found, _ := unstructured.NestedMap(status, "artifact"); found {
		if revision, found, _ := unstructured.NestedString(artifact, "revision"); found {
			summary.Artifact = revision
		}
	}

	// Get last update time
	if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				if lastTransitionTime, found, _ := unstructured.NestedString(condMap, "lastTransitionTime"); found {
					summary.LastUpdated = &lastTransitionTime
					break
				}
			}
		}
	}
}

// normalizeArgoApplication normalizes an Argo CD Application.
func (t *Toolset) normalizeArgoApplication(obj *unstructured.Unstructured, summary *AppSummary) {
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if !found {
		return
	}

	// Check health status
	if health, found, _ := unstructured.NestedMap(status, "health"); found {
		if healthStatus, found, _ := unstructured.NestedString(health, "status"); found {
			switch healthStatus {
			case "Healthy":
				summary.Status = AppStatusReady
				summary.Ready = true
			case "Progressing":
				summary.Status = AppStatusProgressing
			case "Degraded", "Suspended":
				summary.Status = AppStatusDegraded
			default:
				summary.Status = AppStatusUnknown
			}
		}
		if healthMessage, found, _ := unstructured.NestedString(health, "message"); found {
			summary.Message = healthMessage
		}
	}

	// Check sync status
	if sync, found, _ := unstructured.NestedMap(status, "sync"); found {
		if syncStatus, found, _ := unstructured.NestedString(sync, "status"); found {
			if syncStatus == "Synced" && summary.Status == AppStatusUnknown {
				summary.Status = AppStatusReady
				summary.Ready = true
			}
		}
		if syncRevision, found, _ := unstructured.NestedString(sync, "revision"); found {
			summary.Revision = syncRevision
		}
	}

	// Get last update time
	if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				if lastTransitionTime, found, _ := unstructured.NestedString(condMap, "lastTransitionTime"); found {
					summary.LastUpdated = &lastTransitionTime
					break
				}
			}
		}
	}
}

// normalizeConditions normalizes conditions from an unstructured object.
func (t *Toolset) normalizeConditions(obj *unstructured.Unstructured) []AppCondition {
	var conditions []AppCondition

	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if !found {
		return conditions
	}

	if conds, found, _ := unstructured.NestedSlice(status, "conditions"); found {
		for _, cond := range conds {
			if condMap, ok := cond.(map[string]interface{}); ok {
				condition := AppCondition{}
				if typ, ok := condMap["type"].(string); ok {
					condition.Type = typ
				}
				if status, ok := condMap["status"].(string); ok {
					condition.Status = status
				}
				if reason, ok := condMap["reason"].(string); ok {
					condition.Reason = reason
				}
				if message, ok := condMap["message"].(string); ok {
					condition.Message = message
				}
				if lastTransitionTime, ok := condMap["lastTransitionTime"].(string); ok {
					condition.Time = lastTransitionTime
				}
				conditions = append(conditions, condition)
			}
		}
	}

	return conditions
}

// handleAppsList handles the gitops.apps_list tool.
func (t *Toolset) handleAppsList(ctx context.Context, args struct {
	Context       string   `json:"context"`
	Namespace     string   `json:"namespace"`
	LabelSelector string   `json:"label_selector"`
	Kinds         []string `json:"kinds"`
	Limit         int      `json:"limit"`
	Continue      string   `json:"continue"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var apps []AppSummary

	// Determine which kinds to query
	kindsToQuery := make(map[AppKind]bool)
	if len(args.Kinds) == 0 {
		// Query all available kinds
		if t.hasKustomization {
			kindsToQuery[AppKindKustomization] = true
		}
		if t.hasHelmRelease {
			kindsToQuery[AppKindHelmRelease] = true
		}
		if t.hasApplication {
			kindsToQuery[AppKindApplication] = true
		}
	} else {
		for _, kindStr := range args.Kinds {
			switch kindStr {
			case "Kustomization":
				if t.hasKustomization {
					kindsToQuery[AppKindKustomization] = true
				}
			case "HelmRelease":
				if t.hasHelmRelease {
					kindsToQuery[AppKindHelmRelease] = true
				}
			case "Application":
				if t.hasApplication {
					kindsToQuery[AppKindApplication] = true
				}
			}
		}
	}

	// Query each kind
	listOpts := metav1.ListOptions{}
	if args.LabelSelector != "" {
		listOpts.LabelSelector = args.LabelSelector
	}
	if args.Limit > 0 {
		listOpts.Limit = int64(args.Limit)
	}
	if args.Continue != "" {
		listOpts.Continue = args.Continue
	}

	if kindsToQuery[AppKindKustomization] {
		list, err := clientSet.Dynamic.Resource(t.kustomizationGVR).Namespace(args.Namespace).List(ctx, listOpts)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Kustomizations: %w", err)), nil
		}
		for _, item := range list.Items {
			apps = append(apps, t.normalizeAppSummary(&item, AppKindKustomization))
		}
	}

	if kindsToQuery[AppKindHelmRelease] {
		list, err := clientSet.Dynamic.Resource(t.helmReleaseGVR).Namespace(args.Namespace).List(ctx, listOpts)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list HelmReleases: %w", err)), nil
		}
		for _, item := range list.Items {
			apps = append(apps, t.normalizeAppSummary(&item, AppKindHelmRelease))
		}
	}

	if kindsToQuery[AppKindApplication] {
		list, err := clientSet.Dynamic.Resource(t.applicationGVR).Namespace(args.Namespace).List(ctx, listOpts)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Applications: %w", err)), nil
		}
		for _, item := range list.Items {
			apps = append(apps, t.normalizeAppSummary(&item, AppKindApplication))
		}
		// Track continue token from last list operation
		if list.GetContinue() != "" {
			listOpts.Continue = list.GetContinue()
		}
	}

	result := map[string]any{"items": apps}
	if listOpts.Continue != "" {
		result["continue"] = listOpts.Continue
	}
	return mcpHelpers.NewJSONResult(result)
}

// handleAppGet handles the gitops.app_get tool.
func (t *Toolset) handleAppGet(ctx context.Context, args struct {
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
	var appKind AppKind

	switch args.Kind {
	case "Kustomization":
		if !t.hasKustomization {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"error": map[string]any{
					"type":    "FeatureNotInstalled",
					"message": "Kustomization CRD not available",
					"details": "kustomize.toolkit.fluxcd.io/v1/Kustomization CRD not detected in cluster",
				},
			})
			return result, err
		}
		gvr = t.kustomizationGVR
		appKind = AppKindKustomization
		if args.Namespace == "" {
			return mcpHelpers.NewErrorResult(fmt.Errorf("namespace is required for Kustomization")), nil
		}
	case "HelmRelease":
		if !t.hasHelmRelease {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"error": map[string]any{
					"type":    "FeatureNotInstalled",
					"message": "HelmRelease CRD not available",
					"details": "helm.toolkit.fluxcd.io/v2/HelmRelease CRD not detected in cluster",
				},
			})
			return result, err
		}
		gvr = t.helmReleaseGVR
		appKind = AppKindHelmRelease
		if args.Namespace == "" {
			return mcpHelpers.NewErrorResult(fmt.Errorf("namespace is required for HelmRelease")), nil
		}
	case "Application":
		if !t.hasApplication {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"error": map[string]any{
					"type":    "FeatureNotInstalled",
					"message": "Application CRD not available",
					"details": "argoproj.io/v1alpha1/Application CRD not detected in cluster",
				},
			})
			return result, err
		}
		gvr = t.applicationGVR
		appKind = AppKindApplication
		// Application can be cluster-scoped or namespaced, but namespace is typically required
		if args.Namespace == "" {
			return mcpHelpers.NewErrorResult(fmt.Errorf("namespace is required for Application")), nil
		}
	default:
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid kind: %s (must be Kustomization, HelmRelease, or Application)", args.Kind)), nil
	}

	obj, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get %s: %w", args.Kind, err)), nil
	}

	summary := t.normalizeAppSummary(obj, appKind)
	details := AppDetails{
		AppSummary: summary,
		Conditions: t.normalizeConditions(obj),
	}

	result := map[string]any{
		"summary": details,
	}

	if args.Raw {
		result["raw_object"] = obj.Object
	}

	return mcpHelpers.NewJSONResult(result)
}
