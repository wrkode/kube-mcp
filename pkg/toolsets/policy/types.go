package policy

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PolicyEngine represents the policy engine type.
type PolicyEngine string

const (
	PolicyEngineKyverno    PolicyEngine = "kyverno"
	PolicyEngineGatekeeper PolicyEngine = "gatekeeper"
)

// PolicySummary represents a normalized policy summary.
type PolicySummary struct {
	Engine    string `json:"engine"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Ready     bool   `json:"ready,omitempty"`
	Active    bool   `json:"active,omitempty"`
	Message   string `json:"message,omitempty"`
}

// PolicyDetails represents detailed policy information.
type PolicyDetails struct {
	PolicySummary
	Rules []string `json:"rules,omitempty"`
}

// Violation represents a policy violation.
type Violation struct {
	Engine    string `json:"engine"`
	Policy    string `json:"policy"`
	Rule      string `json:"rule,omitempty"`
	Resource  string `json:"resource"` // GVK format: group/version/kind
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp,omitempty"` // RFC3339 string
	Severity  string `json:"severity,omitempty"`
}

// ExplainMatch represents a match from explain_denial.
type ExplainMatch struct {
	Engine      string  `json:"engine"`
	Policy      string  `json:"policy"`
	Rule        string  `json:"rule"`
	Confidence  float64 `json:"confidence"` // 0.0 to 1.0
	Explanation string  `json:"explanation"`
}

// GVKs for Policy CRDs
var (
	// Kyverno CRDs
	KyvernoClusterPolicyGVK = schema.GroupVersionKind{
		Group:   "kyverno.io",
		Version: "v1",
		Kind:    "ClusterPolicy",
	}
	KyvernoPolicyGVK = schema.GroupVersionKind{
		Group:   "kyverno.io",
		Version: "v1",
		Kind:    "Policy",
	}
	KyvernoPolicyReportGVK = schema.GroupVersionKind{
		Group:   "wgpolicyk8s.io",
		Version: "v1alpha2",
		Kind:    "PolicyReport",
	}
	KyvernoClusterPolicyReportGVK = schema.GroupVersionKind{
		Group:   "wgpolicyk8s.io",
		Version: "v1alpha2",
		Kind:    "ClusterPolicyReport",
	}

	// Gatekeeper CRDs
	GatekeeperConstraintTemplateGVK = schema.GroupVersionKind{
		Group:   "templates.gatekeeper.sh",
		Version: "v1beta1",
		Kind:    "ConstraintTemplate",
	}
	GatekeeperConstraintGVK = schema.GroupVersionKind{
		Group:   "constraints.gatekeeper.sh",
		Version: "v1beta1",
		Kind:    "", // Generic - actual kind depends on template
	}
)
