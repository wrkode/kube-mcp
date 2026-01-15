package certs

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// handleStatusExplain handles the certs.status_explain tool.
func (t *Toolset) handleStatusExplain(ctx context.Context, args struct {
	Context   string `json:"context"`
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

	obj, err := clientSet.Dynamic.Resource(t.certificateGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get Certificate: %w", err)), nil
	}

	if args.Raw {
		result, err := mcpHelpers.NewJSONResult(obj.Object)
		return result, err
	}

	summary := t.normalizeCertificateSummary(obj)
	status, found, _ := unstructured.NestedMap(obj.Object, "status")

	explanation := map[string]any{
		"ready":           summary.Status == CertificateStatusReady,
		"status":          string(summary.Status),
		"issuer_ref":      summary.Issuer,
		"secret_name":     summary.SecretName,
		"dns_names":       summary.DNSNames,
		"not_after":       summary.NotAfter,
		"not_before":      summary.NotBefore,
		"conditions":      []map[string]interface{}{},
		"diagnosis_hints": []string{},
	}

	// Extract conditions
	if found {
		if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]interface{}); ok {
					explanation["conditions"] = append(explanation["conditions"].([]map[string]interface{}), condMap)

					// Generate diagnosis hints from conditions
					condType, _ := condMap["type"].(string)
					condStatus, _ := condMap["status"].(string)
					condReason, _ := condMap["reason"].(string)
					condMessage, _ := condMap["message"].(string)

					if condType == "Ready" && condStatus != "True" {
						switch condReason {
						case "Pending":
							explanation["diagnosis_hints"] = append(explanation["diagnosis_hints"].([]string), "Certificate is pending issuance")
						case "Issuing":
							explanation["diagnosis_hints"] = append(explanation["diagnosis_hints"].([]string), "Certificate is currently being issued")
						case "Failed":
							explanation["diagnosis_hints"] = append(explanation["diagnosis_hints"].([]string), fmt.Sprintf("Certificate issuance failed: %s", condMessage))
						default:
							if condMessage != "" {
								explanation["diagnosis_hints"] = append(explanation["diagnosis_hints"].([]string), condMessage)
							}
						}
					}
				}
			}
		}

		// Check for renewal time
		if notAfter, found, _ := unstructured.NestedString(status, "notAfter"); found {
			if notAfterTime, err := time.Parse(time.RFC3339, notAfter); err == nil {
				// Calculate renewal time (typically 2/3 of certificate lifetime)
				if notBefore, found, _ := unstructured.NestedString(status, "notBefore"); found {
					if notBeforeTime, err := time.Parse(time.RFC3339, notBefore); err == nil {
						lifetime := notAfterTime.Sub(notBeforeTime)
						renewalTime := notBeforeTime.Add(lifetime * 2 / 3)
						renewalTimeStr := renewalTime.Format(time.RFC3339)
						explanation["renewal_time"] = renewalTimeStr
					}
				}
			}
		}
	}

	// Additional diagnosis hints
	if summary.Status == CertificateStatusPending {
		explanation["diagnosis_hints"] = append(explanation["diagnosis_hints"].([]string), "Check issuer readiness and DNS challenge status if using ACME")
	}
	if summary.Status == CertificateStatusFailed {
		explanation["diagnosis_hints"] = append(explanation["diagnosis_hints"].([]string), "Review CertificateRequest and Order resources for detailed error information")
	}

	result, err := mcpHelpers.NewJSONResult(explanation)
	return result, err
}
