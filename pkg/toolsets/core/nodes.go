package core

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// handleNodesTop is now implemented in metrics.go
// This function is kept for backward compatibility but delegates to the metrics implementation

// handleNodesSummary handles the nodes_summary tool.
func (t *Toolset) handleNodesSummary(ctx context.Context, args struct {
	Name    string `json:"name"`
	Context string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	nodes, err := clientSet.Typed.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list nodes: %w", err)), nil
	}

	summaries := make([]map[string]any, 0)
	for _, node := range nodes.Items {
		if args.Name != "" && node.Name != args.Name {
			continue
		}

		status := "Unknown"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		summary := map[string]any{
			"name":               node.Name,
			"status":             status,
			"cpu_capacity":       node.Status.Capacity.Cpu().Value(),
			"memory_capacity":    node.Status.Capacity.Memory().Value(),
			"cpu_allocatable":    node.Status.Allocatable.Cpu().Value(),
			"memory_allocatable": node.Status.Allocatable.Memory().Value(),
		}

		summaries = append(summaries, summary)
	}

	return mcpHelpers.NewJSONResult(map[string]any{"nodes": summaries})
}
