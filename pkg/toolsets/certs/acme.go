package certs

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ACMEChallengeSummary represents a normalized ACME challenge summary.
type ACMEChallengeSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"` // "Challenge" or "Order"
	DNSName   string `json:"dns_name,omitempty"`
	Token     string `json:"token,omitempty"`
	State     string `json:"state,omitempty"`
	Status    string `json:"status,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// normalizeACMEChallengeSummary normalizes an unstructured Challenge or Order.
func (t *Toolset) normalizeACMEChallengeSummary(obj *unstructured.Unstructured, kind string) ACMEChallengeSummary {
	summary := ACMEChallengeSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Kind:      kind,
	}

	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if found {
		if state, found, _ := unstructured.NestedString(status, "state"); found {
			summary.State = state
		}
		if reason, found, _ := unstructured.NestedString(status, "reason"); found {
			summary.Reason = reason
		}

		// For Challenge
		if kind == "Challenge" {
			if token, found, _ := unstructured.NestedString(obj.Object, "spec", "token"); found {
				summary.Token = token
			}
			if dnsName, found, _ := unstructured.NestedString(obj.Object, "spec", "dnsName"); found {
				summary.DNSName = dnsName
			}
		}

		// For Order
		if kind == "Order" {
			if dnsNames, found, _ := unstructured.NestedStringSlice(obj.Object, "spec", "dnsNames"); found && len(dnsNames) > 0 {
				summary.DNSName = dnsNames[0]
			}
		}
	}

	return summary
}

// handleACMEChallengesList handles the certs.acme_challenges_list tool.
func (t *Toolset) handleACMEChallengesList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(ChallengeGVK); ok && !t.hasChallenge {
				t.hasChallenge = true
				t.challengeGVR = gvr
			}
			if gvr, ok := t.discovery.GetGVR(OrderGVK); ok && !t.hasOrder {
				t.hasOrder = true
				t.orderGVR = gvr
			}
		}
	}

	if !t.hasChallenge && !t.hasOrder {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "ACME Challenge or Order CRDs not available",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var challenges []ACMEChallengeSummary

	// List Challenges
	if t.hasChallenge {
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
			list, err = clientSet.Dynamic.Resource(t.challengeGVR).Namespace(args.Namespace).List(ctx, listOptions)
		} else {
			list, err = clientSet.Dynamic.Resource(t.challengeGVR).List(ctx, listOptions)
		}

		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Challenges: %w", err)), nil
		}

		for _, item := range list.Items {
			challenges = append(challenges, t.normalizeACMEChallengeSummary(&item, "Challenge"))
		}

		// Handle continue token
		if list.GetContinue() != "" {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"items":                challenges,
				"continue":             list.GetContinue(),
				"remaining_item_count": list.GetRemainingItemCount(),
			})
			return result, err
		}
	}

	// List Orders
	if t.hasOrder {
		listOptions := metav1.ListOptions{
			LabelSelector: args.LabelSelector,
		}
		if args.Limit > 0 && len(challenges) < args.Limit {
			listOptions.Limit = int64(args.Limit - len(challenges))
		}
		if args.Continue != "" && len(challenges) == 0 {
			listOptions.Continue = args.Continue
		}

		var list *unstructured.UnstructuredList
		if args.Namespace != "" {
			list, err = clientSet.Dynamic.Resource(t.orderGVR).Namespace(args.Namespace).List(ctx, listOptions)
		} else {
			list, err = clientSet.Dynamic.Resource(t.orderGVR).List(ctx, listOptions)
		}

		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Orders: %w", err)), nil
		}

		for _, item := range list.Items {
			challenges = append(challenges, t.normalizeACMEChallengeSummary(&item, "Order"))
		}

		// Handle continue token
		if list.GetContinue() != "" {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"items":                challenges,
				"continue":             list.GetContinue(),
				"remaining_item_count": list.GetRemainingItemCount(),
			})
			return result, err
		}
	}

	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"items": challenges,
	})
	return result, err
}
