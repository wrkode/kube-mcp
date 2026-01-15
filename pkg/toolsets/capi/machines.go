package capi

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// normalizeMachineSummary normalizes a Machine object into a summary.
func (t *Toolset) normalizeMachineSummary(obj *unstructured.Unstructured) MachineSummary {
	summary := MachineSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Ready:     false,
	}

	// Extract phase
	if phase, found, _ := unstructured.NestedString(obj.Object, "status", "phase"); found {
		summary.Phase = phase
	}

	// Extract node reference
	if nodeRef, found, _ := unstructured.NestedMap(obj.Object, "status", "nodeRef"); found {
		if name, found, _ := unstructured.NestedString(nodeRef, "name"); found {
			summary.NodeRef = name
		}
	}

	// Extract version
	if version, found, _ := unstructured.NestedString(obj.Object, "spec", "version"); found {
		summary.Version = version
	}

	// Extract conditions
	if conditions, found, _ := unstructured.NestedSlice(obj.Object, "status", "conditions"); found {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				condition := Condition{}
				if typ, ok := condMap["type"].(string); ok {
					condition.Type = typ
					if typ == "Ready" {
						if status, ok := condMap["status"].(string); ok {
							summary.Ready = status == "True"
						}
					}
				}
				if status, ok := condMap["status"].(string); ok {
					condition.Status = status
				}
				if reason, ok := condMap["reason"].(string); ok {
					condition.Reason = reason
				}
				if message, ok := condMap["message"].(string); ok {
					condition.Message = message
				}
				if lastTransitionTime, ok := condMap["lastTransitionTime"].(string); ok {
					condition.Time = lastTransitionTime
				}
				summary.Conditions = append(summary.Conditions, condition)
			}
		}
	}

	return summary
}

// normalizeMachineDeploymentSummary normalizes a MachineDeployment object into a summary.
func (t *Toolset) normalizeMachineDeploymentSummary(obj *unstructured.Unstructured) MachineDeploymentSummary {
	summary := MachineDeploymentSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}

	// Extract replicas
	if replicas, found, _ := unstructured.NestedInt64(obj.Object, "spec", "replicas"); found {
		summary.ReplicasDesired = int(replicas)
	}

	// Extract status
	if status, found, _ := unstructured.NestedMap(obj.Object, "status"); found {
		if readyReplicas, found, _ := unstructured.NestedInt64(status, "readyReplicas"); found {
			summary.ReplicasReady = int(readyReplicas)
		}
		if updatedReplicas, found, _ := unstructured.NestedInt64(status, "updatedReplicas"); found {
			summary.ReplicasUpdated = int(updatedReplicas)
		}
		if paused, found, _ := unstructured.NestedBool(status, "paused"); found {
			summary.Paused = paused
		}

		// Extract conditions
		if conditions, found, _ := unstructured.NestedSlice(status, "conditions"); found {
			for _, cond := range conditions {
				if condMap, ok := cond.(map[string]interface{}); ok {
					condition := Condition{}
					if typ, ok := condMap["type"].(string); ok {
						condition.Type = typ
					}
					if status, ok := condMap["status"].(string); ok {
						condition.Status = status
					}
					if reason, ok := condMap["reason"].(string); ok {
						condition.Reason = reason
					}
					if message, ok := condMap["message"].(string); ok {
						condition.Message = message
					}
					if lastTransitionTime, ok := condMap["lastTransitionTime"].(string); ok {
						condition.Time = lastTransitionTime
					}
					summary.Conditions = append(summary.Conditions, condition)
				}
			}
		}
	}

	return summary
}

// handleMachinesList handles the capi.machines_list tool.
func (t *Toolset) handleMachinesList(ctx context.Context, args struct {
	Context          string `json:"context"`
	ClusterNamespace string `json:"cluster_namespace"`
	ClusterName      string `json:"cluster_name"`
	Limit            int    `json:"limit"`
	Continue         string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(MachineGVK); ok && !t.hasMachine {
				t.hasMachine = true
				t.machineGVR = gvr
			}
		}
	}

	if !t.hasMachine {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "Machine CRD not available",
				"details": "cluster.x-k8s.io/v1beta1/Machine CRD not detected in cluster",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// List machines with label selector for cluster
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("cluster.x-k8s.io/cluster-name=%s", args.ClusterName),
	}
	if args.Limit > 0 {
		listOpts.Limit = int64(args.Limit)
	}
	if args.Continue != "" {
		listOpts.Continue = args.Continue
	}

	list, err := clientSet.Dynamic.Resource(t.machineGVR).Namespace(args.ClusterNamespace).List(ctx, listOpts)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Machines: %w", err)), nil
	}

	machines := make([]MachineSummary, 0, len(list.Items))
	for _, item := range list.Items {
		machines = append(machines, t.normalizeMachineSummary(&item))
	}

	result := map[string]any{"items": machines}
	if list.GetContinue() != "" {
		result["continue"] = list.GetContinue()
	}
	return mcpHelpers.NewJSONResult(result)
}

// handleMachineDeploymentsList handles the capi.machinedeployments_list tool.
func (t *Toolset) handleMachineDeploymentsList(ctx context.Context, args struct {
	Context          string `json:"context"`
	ClusterNamespace string `json:"cluster_namespace"`
	ClusterName      string `json:"cluster_name"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(MachineDeploymentGVK); ok && !t.hasMachineDeployment {
				t.hasMachineDeployment = true
				t.machineDeploymentGVR = gvr
			}
		}
	}

	if !t.hasMachineDeployment {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "MachineDeployment CRD not available",
				"details": "cluster.x-k8s.io/v1beta1/MachineDeployment CRD not detected in cluster",
			},
		})
		return result, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// List machine deployments with label selector for cluster
	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("cluster.x-k8s.io/cluster-name=%s", args.ClusterName),
	}

	list, err := clientSet.Dynamic.Resource(t.machineDeploymentGVR).Namespace(args.ClusterNamespace).List(ctx, listOpts)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list MachineDeployments: %w", err)), nil
	}

	deployments := make([]MachineDeploymentSummary, 0, len(list.Items))
	for _, item := range list.Items {
		deployments = append(deployments, t.normalizeMachineDeploymentSummary(&item))
	}

	return mcpHelpers.NewJSONResult(map[string]any{"items": deployments})
}

// handleRolloutStatus handles the capi.rollout_status tool.
func (t *Toolset) handleRolloutStatus(ctx context.Context, args struct {
	Context          string `json:"context"`
	ClusterNamespace string `json:"cluster_namespace"`
	ClusterName      string `json:"cluster_name"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Get cluster
	cluster, err := clientSet.Dynamic.Resource(t.clusterGVR).Namespace(args.ClusterNamespace).Get(ctx, args.ClusterName, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get Cluster: %w", err)), nil
	}

	status := RolloutStatus{
		Ready:   false,
		Message: "",
	}

	// Check cluster ready status
	if ready, found, _ := unstructured.NestedBool(cluster.Object, "status", "ready"); found {
		status.Ready = ready
	}

	// Count machines if available
	if t.hasMachine {
		listOpts := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("cluster.x-k8s.io/cluster-name=%s", args.ClusterName),
		}
		machineList, err := clientSet.Dynamic.Resource(t.machineGVR).Namespace(args.ClusterNamespace).List(ctx, listOpts)
		if err == nil {
			status.Counts.MachinesDesired = len(machineList.Items)
			for _, machine := range machineList.Items {
				if ready, found, _ := unstructured.NestedBool(machine.Object, "status", "conditions"); found && ready {
					status.Counts.MachinesReady++
				}
				// Check if machine is updated (simplified)
				if phase, found, _ := unstructured.NestedString(machine.Object, "status", "phase"); found && phase == "Running" {
					status.Counts.MachinesUpdated++
				}
			}
		}
	}

	// Extract blockers from conditions
	if conditions, found, _ := unstructured.NestedSlice(cluster.Object, "status", "conditions"); found {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				if typ, ok := condMap["type"].(string); ok {
					if statusVal, ok := condMap["status"].(string); ok && statusVal != "True" {
						if message, ok := condMap["message"].(string); ok {
							status.Blockers = append(status.Blockers, fmt.Sprintf("%s: %s", typ, message))
						}
					}
				}
			}
		}
	}

	if status.Ready {
		status.Message = "Cluster is ready"
	} else {
		status.Message = "Cluster is not ready"
	}

	return mcpHelpers.NewJSONResult(map[string]any{"status": status})
}

// handleScaleMachineDeployment handles the capi.scale_machinedeployment tool.
func (t *Toolset) handleScaleMachineDeployment(ctx context.Context, args struct {
	Context   string `json:"context"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Replicas  int    `json:"replicas"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	// Refresh discovery in case CRDs were installed after startup
	if t.discovery != nil {
		if err := t.discovery.DiscoverCRDs(ctx); err == nil {
			// Re-check CRDs and update flags if found
			if gvr, ok := t.discovery.GetGVR(MachineDeploymentGVK); ok && !t.hasMachineDeployment {
				t.hasMachineDeployment = true
				t.machineDeploymentGVR = gvr
			}
		}
	}

	if !t.hasMachineDeployment {
		return mcpHelpers.NewErrorResult(fmt.Errorf("MachineDeployment CRD not available")), nil
	}

	if !args.Confirm {
		return mcpHelpers.NewErrorResult(fmt.Errorf("confirm must be true to scale")), nil
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// RBAC check
	if errResult, err := t.checkRBAC(ctx, clientSet, "update", t.machineDeploymentGVR, args.Namespace); errResult != nil || err != nil {
		return errResult, err
	}

	// Get current object
	obj, err := clientSet.Dynamic.Resource(t.machineDeploymentGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get MachineDeployment: %w", err)), nil
	}

	// Patch replicas
	if err := unstructured.SetNestedField(obj.Object, int64(args.Replicas), "spec", "replicas"); err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to set replicas: %w", err)), nil
	}

	// Update the object
	updated, err := clientSet.Dynamic.Resource(t.machineDeploymentGVR).Namespace(args.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to scale MachineDeployment: %w", err)), nil
	}

	summary := t.normalizeMachineDeploymentSummary(updated)

	return mcpHelpers.NewJSONResult(map[string]any{"summary": summary})
}
