package certs

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// normalizeCertificateSummary normalizes an unstructured Certificate into a summary.
func (t *Toolset) normalizeCertificateSummary(obj *unstructured.Unstructured) CertificateSummary {
	summary := CertificateSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Status:    CertificateStatusUnknown,
	}

	// Get DNS names from spec
	if dnsNames, found, _ := unstructured.NestedStringSlice(obj.Object, "spec", "dnsNames"); found {
		summary.DNSNames = dnsNames
	}

	// Get secret name
	if secretName, found, _ := unstructured.NestedString(obj.Object, "spec", "secretName"); found {
		summary.SecretName = secretName
	}

	// Get issuer reference
	if issuerRef, found, _ := unstructured.NestedMap(obj.Object, "spec", "issuerRef"); found {
		if name, _ := issuerRef["name"].(string); name != "" {
			summary.Issuer = name
		}
	}

	// Extract status
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if found {
		// Get conditions
		if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]interface{}); ok {
					condType, _ := condMap["type"].(string)
					condStatus, _ := condMap["status"].(string)
					condMessage, _ := condMap["message"].(string)

					if condType == "Ready" {
						if condStatus == "True" {
							summary.Status = CertificateStatusReady
						} else {
							summary.Status = CertificateStatusPending
						}
						summary.Message = condMessage
						if lastTransitionTime, _ := condMap["lastTransitionTime"].(string); lastTransitionTime != "" {
							summary.LastUpdated = &lastTransitionTime
						}
					} else if condType == "Issuing" && condStatus == "True" {
						summary.Status = CertificateStatusIssuing
					}
				}
			}
		}

		// Get notAfter/notBefore from status
		if notAfter, found, _ := unstructured.NestedString(status, "notAfter"); found {
			summary.NotAfter = &notAfter
		}
		if notBefore, found, _ := unstructured.NestedString(status, "notBefore"); found {
			summary.NotBefore = &notBefore
		}
	}

	return summary
}

// handleCertificatesList handles the certs.certificates_list tool.
func (t *Toolset) handleCertificatesList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

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

	var list *unstructured.UnstructuredList
	if args.Namespace != "" {
		list, err = clientSet.Dynamic.Resource(t.certificateGVR).Namespace(args.Namespace).List(ctx, listOptions)
	} else {
		list, err = clientSet.Dynamic.Resource(t.certificateGVR).List(ctx, listOptions)
	}

	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Certificates: %w", err)), nil
	}

	var certificates []CertificateSummary
	for _, item := range list.Items {
		certificates = append(certificates, t.normalizeCertificateSummary(&item))
	}

	resultData := map[string]any{
		"items": certificates,
	}
	if list.GetContinue() != "" {
		resultData["continue"] = list.GetContinue()
		if list.GetRemainingItemCount() != nil {
			resultData["remaining_item_count"] = *list.GetRemainingItemCount()
		}
	}

	result, err := mcpHelpers.NewJSONResult(resultData)
	return result, err
}

// handleCertificateGet handles the certs.certificate_get tool.
func (t *Toolset) handleCertificateGet(ctx context.Context, args struct {
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

	details := CertificateDetails{
		CertificateSummary: t.normalizeCertificateSummary(obj),
	}

	// Extract additional details
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if found {
		if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]interface{}); ok {
					details.Conditions = append(details.Conditions, condMap)
				}
			}
		}
	}

	if issuerRef, found, _ := unstructured.NestedMap(obj.Object, "spec", "issuerRef"); found {
		details.IssuerRef = issuerRef
	}

	result, err := mcpHelpers.NewJSONResult(details)
	return result, err
}

// handleCertificateRenew handles the certs.renew tool.
func (t *Toolset) handleCertificateRenew(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	if !args.Confirm {
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to renew")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// RBAC check
	if errResult, err := t.checkRBAC(ctx, clientSet, "update", t.certificateGVR, args.Namespace); errResult != nil || err != nil {
		return errResult, err
	}

	// Get current object
	obj, err := clientSet.Dynamic.Resource(t.certificateGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get Certificate: %w", err)), nil
	}

	// Add renew annotation
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[CertManagerRenewAnnotation] = "true"
	obj.SetAnnotations(annotations)

	// Update the object
	patched, err := clientSet.Dynamic.Resource(t.certificateGVR).Namespace(args.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to renew Certificate: %w", err)), nil
	}

	summary := t.normalizeCertificateSummary(patched)
	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"result": map[string]any{
			"annotation_applied": CertManagerRenewAnnotation,
		},
		"summary": summary,
	})
	return result, err
}
