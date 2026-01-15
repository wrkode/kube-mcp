package net

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// normalizeNetworkPolicyExplain normalizes a NetworkPolicy into an explain structure.
func (t *Toolset) normalizeNetworkPolicyExplain(np *networkingv1.NetworkPolicy) NetworkPolicyExplain {
	explain := NetworkPolicyExplain{
		Name:      np.Name,
		Namespace: np.Namespace,
	}

	// Normalize ingress rules
	for _, rule := range np.Spec.Ingress {
		npRule := NetworkPolicyRule{
			Direction: "ingress",
		}

		// Extract peers
		for _, peer := range rule.From {
			npPeer := NetworkPolicyPeer{}
			if peer.PodSelector != nil {
				npPeer.PodSelector = peer.PodSelector.MatchLabels
			}
			if peer.NamespaceSelector != nil {
				npPeer.NamespaceSelector = peer.NamespaceSelector.MatchLabels
			}
			if peer.IPBlock != nil {
				npPeer.IPBlock = &IPBlock{
					CIDR:   peer.IPBlock.CIDR,
					Except: peer.IPBlock.Except,
				}
			}
			npRule.Peers = append(npRule.Peers, npPeer)
		}

		// Extract ports
		for _, port := range rule.Ports {
			npPort := NetworkPolicyPort{}
			if port.Protocol != nil {
				npPort.Protocol = string(*port.Protocol)
			}
			if port.Port != nil {
				if port.Port.IntVal > 0 {
					npPort.Port = strconv.Itoa(int(port.Port.IntVal))
				} else if port.Port.StrVal != "" {
					npPort.Port = port.Port.StrVal
				}
			}
			npRule.Ports = append(npRule.Ports, npPort)
		}

		explain.Ingress = append(explain.Ingress, npRule)
	}

	// Normalize egress rules
	for _, rule := range np.Spec.Egress {
		npRule := NetworkPolicyRule{
			Direction: "egress",
		}

		// Extract peers
		for _, peer := range rule.To {
			npPeer := NetworkPolicyPeer{}
			if peer.PodSelector != nil {
				npPeer.PodSelector = peer.PodSelector.MatchLabels
			}
			if peer.NamespaceSelector != nil {
				npPeer.NamespaceSelector = peer.NamespaceSelector.MatchLabels
			}
			if peer.IPBlock != nil {
				npPeer.IPBlock = &IPBlock{
					CIDR:   peer.IPBlock.CIDR,
					Except: peer.IPBlock.Except,
				}
			}
			npRule.Peers = append(npRule.Peers, npPeer)
		}

		// Extract ports
		for _, port := range rule.Ports {
			npPort := NetworkPolicyPort{}
			if port.Protocol != nil {
				npPort.Protocol = string(*port.Protocol)
			}
			if port.Port != nil {
				if port.Port.IntVal > 0 {
					npPort.Port = strconv.Itoa(int(port.Port.IntVal))
				} else if port.Port.StrVal != "" {
					npPort.Port = port.Port.StrVal
				}
			}
			npRule.Ports = append(npRule.Ports, npPort)
		}

		explain.Egress = append(explain.Egress, npRule)
	}

	return explain
}

// handleNetworkPoliciesList handles the net.networkpolicies_list tool.
func (t *Toolset) handleNetworkPoliciesList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
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

	var list *networkingv1.NetworkPolicyList
	if args.Namespace != "" {
		list, err = clientSet.Typed.NetworkingV1().NetworkPolicies(args.Namespace).List(ctx, listOptions)
	} else {
		list, err = clientSet.Typed.NetworkingV1().NetworkPolicies("").List(ctx, listOptions)
	}

	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list NetworkPolicies: %w", err)), nil
	}

	var summaries []map[string]interface{}
	for _, item := range list.Items {
		summary := map[string]interface{}{
			"name":      item.Name,
			"namespace": item.Namespace,
		}
		summaries = append(summaries, summary)
	}

	resultData := map[string]any{
		"items": summaries,
	}
	if list.Continue != "" {
		resultData["continue"] = list.Continue
		if list.RemainingItemCount != nil {
			resultData["remaining_item_count"] = *list.RemainingItemCount
		}
	}

	result, err := mcpHelpers.NewJSONResult(resultData)
	return result, err
}

// handleNetworkPolicyExplain handles the net.networkpolicy_explain tool.
func (t *Toolset) handleNetworkPolicyExplain(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	np, err := clientSet.Typed.NetworkingV1().NetworkPolicies(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get NetworkPolicy: %w", err)), nil
	}

	explain := t.normalizeNetworkPolicyExplain(np)
	result, err := mcpHelpers.NewJSONResult(explain)
	return result, err
}

// handleConnectivityHint handles the net.connectivity_hint tool.
func (t *Toolset) handleConnectivityHint(ctx context.Context, args struct {
	Context      string            `json:"context"`
	SrcNamespace string            `json:"src_namespace"`
	SrcLabels    map[string]string `json:"src_labels"`
	DstNamespace string            `json:"dst_namespace"`
	DstLabels    map[string]string `json:"dst_labels"`
	Port         string            `json:"port"`
	Protocol     string            `json:"protocol"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// List NetworkPolicies in both namespaces
	hint := ConnectivityHint{
		LikelyAllowed:     "unknown",
		Reasons:           []string{},
		EvaluatedPolicies: []string{},
	}

	// Check source namespace policies
	srcNPs, err := clientSet.Typed.NetworkingV1().NetworkPolicies(args.SrcNamespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, np := range srcNPs.Items {
			hint.EvaluatedPolicies = append(hint.EvaluatedPolicies, fmt.Sprintf("%s/%s", args.SrcNamespace, np.Name))
			// Best-effort matching - check if egress rules might allow
			for _, egress := range np.Spec.Egress {
				matchesNamespace := false
				if egress.To == nil || len(egress.To) == 0 {
					matchesNamespace = true // Allow all
				} else {
					for _, to := range egress.To {
						if to.NamespaceSelector != nil {
							// Simplified matching
							matchesNamespace = true
						}
					}
				}
				if matchesNamespace {
					hint.Reasons = append(hint.Reasons, fmt.Sprintf("Egress rule in %s/%s may allow traffic", args.SrcNamespace, np.Name))
					hint.LikelyAllowed = "true"
				}
			}
		}
	}

	// Check destination namespace policies
	dstNPs, err := clientSet.Typed.NetworkingV1().NetworkPolicies(args.DstNamespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, np := range dstNPs.Items {
			hint.EvaluatedPolicies = append(hint.EvaluatedPolicies, fmt.Sprintf("%s/%s", args.DstNamespace, np.Name))
			// Best-effort matching - check if ingress rules might allow
			for _, ingress := range np.Spec.Ingress {
				matchesNamespace := false
				if ingress.From == nil || len(ingress.From) == 0 {
					matchesNamespace = true // Allow all
				} else {
					for _, from := range ingress.From {
						if from.NamespaceSelector != nil {
							matchesNamespace = true
						}
					}
				}
				if matchesNamespace {
					hint.Reasons = append(hint.Reasons, fmt.Sprintf("Ingress rule in %s/%s may allow traffic", args.DstNamespace, np.Name))
					hint.LikelyAllowed = "true"
				}
			}
		}
	}

	// Add disclaimer
	hint.Reasons = append(hint.Reasons, "Note: This is a best-effort analysis. Actual policy evaluation is complex and depends on pod selectors, IP blocks, and other factors.")

	result, err := mcpHelpers.NewJSONResult(hint)
	return result, err
}
