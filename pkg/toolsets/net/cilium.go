package net

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// normalizeCiliumPolicySummary normalizes an unstructured Cilium policy into a summary.
func (t *Toolset) normalizeCiliumPolicySummary(obj *unstructured.Unstructured, kind string) CiliumPolicySummary {
	summary := CiliumPolicySummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Kind:      kind,
	}

	// Extract endpoint selectors (best-effort)
	if spec, found, _ := unstructured.NestedMap(obj.Object, "spec"); found {
		if endpointSelector, found, _ := unstructured.NestedMap(spec, "endpointSelector"); found {
			if matchLabels, found, _ := unstructured.NestedMap(endpointSelector, "matchLabels"); found {
				for k, v := range matchLabels {
					if vStr, ok := v.(string); ok {
						summary.Endpoints = append(summary.Endpoints, fmt.Sprintf("%s=%s", k, vStr))
					}
				}
			}
		}
	}

	return summary
}

// handleCiliumPoliciesList handles the net.cilium_policies_list tool.
func (t *Toolset) handleCiliumPoliciesList(ctx context.Context, args struct {
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
			if gvr, ok := t.discovery.GetGVR(CiliumNetworkPolicyGVK); ok && t.ciliumNetworkPolicyGVR.Resource == "" {
				t.hasCilium = true
				t.ciliumNetworkPolicyGVR = gvr
			}
			if gvr, ok := t.discovery.GetGVR(CiliumClusterwideNetworkPolicyGVK); ok && t.ciliumClusterwideNetworkPolicyGVR.Resource == "" {
				t.hasCilium = true
				t.ciliumClusterwideNetworkPolicyGVR = gvr
			}
		}
	}

	if !t.hasCilium {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "Cilium CRDs not available",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var policies []CiliumPolicySummary

	// List CiliumNetworkPolicy
	if t.ciliumNetworkPolicyGVR.Resource != "" {
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
			list, err = clientSet.Dynamic.Resource(t.ciliumNetworkPolicyGVR).Namespace(args.Namespace).List(ctx, listOptions)
		} else {
			list, err = clientSet.Dynamic.Resource(t.ciliumNetworkPolicyGVR).List(ctx, listOptions)
		}

		if err == nil {
			for _, item := range list.Items {
				policies = append(policies, t.normalizeCiliumPolicySummary(&item, "CiliumNetworkPolicy"))
			}

			if list.GetContinue() != "" {
				result, err := mcpHelpers.NewJSONResult(map[string]any{
					"items":                policies,
					"continue":             list.GetContinue(),
					"remaining_item_count": list.GetRemainingItemCount(),
				})
				return result, err
			}
		}
	}

	// List CiliumClusterwideNetworkPolicy
	if t.ciliumClusterwideNetworkPolicyGVR.Resource != "" {
		listOptions := metav1.ListOptions{
			LabelSelector: args.LabelSelector,
		}
		if args.Limit > 0 && len(policies) < args.Limit {
			listOptions.Limit = int64(args.Limit - len(policies))
		}

		list, err := clientSet.Dynamic.Resource(t.ciliumClusterwideNetworkPolicyGVR).List(ctx, listOptions)
		if err == nil {
			for _, item := range list.Items {
				policies = append(policies, t.normalizeCiliumPolicySummary(&item, "CiliumClusterwideNetworkPolicy"))
			}
		}
	}

	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"items": policies,
	})
	return result, err
}

// handleCiliumPolicyGet handles the net.cilium_policy_get tool.
func (t *Toolset) handleCiliumPolicyGet(ctx context.Context, args struct {
	Context   string `json:"context"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Raw       bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(CiliumNetworkPolicyGVK); ok && t.ciliumNetworkPolicyGVR.Resource == "" {
				t.hasCilium = true
				t.ciliumNetworkPolicyGVR = gvr
			}
			if gvr, ok := t.discovery.GetGVR(CiliumClusterwideNetworkPolicyGVK); ok && t.ciliumClusterwideNetworkPolicyGVR.Resource == "" {
				t.hasCilium = true
				t.ciliumClusterwideNetworkPolicyGVR = gvr
			}
		}
	}

	if !t.hasCilium {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "Cilium CRDs not available",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var gvr schema.GroupVersionResource
	if args.Kind == "CiliumNetworkPolicy" {
		gvr = t.ciliumNetworkPolicyGVR
	} else if args.Kind == "CiliumClusterwideNetworkPolicy" {
		gvr = t.ciliumClusterwideNetworkPolicyGVR
	} else {
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid kind: %s", args.Kind)), nil
	}

	var obj *unstructured.Unstructured
	if args.Kind == "CiliumClusterwideNetworkPolicy" {
		obj, err = clientSet.Dynamic.Resource(gvr).Get(ctx, args.Name, metav1.GetOptions{})
	} else {
		obj, err = clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	}

	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get %s: %w", args.Kind, err)), nil
	}

	if args.Raw {
		result, err := mcpHelpers.NewJSONResult(obj.Object)
		return result, err
	}

	summary := t.normalizeCiliumPolicySummary(obj, args.Kind)
	result, err := mcpHelpers.NewJSONResult(summary)
	return result, err
}
