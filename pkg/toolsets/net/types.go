package net

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NetworkPolicyRule represents a normalized network policy rule.
type NetworkPolicyRule struct {
	Direction string              `json:"direction"` // "ingress" or "egress"
	Peers     []NetworkPolicyPeer `json:"peers,omitempty"`
	Ports     []NetworkPolicyPort `json:"ports,omitempty"`
}

// NetworkPolicyPeer represents a peer selector.
type NetworkPolicyPeer struct {
	PodSelector       map[string]string `json:"pod_selector,omitempty"`
	NamespaceSelector map[string]string `json:"namespace_selector,omitempty"`
	IPBlock           *IPBlock          `json:"ip_block,omitempty"`
}

// IPBlock represents an IP block.
type IPBlock struct {
	CIDR   string   `json:"cidr"`
	Except []string `json:"except,omitempty"`
}

// NetworkPolicyPort represents a port/protocol combination.
type NetworkPolicyPort struct {
	Protocol string `json:"protocol,omitempty"` // "TCP", "UDP", "SCTP"
	Port     string `json:"port,omitempty"`     // Port number or range
}

// NetworkPolicyExplain represents an explained network policy.
type NetworkPolicyExplain struct {
	Name      string              `json:"name"`
	Namespace string              `json:"namespace"`
	Ingress   []NetworkPolicyRule `json:"ingress,omitempty"`
	Egress    []NetworkPolicyRule `json:"egress,omitempty"`
}

// ConnectivityHint represents a connectivity analysis result.
type ConnectivityHint struct {
	LikelyAllowed     string   `json:"likely_allowed"` // "true", "false", "unknown"
	Reasons           []string `json:"reasons,omitempty"`
	EvaluatedPolicies []string `json:"evaluated_policies,omitempty"`
}

// CiliumPolicySummary represents a normalized Cilium policy summary.
type CiliumPolicySummary struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Kind      string   `json:"kind"` // "CiliumNetworkPolicy" or "CiliumClusterwideNetworkPolicy"
	Endpoints []string `json:"endpoints,omitempty"`
}

// HubbleFlow represents a normalized Hubble flow.
type HubbleFlow struct {
	Time        string        `json:"time,omitempty"`
	Verdict     string        `json:"verdict,omitempty"`
	Source      *FlowEndpoint `json:"source,omitempty"`
	Destination *FlowEndpoint `json:"destination,omitempty"`
	Protocol    string        `json:"protocol,omitempty"`
	Port        *int32        `json:"port,omitempty"`
}

// FlowEndpoint represents a flow endpoint.
type FlowEndpoint struct {
	Namespace string            `json:"namespace,omitempty"`
	Pod       string            `json:"pod,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	IP        string            `json:"ip,omitempty"`
}

// GVKs for Cilium CRDs
var (
	CiliumNetworkPolicyGVK = schema.GroupVersionKind{
		Group:   "cilium.io",
		Version: "v2",
		Kind:    "CiliumNetworkPolicy",
	}
	CiliumClusterwideNetworkPolicyGVK = schema.GroupVersionKind{
		Group:   "cilium.io",
		Version: "v2",
		Kind:    "CiliumClusterwideNetworkPolicy",
	}
)
