package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// handleResourcesValidate handles the resources_validate tool.
// It validates a resource manifest without applying it.
func (t *Toolset) handleResourcesValidate(ctx context.Context, args struct {
	Manifest     map[string]interface{} `json:"manifest"`
	SchemaVersion string                `json:"schema_version"` // Optional: for version-specific validation
	Context      string                 `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Convert manifest to unstructured
	obj := &unstructured.Unstructured{Object: args.Manifest}

	gvk := obj.GroupVersionKind()
	if gvk.Empty() {
		return mcpHelpers.NewErrorResult(fmt.Errorf("manifest must include apiVersion and kind")), nil
	}

	// Map GVK to GVR
	mapping, err := clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to map GVK to GVR: %w", err)), nil
	}

	gvr := mapping.Resource
	namespace := obj.GetNamespace()

	// Basic validation checks
	validationErrors := []string{}

	// Check required fields
	if obj.GetName() == "" {
		validationErrors = append(validationErrors, "metadata.name is required")
	}

	// Validate namespace (if provided)
	if namespace != "" {
		// Check if namespace exists (for namespaced resources)
		if mapping.Scope.Name() == "Namespaced" {
			_, err := clientSet.Typed.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
			if err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("namespace %q does not exist", namespace))
			}
		}
	}

	// Use dry-run apply to validate the manifest
	// This catches schema validation errors, required fields, etc.
	fieldManager := "kube-mcp-validate"
	applyOptions := metav1.ApplyOptions{
		FieldManager: fieldManager,
		DryRun:       []string{metav1.DryRunAll},
	}

	_, err = clientSet.Dynamic.Resource(gvr).Namespace(namespace).
		Apply(ctx, obj.GetName(), obj, applyOptions)
	if err != nil {
		// Kubernetes API validation error
		validationErrors = append(validationErrors, fmt.Sprintf("Kubernetes API validation failed: %v", err))
	}

	// Additional best practice checks
	bestPracticeWarnings := []string{}

	// Check for common issues
	if obj.GetLabels() == nil || len(obj.GetLabels()) == 0 {
		bestPracticeWarnings = append(bestPracticeWarnings, "Consider adding labels for better resource management")
	}

	// Check resource-specific validations
	switch gvk.Kind {
	case "Deployment", "StatefulSet", "DaemonSet":
		// Check for replicas
		replicas, found, _ := unstructured.NestedInt64(obj.Object, "spec", "replicas")
		if !found || replicas < 0 {
			bestPracticeWarnings = append(bestPracticeWarnings, "Consider setting spec.replicas explicitly")
		}
	case "Service":
		// Check for selector
		selector, found, _ := unstructured.NestedMap(obj.Object, "spec", "selector")
		if !found || selector == nil || len(selector) == 0 {
			bestPracticeWarnings = append(bestPracticeWarnings, "Service should have spec.selector defined")
		}
	case "ConfigMap", "Secret":
		// Check for data
		data, found, _ := unstructured.NestedMap(obj.Object, "data")
		binaryData, binaryFound, _ := unstructured.NestedMap(obj.Object, "binaryData")
		if !found && !binaryFound || (data == nil && binaryData == nil) {
			bestPracticeWarnings = append(bestPracticeWarnings, "ConfigMap/Secret should have data or binaryData")
		}
	}

	// Build result
	result := map[string]interface{}{
		"valid":   len(validationErrors) == 0,
		"gvk":     gvk.String(),
		"name":    obj.GetName(),
		"namespace": namespace,
	}

	if len(validationErrors) > 0 {
		result["errors"] = validationErrors
	}

	if len(bestPracticeWarnings) > 0 {
		result["warnings"] = bestPracticeWarnings
	}

	// If there are errors, return error result
	if len(validationErrors) > 0 {
		errorMsg := strings.Join(validationErrors, "; ")
		if len(bestPracticeWarnings) > 0 {
			errorMsg += " | Warnings: " + strings.Join(bestPracticeWarnings, "; ")
		}
		resultJSON, jsonErr := json.Marshal(result)
		if jsonErr != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("validation failed: %s", errorMsg)), nil
		}
		return mcpHelpers.NewErrorResult(fmt.Errorf("validation failed: %s\nDetails: %s", errorMsg, string(resultJSON))), nil
	}

	// Success with optional warnings
	if len(bestPracticeWarnings) > 0 {
		resultJSON, jsonErr := mcpHelpers.NewJSONResult(result)
		if jsonErr != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", jsonErr)), nil
		}
		return resultJSON, nil
	}

	resultJSON, jsonErr := mcpHelpers.NewJSONResult(result)
	if jsonErr != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", jsonErr)), nil
	}
	return resultJSON, nil
}

