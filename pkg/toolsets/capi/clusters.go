package capi

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// normalizeClusterSummary normalizes a Cluster object into a summary.
func (t *Toolset) normalizeClusterSummary(obj *unstructured.Unstructured) ClusterSummary {
	summary := ClusterSummary{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
		Ready:     false,
	}

	// Extract Kubernetes version
	if spec, found, _ := unstructured.NestedMap(obj.Object, "spec"); found {
		if topology, found, _ := unstructured.NestedMap(spec, "topology"); found {
			if version, found, _ := unstructured.NestedString(topology, "version"); found {
				summary.KubernetesVersion = version
			}
		}
	}

	// Extract references
	if spec, found, _ := unstructured.NestedMap(obj.Object, "spec"); found {
		if controlPlaneRef, found, _ := unstructured.NestedMap(spec, "controlPlaneRef"); found {
			ref := &ObjectReference{}
			if apiVersion, found, _ := unstructured.NestedString(controlPlaneRef, "apiVersion"); found {
				ref.APIVersion = apiVersion
			}
			if kind, found, _ := unstructured.NestedString(controlPlaneRef, "kind"); found {
				ref.Kind = kind
			}
			if name, found, _ := unstructured.NestedString(controlPlaneRef, "name"); found {
				ref.Name = name
			}
			if namespace, found, _ := unstructured.NestedString(controlPlaneRef, "namespace"); found {
				ref.Namespace = namespace
			}
			summary.ControlPlaneRef = ref
		}

		if infrastructureRef, found, _ := unstructured.NestedMap(spec, "infrastructureRef"); found {
			ref := &ObjectReference{}
			if apiVersion, found, _ := unstructured.NestedString(infrastructureRef, "apiVersion"); found {
				ref.APIVersion = apiVersion
			}
			if kind, found, _ := unstructured.NestedString(infrastructureRef, "kind"); found {
				ref.Kind = kind
			}
			if name, found, _ := unstructured.NestedString(infrastructureRef, "name"); found {
				ref.Name = name
			}
			if namespace, found, _ := unstructured.NestedString(infrastructureRef, "namespace"); found {
				ref.Namespace = namespace
			}
			summary.InfrastructureRef = ref
		}
	}

	// Extract status
	status, found, _ := unstructured.NestedMap(obj.Object, "status")
	if found {
		if ready, found, _ := unstructured.NestedBool(status, "ready"); found {
			summary.Ready = ready
		}
		if controlPlaneReady, found, _ := unstructured.NestedBool(status, "controlPlaneReady"); found {
			summary.ControlPlaneReady = controlPlaneReady
		}
		if infrastructureReady, found, _ := unstructured.NestedBool(status, "infrastructureReady"); found {
			summary.InfrastructureReady = infrastructureReady
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

// handleClustersList handles the capi.clusters_list tool.
func (t *Toolset) handleClustersList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	listOpts := metav1.ListOptions{}
	if args.LabelSelector != "" {
		listOpts.LabelSelector = args.LabelSelector
	}

	list, err := clientSet.Dynamic.Resource(t.clusterGVR).Namespace(args.Namespace).List(ctx, listOpts)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list Clusters: %w", err)), nil
	}

	clusters := make([]ClusterSummary, 0, len(list.Items))
	for _, item := range list.Items {
		clusters = append(clusters, t.normalizeClusterSummary(&item))
	}

	return mcpHelpers.NewJSONResult(map[string]any{"items": clusters})
}

// handleClusterGet handles the capi.cluster_get tool.
func (t *Toolset) handleClusterGet(ctx context.Context, args struct {
	Context   string `json:"context"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Raw       bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	if errResult, err := t.checkFeatureEnabled(); errResult != nil || err != nil {
		return errResult, err
	}

	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	obj, err := clientSet.Dynamic.Resource(t.clusterGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get Cluster: %w", err)), nil
	}

	summary := t.normalizeClusterSummary(obj)

	result := map[string]any{
		"summary": summary,
	}

	if args.Raw {
		result["raw_object"] = obj.Object
	}

	return mcpHelpers.NewJSONResult(result)
}
