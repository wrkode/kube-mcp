package policy

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// normalizeKyvernoPolicy normalizes a Kyverno policy.
func (t *Toolset) normalizeKyvernoPolicy(obj *unstructured.Unstructured, isClusterPolicy bool) PolicySummary {
	summary := PolicySummary{
		Engine:    string(PolicyEngineKyverno),
		Kind:      obj.GetKind(),
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Ready:     false,
		Active:    false,
	}

	// Check validationFailureAction to determine if active
	if action, found, _ := unstructured.NestedString(obj.Object, "spec", "validationFailureAction"); found {
		summary.Active = action != "audit" // Active if not audit-only
	}

	// Check status
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if found {
		if ready, found, _ := unstructured.NestedBool(status, "ready"); found {
			summary.Ready = ready
		}
		if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]interface{}); ok {
					if condType, _ := condMap["type"].(string); condType == "Ready" {
						if condStatus, _ := condMap["status"].(string); condStatus == "True" {
							summary.Ready = true
						}
						if condMessage, _ := condMap["message"].(string); condMessage != "" {
							summary.Message = condMessage
						}
					}
				}
			}
		}
	}

	return summary
}

// normalizeGatekeeperConstraintTemplate normalizes a Gatekeeper ConstraintTemplate.
func (t *Toolset) normalizeGatekeeperConstraintTemplate(obj *unstructured.Unstructured) PolicySummary {
	summary := PolicySummary{
		Engine: string(PolicyEngineGatekeeper),
		Kind:   obj.GetKind(),
		Name:   obj.GetName(),
		Ready:  false,
		Active: false,
	}

	// Check status
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if found {
		if created, found, _ := unstructured.NestedBool(status, "created"); found {
			summary.Ready = created
		}
		if byPod, found, _ := unstructured.NestedSlice(status, "byPod"); found && len(byPod) > 0 {
			summary.Active = true
		}
	}

	return summary
}

// extractRules extracts rule names from a policy object.
func (t *Toolset) extractRules(obj *unstructured.Unstructured, engine PolicyEngine) []string {
	var rules []string

	if engine == PolicyEngineKyverno {
		if ruleList, found, _ := unstructured.NestedSlice(obj.Object, "spec", "rules"); found {
			for _, rule := range ruleList {
				if ruleMap, ok := rule.(map[string]interface{}); ok {
					if name, found, _ := unstructured.NestedString(ruleMap, "name"); found {
						rules = append(rules, name)
					}
				}
			}
		}
	} else if engine == PolicyEngineGatekeeper {
		// For Gatekeeper, rules are in the CRD spec
		if _, found, _ := unstructured.NestedMap(obj.Object, "spec", "crd", "spec", "versions"); found {
			// Extract validation rules from CRD spec
			// This is simplified - actual rule extraction would need to parse the CRD
			rules = append(rules, "validation")
		}
	}

	return rules
}

// handlePoliciesList handles the policy.policies_list tool.
func (t *Toolset) handlePoliciesList(ctx context.Context, args struct {
	Context   string `json:"context"`
	Namespace string `json:"namespace"`
	Engine    string `json:"engine"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var policies []PolicySummary

	// Determine which engines to query
	queryKyverno := false
	queryGatekeeper := false

	if args.Engine == "" || args.Engine == "all" {
		queryKyverno = t.hasKyverno
		queryGatekeeper = t.hasGatekeeper
	} else if args.Engine == "kyverno" {
		queryKyverno = t.hasKyverno
	} else if args.Engine == "gatekeeper" {
		queryGatekeeper = t.hasGatekeeper
	}

	// Query Kyverno ClusterPolicy
	if queryKyverno && t.clusterPolicyGVR.Resource != "" {
		list, err := clientSet.Dynamic.Resource(t.clusterPolicyGVR).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, item := range list.Items {
				policies = append(policies, t.normalizeKyvernoPolicy(&item, true))
			}
		}
	}

	// Query Kyverno Policy (namespaced)
	if queryKyverno && t.policyGVR.Resource != "" {
		list, err := clientSet.Dynamic.Resource(t.policyGVR).Namespace(args.Namespace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, item := range list.Items {
				policies = append(policies, t.normalizeKyvernoPolicy(&item, false))
			}
		}
	}

	// Query Gatekeeper ConstraintTemplate
	if queryGatekeeper && t.constraintTemplateGVR.Resource != "" {
		list, err := clientSet.Dynamic.Resource(t.constraintTemplateGVR).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, item := range list.Items {
				policies = append(policies, t.normalizeGatekeeperConstraintTemplate(&item))
			}
		}
	}

	return mcpHelpers.NewJSONResult(map[string]any{"items": policies})
}

// handlePolicyGet handles the policy.policy_get tool.
func (t *Toolset) handlePolicyGet(ctx context.Context, args struct {
	Context   string `json:"context"`
	Engine    string `json:"engine"`
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
	var engine PolicyEngine
	var isClusterScoped bool

	switch args.Engine {
	case "kyverno":
		if !t.hasKyverno {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"error": map[string]any{
					"type":    "FeatureNotInstalled",
					"message": "Kyverno CRDs not available",
					"details": "kyverno.io/v1/ClusterPolicy or kyverno.io/v1/Policy CRD not detected in cluster",
				},
			})
			return result, err
		}
		engine = PolicyEngineKyverno
		if args.Kind == "ClusterPolicy" {
			gvr = t.clusterPolicyGVR
			isClusterScoped = true
		} else if args.Kind == "Policy" {
			gvr = t.policyGVR
			isClusterScoped = false
			if args.Namespace == "" {
				return mcpHelpers.NewErrorResult(fmt.Errorf("namespace is required for Policy")), nil
			}
		} else {
			return mcpHelpers.NewErrorResult(fmt.Errorf("invalid kind for Kyverno: %s (must be ClusterPolicy or Policy)", args.Kind)), nil
		}
	case "gatekeeper":
		if !t.hasGatekeeper {
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"error": map[string]any{
					"type":    "FeatureNotInstalled",
					"message": "Gatekeeper CRDs not available",
					"details": "templates.gatekeeper.sh/v1beta1/ConstraintTemplate CRD not detected in cluster",
				},
			})
			return result, err
		}
		engine = PolicyEngineGatekeeper
		if args.Kind != "ConstraintTemplate" {
			return mcpHelpers.NewErrorResult(fmt.Errorf("invalid kind for Gatekeeper: %s (must be ConstraintTemplate)", args.Kind)), nil
		}
		gvr = t.constraintTemplateGVR
		isClusterScoped = true
	default:
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid engine: %s (must be kyverno or gatekeeper)", args.Engine)), nil
	}

	var obj *unstructured.Unstructured
	if isClusterScoped {
		obj, err = clientSet.Dynamic.Resource(gvr).Get(ctx, args.Name, metav1.GetOptions{})
	} else {
		obj, err = clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	}
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get %s: %w", args.Kind, err)), nil
	}

	var summary PolicySummary
	if engine == PolicyEngineKyverno {
		summary = t.normalizeKyvernoPolicy(obj, args.Kind == "ClusterPolicy")
	} else {
		summary = t.normalizeGatekeeperConstraintTemplate(obj)
	}

	details := PolicyDetails{
		PolicySummary: summary,
		Rules:         t.extractRules(obj, engine),
	}

	result := map[string]any{
		"summary": details,
	}

	if args.Raw {
		result["raw_object"] = obj.Object
	}

	return mcpHelpers.NewJSONResult(result)
}
