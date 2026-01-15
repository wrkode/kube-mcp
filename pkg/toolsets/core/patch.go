package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// handleResourcesPatch handles the resources_patch tool.
func (t *Toolset) handleResourcesPatch(ctx context.Context, args struct {
	Group        string      `json:"group"`
	Version      string      `json:"version"`
	Kind         string      `json:"kind"`
	Name         string      `json:"name"`
	Namespace    string      `json:"namespace"`
	PatchType    string      `json:"patch_type"` // "merge", "json", or "strategic"
	PatchData    interface{} `json:"patch_data"` // object for merge/strategic, array for json patch
	FieldManager string      `json:"field_manager"`
	DryRun       bool        `json:"dry_run"`
	Context      string      `json:"context"`
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

	// Check RBAC before patch (even in dry-run mode to validate permissions)
	if rbacResult, rbacErr := t.checkRBAC(ctx, clientSet, "patch", gvr, args.Namespace); rbacErr != nil || rbacResult != nil {
		if rbacResult != nil {
			return rbacResult, nil
		}
		return mcpHelpers.NewErrorResult(rbacErr), nil
	}

	// Determine patch type
	var patchType types.PatchType
	switch args.PatchType {
	case "merge", "":
		patchType = types.MergePatchType
	case "json":
		patchType = types.JSONPatchType
	case "strategic":
		patchType = types.StrategicMergePatchType
	default:
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid patch_type: %s (must be 'merge', 'json', or 'strategic')", args.PatchType)), nil
	}

	// Marshal patch data to JSON
	var patchBytes []byte
	if patchType == types.JSONPatchType {
		// JSON Patch requires an array of patch operations
		// Handle both array and single operation object
		switch v := args.PatchData.(type) {
		case []interface{}:
			// Already an array, marshal as-is
			patchBytes, err = json.Marshal(v)
		case map[string]interface{}:
			// Single operation object, check if it has "op" field
			if _, hasOp := v["op"]; hasOp {
				// Wrap single operation in array
				patchBytes, err = json.Marshal([]map[string]interface{}{v})
			} else {
				// Not a valid JSON Patch operation, marshal as-is and let API validate
				patchBytes, err = json.Marshal(v)
			}
		default:
			// Try to marshal as-is
			patchBytes, err = json.Marshal(args.PatchData)
		}
	} else {
		// Merge and Strategic Merge patches use object format
		// Convert to map if needed
		if mapData, ok := args.PatchData.(map[string]interface{}); ok {
			patchBytes, err = json.Marshal(mapData)
		} else {
			// Try to marshal as-is
			patchBytes, err = json.Marshal(args.PatchData)
		}
	}
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to marshal patch data: %w", err)), nil
	}

	// Build patch options
	patchOptions := metav1.PatchOptions{}
	if args.DryRun {
		patchOptions.DryRun = []string{metav1.DryRunAll}
	}
	if args.FieldManager != "" {
		patchOptions.FieldManager = args.FieldManager
	}

	// Apply patch
	patched, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).
		Patch(ctx, args.Name, patchType, patchBytes, patchOptions)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to patch resource: %w", err)), nil
	}

	status := "patched"
	if args.DryRun {
		status = "dry-run-patched"
	}

	return mcpHelpers.NewJSONResult(map[string]any{
		"name":      patched.GetName(),
		"namespace": patched.GetNamespace(),
		"kind":      patched.GetKind(),
		"status":    status,
		"dry_run":   args.DryRun,
		"patch_type": args.PatchType,
	})
}

