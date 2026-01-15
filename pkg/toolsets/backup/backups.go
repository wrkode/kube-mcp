package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// normalizeBackupSummary normalizes an unstructured Backup into a summary.
func (t *Toolset) normalizeBackupSummary(obj *unstructured.Unstructured) BackupSummary {
	summary := BackupSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Phase:     BackupPhaseUnknown,
	}

	// Get phase
	if phase, found, _ := unstructured.NestedString(obj.Object, "status", "phase"); found {
		switch phase {
		case "New":
			summary.Phase = BackupPhaseNew
		case "InProgress":
			summary.Phase = BackupPhaseInProgress
		case "Completed":
			summary.Phase = BackupPhaseCompleted
		case "Failed":
			summary.Phase = BackupPhaseFailed
		case "PartiallyFailed":
			summary.Phase = BackupPhasePartiallyFailed
		default:
			summary.Phase = BackupPhaseUnknown
		}
	}

	// Get timestamps
	if startTime, found, _ := unstructured.NestedString(obj.Object, "status", "startTimestamp"); found {
		summary.StartTimestamp = &startTime
	}
	if completionTime, found, _ := unstructured.NestedString(obj.Object, "status", "completionTimestamp"); found {
		summary.CompletionTimestamp = &completionTime
	}
	if expiration, found, _ := unstructured.NestedString(obj.Object, "status", "expiration"); found {
		summary.Expiration = &expiration
	}

	// Get errors/warnings
	if errors, found, _ := unstructured.NestedInt64(obj.Object, "status", "errors"); found {
		summary.Errors = int(errors)
	}
	if warnings, found, _ := unstructured.NestedInt64(obj.Object, "status", "warnings"); found {
		summary.Warnings = int(warnings)
	}

	// Get message from validation errors
	if validationErrors, found, _ := unstructured.NestedSlice(obj.Object, "status", "validationErrors"); found && len(validationErrors) > 0 {
		if msg, ok := validationErrors[0].(string); ok {
			summary.Message = msg
		}
	}

	return summary
}

// handleBackupsList handles the backup.backups_list tool.
func (t *Toolset) handleBackupsList(ctx context.Context, args struct {
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
		list, err = clientSet.Dynamic.Resource(t.backupGVR).Namespace(args.Namespace).List(ctx, listOptions)
	} else {
		list, err = clientSet.Dynamic.Resource(t.backupGVR).List(ctx, listOptions)
	}

	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Backups: %w", err)), nil
	}

	var backups []BackupSummary
	for _, item := range list.Items {
		backups = append(backups, t.normalizeBackupSummary(&item))
	}

	resultData := map[string]any{
		"items": backups,
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

// handleBackupGet handles the backup.backup_get tool.
func (t *Toolset) handleBackupGet(ctx context.Context, args struct {
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

	obj, err := clientSet.Dynamic.Resource(t.backupGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get Backup: %w", err)), nil
	}

	if args.Raw {
		result, err := mcpHelpers.NewJSONResult(obj.Object)
		return result, err
	}

	summary := t.normalizeBackupSummary(obj)
	details := BackupDetails{
		BackupSummary: summary,
	}

	// Extract additional details
	if spec, found, _ := unstructured.NestedMap(obj.Object, "spec"); found {
		if includedNamespaces, found, _ := unstructured.NestedStringSlice(spec, "includedNamespaces"); found {
			details.IncludedNamespaces = includedNamespaces
		}
		if excludedNamespaces, found, _ := unstructured.NestedStringSlice(spec, "excludedNamespaces"); found {
			details.ExcludedNamespaces = excludedNamespaces
		}
		if includedResources, found, _ := unstructured.NestedStringSlice(spec, "includedResources"); found {
			details.IncludedResources = includedResources
		}
		if excludedResources, found, _ := unstructured.NestedStringSlice(spec, "excludedResources"); found {
			details.ExcludedResources = excludedResources
		}
		if labelSelector, found, _ := unstructured.NestedMap(spec, "labelSelector"); found {
			labelMap := make(map[string]string)
			if matchLabels, found, _ := unstructured.NestedMap(labelSelector, "matchLabels"); found {
				for k, v := range matchLabels {
					if vStr, ok := v.(string); ok {
						labelMap[k] = vStr
					}
				}
			}
			details.LabelSelector = labelMap
		}
		if snapshotVolumes, found, _ := unstructured.NestedBool(spec, "snapshotVolumes"); found {
			details.SnapshotVolumes = &snapshotVolumes
		}
		if includeClusterResources, found, _ := unstructured.NestedBool(spec, "includeClusterResources"); found {
			details.IncludeClusterResources = &includeClusterResources
		}
		if storageLocation, found, _ := unstructured.NestedString(spec, "storageLocation"); found {
			details.StorageLocation = storageLocation
		}
	}

	if status, found, _ := unstructured.NestedMap(obj.Object, "status"); found {
		if volumeSnapshots, found, _ := unstructured.NestedStringSlice(status, "volumeSnapshots"); found {
			details.VolumeSnapshots = volumeSnapshots
		}
		if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]interface{}); ok {
					details.Conditions = append(details.Conditions, condMap)
				}
			}
		}
	}

	result, err := mcpHelpers.NewJSONResult(details)
	return result, err
}

// handleBackupCreate handles the backup.backup_create tool.
func (t *Toolset) handleBackupCreate(ctx context.Context, args struct {
	Context                 string            `json:"context"`
	Name                    string            `json:"name"`
	Namespace               string            `json:"namespace"`
	TTL                     string            `json:"ttl,omitempty"`
	IncludedNamespaces      []string          `json:"included_namespaces,omitempty"`
	ExcludedNamespaces      []string          `json:"excluded_namespaces,omitempty"`
	LabelSelector           map[string]string `json:"label_selector,omitempty"`
	SnapshotVolumes         *bool             `json:"snapshot_volumes,omitempty"`
	IncludeClusterResources *bool             `json:"include_cluster_resources,omitempty"`
	Confirm                 bool              `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	if !args.Confirm {
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to create backup")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// RBAC check
	if errResult, err := t.checkRBAC(ctx, clientSet, "create", t.backupGVR, args.Namespace); errResult != nil || err != nil {
		return errResult, err
	}

	// Generate name if not provided
	name := args.Name
	if name == "" {
		name = fmt.Sprintf("backup-%d", time.Now().Unix())
	}

	// Build Backup spec
	backup := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "velero.io/v1",
			"kind":       "Backup",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": args.Namespace,
			},
			"spec": map[string]interface{}{},
		},
	}

	if args.TTL != "" {
		backup.Object["spec"].(map[string]interface{})["ttl"] = args.TTL
	}
	if len(args.IncludedNamespaces) > 0 {
		backup.Object["spec"].(map[string]interface{})["includedNamespaces"] = args.IncludedNamespaces
	}
	if len(args.ExcludedNamespaces) > 0 {
		backup.Object["spec"].(map[string]interface{})["excludedNamespaces"] = args.ExcludedNamespaces
	}
	if len(args.LabelSelector) > 0 {
		backup.Object["spec"].(map[string]interface{})["labelSelector"] = map[string]interface{}{
			"matchLabels": args.LabelSelector,
		}
	}
	if args.SnapshotVolumes != nil {
		backup.Object["spec"].(map[string]interface{})["snapshotVolumes"] = *args.SnapshotVolumes
	}
	if args.IncludeClusterResources != nil {
		backup.Object["spec"].(map[string]interface{})["includeClusterResources"] = *args.IncludeClusterResources
	}

	// Create the Backup
	created, err := clientSet.Dynamic.Resource(t.backupGVR).Namespace(args.Namespace).Create(ctx, backup, metav1.CreateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create Backup: %w", err)), nil
	}

	summary := t.normalizeBackupSummary(created)
	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"result": map[string]any{
			"name":      created.GetName(),
			"namespace": created.GetNamespace(),
		},
		"summary": summary,
	})
	return result, err
}
