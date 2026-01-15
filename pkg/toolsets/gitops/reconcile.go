package gitops

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handleAppReconcile handles the gitops.app_reconcile tool.
func (t *Toolset) handleAppReconcile(ctx context.Context, args struct {
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
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to reconcile")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var gvr schema.GroupVersionResource
	var appKind AppKind
	var annotationKey string

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(KustomizationGVK); ok && !t.hasKustomization {
				t.hasKustomization = true
				t.kustomizationGVR = gvr
			}
			if gvr, ok := t.discovery.GetGVR(HelmReleaseGVK); ok && !t.hasHelmRelease {
				t.hasHelmRelease = true
				t.helmReleaseGVR = gvr
			}
		}
	}

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
		annotationKey = FluxKustomizationReconcileAnnotation
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
		annotationKey = FluxHelmReleaseReconcileAnnotation
		if args.Namespace == "" {
			return mcpHelpers.NewErrorResult(fmt.Errorf("namespace is required for HelmRelease")), nil
		}
	case "Application":
		// Argo CD: best-effort refresh trigger
		// Note: Argo CD doesn't have a standard reconcile annotation like Flux.
		// We use the refresh annotation as a best-effort trigger.
		// If no safe annotation exists, return FeatureDisabled for reconcile on Application.
		// For v1, we'll implement reconcile only for Flux and document it clearly.
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureDisabled",
				"message": "Reconcile is not supported for Argo CD Application in CRD-only mode",
				"details": "Argo CD Application reconcile requires Argo CD API access. Use Flux Kustomization or HelmRelease for reconcile operations.",
			},
		})
		return result, err
	default:
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid kind: %s (must be Kustomization or HelmRelease for reconcile)", args.Kind)), nil
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

	// Add reconcile annotation with timestamp
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[annotationKey] = time.Now().Format(time.RFC3339)
	obj.SetAnnotations(annotations)

	// Patch the object
	patched, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to reconcile %s: %w", args.Kind, err)), nil
	}

	// Return refreshed summary
	summary := t.normalizeAppSummary(patched, appKind)
	details := AppDetails{
		AppSummary: summary,
		Conditions: t.normalizeConditions(patched),
	}

	return mcpHelpers.NewJSONResult(map[string]any{
		"result": map[string]any{
			"annotation_applied": annotationKey,
			"timestamp":          time.Now().Format(time.RFC3339),
		},
		"summary": details,
	})
}
