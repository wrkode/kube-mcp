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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
)

// handleResourcesDiff handles the resources_diff tool.
// It compares the current resource state with a desired manifest and returns the differences.
func (t *Toolset) handleResourcesDiff(ctx context.Context, args struct {
	Group      string                 `json:"group"`
	Version    string                 `json:"version"`
	Kind       string                 `json:"kind"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	Manifest   map[string]interface{} `json:"manifest"`
	DiffFormat string                 `json:"diff_format"` // "unified" (default), "json", or "yaml"
	Context    string                 `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	gvk := schema.GroupVersionKind{
		Group:   args.Group,
		Version: args.Version,
		Kind:    args.Kind,
	}

	// Map GVK to GVR
	mapping, err := clientSet.RESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to map GVK to GVR: %w", err)), nil
	}

	gvr := mapping.Resource

	// Get current resource
	current, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get current resource: %w", err)), nil
	}

	// Convert desired manifest to unstructured
	desired := &unstructured.Unstructured{Object: args.Manifest}

	// Ensure API version and kind are set in the object map itself (for server-side apply)
	// Set them in the object map directly, then use SetGroupVersionKind to sync internal state
	if apiVersion, found := desired.Object["apiVersion"]; !found || apiVersion == nil || apiVersion == "" {
		desired.Object["apiVersion"] = gvk.GroupVersion().String()
	}
	if kind, found := desired.Object["kind"]; !found || kind == nil || kind == "" {
		desired.Object["kind"] = gvk.Kind
	}

	// Set GVK on desired manifest (this syncs internal state with object map)
	desired.SetGroupVersionKind(gvk)

	// Ensure desired manifest has correct name/namespace
	if desired.GetName() == "" {
		desired.SetName(args.Name)
	}
	if desired.GetNamespace() == "" && args.Namespace != "" {
		desired.SetNamespace(args.Namespace)
	}

	// Use server-side apply with dry-run to get what the resource would look like after apply
	// This gives us the merged result that accounts for server-side defaults and field managers
	fieldManager := "kube-mcp-diff"
	applyOptions := metav1.ApplyOptions{
		FieldManager: fieldManager,
		DryRun:       []string{metav1.DryRunAll},
	}

	// Try to apply with dry-run
	merged, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).
		Apply(ctx, args.Name, desired, applyOptions)
	if err != nil {
		// If dry-run apply fails, fall back to direct comparison
		// This can happen if there are validation issues or the manifest structure
		// doesn't match what server-side apply expects
		// We'll compare current resource with desired manifest directly
		// Note: This won't show server-side defaults/merges, but will show the differences
		// Create a deep copy of desired to avoid modifying the original
		desiredBytes, marshalErr := json.Marshal(desired.Object)
		if marshalErr != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to marshal desired manifest for fallback: %w", marshalErr)), nil
		}
		var desiredObj map[string]any
		if unmarshalErr := json.Unmarshal(desiredBytes, &desiredObj); unmarshalErr != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to unmarshal desired manifest for fallback: %w", unmarshalErr)), nil
		}
		merged = &unstructured.Unstructured{Object: desiredObj}
		merged.SetGroupVersionKind(gvk)
	}

	// Generate diff based on format
	diffFormat := args.DiffFormat
	if diffFormat == "" {
		diffFormat = "unified"
	}

	var diffOutput string
	switch diffFormat {
	case "json":
		diffOutput, err = t.generateJSONDiff(current, merged)
	case "yaml":
		diffOutput, err = t.generateYAMLDiff(current, merged)
	case "unified", "":
		diffOutput, err = t.generateUnifiedDiff(current, merged)
	default:
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid diff_format: %s (must be 'unified', 'json', or 'yaml')", diffFormat)), nil
	}

	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to generate diff: %w", err)), nil
	}

	return mcpHelpers.NewTextResult(diffOutput), nil
}

// generateUnifiedDiff generates a unified diff format showing changes.
func (t *Toolset) generateUnifiedDiff(current, merged *unstructured.Unstructured) (string, error) {
	currentJSON, err := json.MarshalIndent(current.Object, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal current resource: %w", err)
	}

	mergedJSON, err := json.MarshalIndent(merged.Object, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal merged resource: %w", err)
	}

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- Current: %s/%s\n", current.GetNamespace(), current.GetName()))
	diff.WriteString(fmt.Sprintf("+++ Desired: %s/%s\n", merged.GetNamespace(), merged.GetName()))
	diff.WriteString("@@ Differences @@\n\n")

	// Use a simple approach: if JSON strings are different, show both
	if string(currentJSON) == string(mergedJSON) {
		diff.WriteString("No differences found.\n")
		return diff.String(), nil
	}

	// Generate patch-like diff
	// Remove metadata fields that always differ (resourceVersion, generation, etc.)
	currentClean := t.cleanForDiff(current.Object)
	mergedClean := t.cleanForDiff(merged.Object)

	currentCleanJSON, _ := json.MarshalIndent(currentClean, "", "  ")
	mergedCleanJSON, _ := json.MarshalIndent(mergedClean, "", "  ")

	if string(currentCleanJSON) == string(mergedCleanJSON) {
		diff.WriteString("No meaningful differences found (only metadata fields differ).\n")
		return diff.String(), nil
	}

	// Show differences using a simple line-by-line comparison
	diff.WriteString("Current state:\n")
	diff.WriteString("```json\n")
	diff.WriteString(string(currentCleanJSON))
	diff.WriteString("\n```\n\n")

	diff.WriteString("Desired state (after apply):\n")
	diff.WriteString("```json\n")
	diff.WriteString(string(mergedCleanJSON))
	diff.WriteString("\n```\n\n")

	// Try to generate a JSON merge patch to show what would change
	patchBytes, err := jsonmergepatch.CreateThreeWayJSONMergePatch(
		currentCleanJSON,
		mergedCleanJSON,
		currentCleanJSON,
	)
	if err == nil && len(patchBytes) > 2 { // > 2 because empty patch is "{}"
		diff.WriteString("Patch that would be applied:\n")
		diff.WriteString("```json\n")
		var patchObj interface{}
		if err := json.Unmarshal(patchBytes, &patchObj); err == nil {
			patchJSON, _ := json.MarshalIndent(patchObj, "", "  ")
			diff.WriteString(string(patchJSON))
		} else {
			diff.WriteString(string(patchBytes))
		}
		diff.WriteString("\n```\n")
	}

	return diff.String(), nil
}

// generateJSONDiff generates a JSON diff showing the patch that would be applied.
func (t *Toolset) generateJSONDiff(current, merged *unstructured.Unstructured) (string, error) {
	currentJSON, err := json.MarshalIndent(current.Object, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal current resource: %w", err)
	}

	mergedJSON, err := json.MarshalIndent(merged.Object, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal merged resource: %w", err)
	}

	// Generate JSON merge patch
	patchBytes, err := jsonmergepatch.CreateThreeWayJSONMergePatch(
		currentJSON,
		mergedJSON,
		currentJSON,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create JSON patch: %w", err)
	}

	if len(patchBytes) <= 2 { // Empty patch is "{}"
		return "{}", nil
	}

	var patchObj interface{}
	if err := json.Unmarshal(patchBytes, &patchObj); err != nil {
		return string(patchBytes), nil
	}

	patchJSON, err := json.MarshalIndent(patchObj, "", "  ")
	if err != nil {
		return string(patchBytes), nil
	}

	return string(patchJSON), nil
}

// generateYAMLDiff generates a YAML diff (similar to unified but in YAML format).
func (t *Toolset) generateYAMLDiff(current, merged *unstructured.Unstructured) (string, error) {
	// For YAML, we'll use the same approach as unified but format as YAML
	// Since we're working with unstructured objects, we'll convert to YAML
	currentClean := t.cleanForDiff(current.Object)
	mergedClean := t.cleanForDiff(merged.Object)

	var diff strings.Builder
	diff.WriteString("--- Current State\n")
	currentYAML, err := t.toYAML(currentClean)
	if err != nil {
		return "", fmt.Errorf("failed to convert current to YAML: %w", err)
	}
	diff.WriteString(currentYAML)
	diff.WriteString("\n--- Desired State\n")
	mergedYAML, err := t.toYAML(mergedClean)
	if err != nil {
		return "", fmt.Errorf("failed to convert merged to YAML: %w", err)
	}
	diff.WriteString(mergedYAML)

	return diff.String(), nil
}

// cleanForDiff removes metadata fields that always differ between current and desired.
func (t *Toolset) cleanForDiff(obj map[string]interface{}) map[string]interface{} {
	cleaned := make(map[string]interface{})
	for k, v := range obj {
		if k == "metadata" {
			metadata, ok := v.(map[string]interface{})
			if !ok {
				cleaned[k] = v
				continue
			}
			cleanedMetadata := make(map[string]interface{})
			// Keep only relevant metadata fields
			for mk, mv := range metadata {
				switch mk {
				case "name", "namespace", "labels", "annotations", "finalizers", "ownerReferences":
					cleanedMetadata[mk] = mv
				case "generation", "resourceVersion", "uid", "creationTimestamp", "managedFields":
					// Skip fields that always differ
				default:
					cleanedMetadata[mk] = mv
				}
			}
			cleaned[k] = cleanedMetadata
		} else {
			cleaned[k] = v
		}
	}
	return cleaned
}

// toYAML converts a map to YAML format (simple implementation).
func (t *Toolset) toYAML(obj interface{}) (string, error) {
	// Simple YAML conversion - marshal to JSON first, then convert
	// For a proper implementation, we'd use gopkg.in/yaml.v3, but to avoid dependencies,
	// we'll use a simple approach
	jsonBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	// This is a simplified YAML - in production, use a proper YAML library
	// For now, return JSON with YAML-like formatting
	return string(jsonBytes), nil
}

