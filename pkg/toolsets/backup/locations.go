package backup

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// normalizeBackupStorageLocationSummary normalizes an unstructured BackupStorageLocation into a summary.
func (t *Toolset) normalizeBackupStorageLocationSummary(obj *unstructured.Unstructured) BackupStorageLocationSummary {
	summary := BackupStorageLocationSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	// Get provider
	if provider, found, _ := unstructured.NestedString(obj.Object, "spec", "provider"); found {
		summary.Provider = provider
	}

	// Get storage config
	if storageType, found, _ := unstructured.NestedMap(obj.Object, "spec", "storageType"); found {
		if bucket, found, _ := unstructured.NestedString(storageType, "bucket"); found {
			summary.Bucket = bucket
		}
		if region, found, _ := unstructured.NestedString(storageType, "region"); found {
			summary.Region = region
		}
	}

	// Get access mode
	if accessMode, found, _ := unstructured.NestedString(obj.Object, "spec", "accessMode"); found {
		summary.AccessMode = accessMode
	}

	// Get phase
	if phase, found, _ := unstructured.NestedString(obj.Object, "status", "phase"); found {
		summary.Phase = phase
	}

	return summary
}

// handleLocationsList handles the backup.locations_list tool.
func (t *Toolset) handleLocationsList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(BackupStorageLocationGVK); ok && !t.hasBackupStorageLocation {
				t.hasBackupStorageLocation = true
				t.backupStorageLocationGVR = gvr
			}
		}
	}

	if !t.hasBackupStorageLocation {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "BackupStorageLocation CRD not available",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: args.LabelSelector,
	}

	var list *unstructured.UnstructuredList
	if args.Namespace != "" {
		list, err = clientSet.Dynamic.Resource(t.backupStorageLocationGVR).Namespace(args.Namespace).List(ctx, listOptions)
	} else {
		list, err = clientSet.Dynamic.Resource(t.backupStorageLocationGVR).List(ctx, listOptions)
	}

	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list BackupStorageLocations: %w", err)), nil
	}

	var locations []BackupStorageLocationSummary
	for _, item := range list.Items {
		locations = append(locations, t.normalizeBackupStorageLocationSummary(&item))
	}

	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"items": locations,
	})
	return result, err
}
