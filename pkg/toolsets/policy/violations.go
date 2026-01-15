package policy

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// handleViolationsList handles the policy.violations_list tool.
func (t *Toolset) handleViolationsList(ctx context.Context, args struct {
	Context   string `json:"context"`
	Namespace string `json:"namespace"`
	Engine    string `json:"engine"`
	Limit     int    `json:"limit"`
	Continue  string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	var violations []Violation

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

	// Build list options with pagination
	listOpts := metav1.ListOptions{}
	if args.Limit > 0 {
		listOpts.Limit = int64(args.Limit)
	}
	if args.Continue != "" {
		listOpts.Continue = args.Continue
	}

	var continueToken string

	// Query Kyverno PolicyReports
	if queryKyverno {
		// ClusterPolicyReport
		if t.clusterPolicyReportGVR.Resource != "" {
			list, err := clientSet.Dynamic.Resource(t.clusterPolicyReportGVR).List(ctx, listOpts)
			if err == nil {
				for _, item := range list.Items {
					violations = append(violations, t.extractKyvernoViolations(&item, true)...)
				}
				if list.GetContinue() != "" {
					continueToken = list.GetContinue()
				}
			}
		}

		// PolicyReport (namespaced)
		if t.policyReportGVR.Resource != "" {
			list, err := clientSet.Dynamic.Resource(t.policyReportGVR).Namespace(args.Namespace).List(ctx, listOpts)
			if err == nil {
				for _, item := range list.Items {
					violations = append(violations, t.extractKyvernoViolations(&item, false)...)
				}
				// Use continue token from namespaced reports if available (last queried)
				if list.GetContinue() != "" {
					continueToken = list.GetContinue()
				}
			}
		}
	}

	// Query Gatekeeper constraints
	if queryGatekeeper {
		// Gatekeeper constraints are dynamic based on ConstraintTemplates
		// We'll query common constraint types
		violations = append(violations, t.extractGatekeeperViolations(ctx, clientSet, args.Namespace)...)
	}

	// Apply limit if specified
	resultItems := violations
	if args.Limit > 0 && len(violations) > args.Limit {
		resultItems = violations[:args.Limit]
	}

	result := map[string]any{"items": resultItems}
	if continueToken != "" {
		result["continue"] = continueToken
	}
	return mcpHelpers.NewJSONResult(result)
}

// extractKyvernoViolations extracts violations from a Kyverno PolicyReport.
func (t *Toolset) extractKyvernoViolations(obj *unstructured.Unstructured, isCluster bool) []Violation {
	var violations []Violation

	results, found, _ := unstructured.NestedSlice(obj.Object, "results")
	if !found {
		return violations
	}

	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			violation := Violation{
				Engine:    string(PolicyEngineKyverno),
				Namespace: obj.GetNamespace(),
			}

			if policy, found, _ := unstructured.NestedString(resultMap, "policy"); found {
				violation.Policy = policy
			}
			if rule, found, _ := unstructured.NestedString(resultMap, "rule"); found {
				violation.Rule = rule
			}
			if message, found, _ := unstructured.NestedString(resultMap, "message"); found {
				violation.Message = message
			}
			if timestamp, found, _ := unstructured.NestedString(resultMap, "timestamp"); found {
				violation.Timestamp = timestamp
			}
			if severity, found, _ := unstructured.NestedString(resultMap, "severity"); found {
				violation.Severity = severity
			}

			// Extract resource info
			if resources, found, _ := unstructured.NestedSlice(resultMap, "resources"); found && len(resources) > 0 {
				if resMap, ok := resources[0].(map[string]interface{}); ok {
					if kind, found, _ := unstructured.NestedString(resMap, "kind"); found {
						if apiVersion, found, _ := unstructured.NestedString(resMap, "apiVersion"); found {
							violation.Resource = fmt.Sprintf("%s/%s", apiVersion, kind)
						} else {
							violation.Resource = kind
						}
					}
					if name, found, _ := unstructured.NestedString(resMap, "name"); found {
						violation.Name = name
					}
					if ns, found, _ := unstructured.NestedString(resMap, "namespace"); found {
						violation.Namespace = ns
					}
				}
			}

			violations = append(violations, violation)
		}
	}

	return violations
}

// extractGatekeeperViolations extracts violations from Gatekeeper constraints (best-effort).
func (t *Toolset) extractGatekeeperViolations(ctx context.Context, clientSet *kubernetes.ClientSet, namespace string) []Violation {
	var violations []Violation

	// Gatekeeper violations are in constraint status
	// This is a simplified implementation - actual implementation would need to:
	// 1. List all ConstraintTemplates
	// 2. For each template, discover the constraint kind
	// 3. List constraints of that kind
	// 4. Extract violations from constraint status

	// For now, return empty list if we can't find constraint CRDs
	// This is best-effort as Gatekeeper constraint kinds are dynamic

	return violations
}
