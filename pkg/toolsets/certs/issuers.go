package certs

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// normalizeIssuerSummary normalizes an unstructured Issuer/ClusterIssuer into a summary.
func (t *Toolset) normalizeIssuerSummary(obj *unstructured.Unstructured, kind string) IssuerSummary {
	summary := IssuerSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Kind:      kind,
		Ready:     false,
	}

	// Determine issuer type from spec
	if spec, found, _ := unstructured.NestedMap(obj.Object, "spec"); found {
		// Check for ACME
		if _, found := spec["acme"]; found {
			summary.Type = "ACME"
		} else if _, found := spec["ca"]; found {
			summary.Type = "CA"
		} else if _, found := spec["vault"]; found {
			summary.Type = "Vault"
		} else if _, found := spec["venafi"]; found {
			summary.Type = "Venafi"
		} else if _, found := spec["selfSigned"]; found {
			summary.Type = "SelfSigned"
		}
	}

	// Check status/conditions
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if found {
		if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]interface{}); ok {
					condType, _ := condMap["type"].(string)
					condStatus, _ := condMap["status"].(string)
					condMessage, _ := condMap["message"].(string)

					if condType == "Ready" && condStatus == "True" {
						summary.Ready = true
						summary.Message = condMessage
						break
					} else if condType == "Ready" {
						summary.Message = condMessage
					}
				}
			}
		}
	}

	return summary
}

// handleIssuersList handles the certs.issuers_list tool.
func (t *Toolset) handleIssuersList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
	Raw           bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(IssuerGVK); ok && !t.hasIssuer {
				t.hasIssuer = true
				t.issuerGVR = gvr
			}
			if gvr, ok := t.discovery.GetGVR(ClusterIssuerGVK); ok && !t.hasClusterIssuer {
				t.hasClusterIssuer = true
				t.clusterIssuerGVR = gvr
			}
		}
	}

	if !t.hasIssuer && !t.hasClusterIssuer {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "Issuer or ClusterIssuer CRDs not available",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var issuers []IssuerSummary

	// List Issuers
	if t.hasIssuer {
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
			list, err = clientSet.Dynamic.Resource(t.issuerGVR).Namespace(args.Namespace).List(ctx, listOptions)
		} else {
			list, err = clientSet.Dynamic.Resource(t.issuerGVR).List(ctx, listOptions)
		}

		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Issuers: %w", err)), nil
		}

		for _, item := range list.Items {
			if args.Raw {
				// For raw mode, we'd return the full object, but for now return summary
				issuers = append(issuers, t.normalizeIssuerSummary(&item, "Issuer"))
			} else {
				issuers = append(issuers, t.normalizeIssuerSummary(&item, "Issuer"))
			}
		}

		// Handle continue token
		if list.GetContinue() != "" {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"items":                issuers,
				"continue":             list.GetContinue(),
				"remaining_item_count": list.GetRemainingItemCount(),
			})
			return result, err
		}
	}

	// List ClusterIssuers (cluster-scoped, ignore namespace)
	if t.hasClusterIssuer {
		listOptions := metav1.ListOptions{
			LabelSelector: args.LabelSelector,
		}
		if args.Limit > 0 && len(issuers) < args.Limit {
			listOptions.Limit = int64(args.Limit - len(issuers))
		}
		if args.Continue != "" && len(issuers) == 0 {
			listOptions.Continue = args.Continue
		}

		list, err := clientSet.Dynamic.Resource(t.clusterIssuerGVR).List(ctx, listOptions)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list ClusterIssuers: %w", err)), nil
		}

		for _, item := range list.Items {
			issuers = append(issuers, t.normalizeIssuerSummary(&item, "ClusterIssuer"))
		}

		// Handle continue token
		if list.GetContinue() != "" {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"items":                issuers,
				"continue":             list.GetContinue(),
				"remaining_item_count": list.GetRemainingItemCount(),
			})
			return result, err
		}
	}

	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"items": issuers,
	})
	return result, err
}
