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

// normalizeRestoreSummary normalizes an unstructured Restore into a summary.
func (t *Toolset) normalizeRestoreSummary(obj *unstructured.Unstructured) RestoreSummary {
	summary := RestoreSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Phase:     BackupPhaseUnknown,
	}

	// Get backup name
	if backupName, found, _ := unstructured.NestedString(obj.Object, "spec", "backupName"); found {
		summary.BackupName = backupName
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

	// Get errors/warnings
	if errors, found, _ := unstructured.NestedInt64(obj.Object, "status", "errors"); found {
		summary.Errors = int(errors)
	}
	if warnings, found, _ := unstructured.NestedInt64(obj.Object, "status", "warnings"); found {
		summary.Warnings = int(warnings)
	}

	return summary
}

// handleRestoresList handles the backup.restores_list tool.
func (t *Toolset) handleRestoresList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(RestoreGVK); ok && !t.hasRestore {
				t.hasRestore = true
				t.restoreGVR = gvr
			}
		}
	}

	if !t.hasRestore {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "Restore CRD not available",
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
	if args.Limit > 0 {
		listOptions.Limit = int64(args.Limit)
	}
	if args.Continue != "" {
		listOptions.Continue = args.Continue
	}

	var list *unstructured.UnstructuredList
	if args.Namespace != "" {
		list, err = clientSet.Dynamic.Resource(t.restoreGVR).Namespace(args.Namespace).List(ctx, listOptions)
	} else {
		list, err = clientSet.Dynamic.Resource(t.restoreGVR).List(ctx, listOptions)
	}

	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Restores: %w", err)), nil
	}

	var restores []RestoreSummary
	for _, item := range list.Items {
		restores = append(restores, t.normalizeRestoreSummary(&item))
	}

	resultData := map[string]any{
		"items": restores,
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

// handleRestoreCreate handles the backup.restore_create tool.
func (t *Toolset) handleRestoreCreate(ctx context.Context, args struct {
	Context            string   `json:"context"`
	Name               string   `json:"name"`
	Namespace          string   `json:"namespace"`
	BackupName         string   `json:"backup_name"`
	IncludedNamespaces []string `json:"included_namespaces,omitempty"`
	ExcludedNamespaces []string `json:"excluded_namespaces,omitempty"`
	Confirm            bool     `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(RestoreGVK); ok && !t.hasRestore {
				t.hasRestore = true
				t.restoreGVR = gvr
			}
		}
	}

	if !t.hasRestore {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "Restore CRD not available",
			},
		})
		return result, err
	}

	if !args.Confirm {
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to create restore")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// RBAC check
	if errResult, err := t.checkRBAC(ctx, clientSet, "create", t.restoreGVR, args.Namespace); errResult != nil || err != nil {
		return errResult, err
	}

	// Generate name if not provided
	name := args.Name
	if name == "" {
		name = fmt.Sprintf("restore-%d", time.Now().Unix())
	}

	// Build Restore spec
	restore := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "velero.io/v1",
			"kind":       "Restore",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": args.Namespace,
			},
			"spec": map[string]interface{}{
				"backupName": args.BackupName,
			},
		},
	}

	if len(args.IncludedNamespaces) > 0 {
		restore.Object["spec"].(map[string]interface{})["includedNamespaces"] = args.IncludedNamespaces
	}
	if len(args.ExcludedNamespaces) > 0 {
		restore.Object["spec"].(map[string]interface{})["excludedNamespaces"] = args.ExcludedNamespaces
	}

	// Create the Restore
	created, err := clientSet.Dynamic.Resource(t.restoreGVR).Namespace(args.Namespace).Create(ctx, restore, metav1.CreateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create Restore: %w", err)), nil
	}

	summary := t.normalizeRestoreSummary(created)
	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"result": map[string]any{
			"name":      created.GetName(),
			"namespace": created.GetNamespace(),
		},
		"summary": summary,
	})
	return result, err
}
