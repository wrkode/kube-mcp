package policy

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// handleExplainDenial handles the policy.explain_denial tool.
func (t *Toolset) handleExplainDenial(ctx context.Context, args struct {
	Context string `json:"context"`
	Message string `json:"message"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	matches := []ExplainMatch{}

	// Heuristic matching: search for policy names and rule names in the message
	// This is best-effort and documented as such

	// Try to match Kyverno policies
	if t.hasKyverno {
		matches = append(matches, t.matchKyvernoPolicies(ctx, clientSet, args.Message)...)
	}

	// Try to match Gatekeeper policies
	if t.hasGatekeeper {
		matches = append(matches, t.matchGatekeeperPolicies(ctx, clientSet, args.Message)...)
	}

	return mcpHelpers.NewJSONResult(map[string]any{"matches": matches})
}

// matchKyvernoPolicies attempts to match Kyverno policies from the denial message.
func (t *Toolset) matchKyvernoPolicies(ctx context.Context, clientSet *kubernetes.ClientSet, message string) []ExplainMatch {
	var matches []ExplainMatch

	// Heuristic: look for policy names in the message
	// This is simplified - actual implementation would parse the message more carefully

	// List ClusterPolicies
	if t.clusterPolicyGVR.Resource != "" {
		list, err := clientSet.Dynamic.Resource(t.clusterPolicyGVR).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, item := range list.Items {
				policyName := item.GetName()
				if strings.Contains(message, policyName) {
					matches = append(matches, ExplainMatch{
						Engine:      string(PolicyEngineKyverno),
						Policy:      policyName,
						Rule:        "unknown",
						Confidence:  0.5, // Low confidence for heuristic matching
						Explanation: fmt.Sprintf("Message contains policy name '%s'", policyName),
					})
				}
			}
		}
	}

	// List Policies
	if t.policyGVR.Resource != "" {
		list, err := clientSet.Dynamic.Resource(t.policyGVR).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, item := range list.Items {
				policyName := item.GetName()
				if strings.Contains(message, policyName) {
					matches = append(matches, ExplainMatch{
						Engine:      string(PolicyEngineKyverno),
						Policy:      policyName,
						Rule:        "unknown",
						Confidence:  0.5,
						Explanation: fmt.Sprintf("Message contains policy name '%s'", policyName),
					})
				}
			}
		}
	}

	return matches
}

// matchGatekeeperPolicies attempts to match Gatekeeper policies from the denial message.
func (t *Toolset) matchGatekeeperPolicies(ctx context.Context, clientSet *kubernetes.ClientSet, message string) []ExplainMatch {
	var matches []ExplainMatch

	// Heuristic: look for constraint template names in the message
	if t.constraintTemplateGVR.Resource != "" {
		list, err := clientSet.Dynamic.Resource(t.constraintTemplateGVR).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, item := range list.Items {
				templateName := item.GetName()
				if strings.Contains(message, templateName) {
					matches = append(matches, ExplainMatch{
						Engine:      string(PolicyEngineGatekeeper),
						Policy:      templateName,
						Rule:        "validation",
						Confidence:  0.5,
						Explanation: fmt.Sprintf("Message contains constraint template name '%s'", templateName),
					})
				}
			}
		}
	}

	return matches
}
