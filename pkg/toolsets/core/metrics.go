package core

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// metricsAvailable checks if metrics API is available.
func (t *Toolset) metricsAvailable(clientSet *kubernetes.ClientSet) bool {
	return clientSet.Metrics != nil
}

// handlePodsTop handles the pods_top tool with real metrics API implementation.
func (t *Toolset) handlePodsTop(ctx context.Context, args struct {
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Check if metrics API is available
	if !t.metricsAvailable(clientSet) {
		return mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "MetricsUnavailable",
				"message": "Metrics server is not available",
				"details": "The metrics.k8s.io API is not available. Ensure metrics-server is installed and running.",
			},
			"metrics_available": false,
		})
	}

	// Get pod metrics
	metrics, err := clientSet.Metrics.MetricsV1beta1().PodMetricses(args.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "MetricsError",
				"message": fmt.Sprintf("Failed to retrieve pod metrics: %v", err),
				"details": err.Error(),
			},
			"metrics_available": false,
		})
	}

	// Process metrics
	podMetrics := make([]map[string]any, 0, len(metrics.Items))
	for _, metric := range metrics.Items {
		podData := map[string]any{
			"name":       metric.Name,
			"namespace":  metric.Namespace,
			"containers": make([]map[string]any, 0, len(metric.Containers)),
		}

		var totalCPU int64
		var totalMemory int64

		for _, container := range metric.Containers {
			cpu := container.Usage.Cpu().MilliValue()
			memory := container.Usage.Memory().Value()

			totalCPU += cpu
			totalMemory += memory

			podData["containers"] = append(podData["containers"].([]map[string]any), map[string]any{
				"name":   container.Name,
				"cpu_m":  cpu,
				"memory": memory,
			})
		}

		podData["cpu_m"] = totalCPU
		podData["memory"] = totalMemory

		podMetrics = append(podMetrics, podData)
	}

	return mcpHelpers.NewJSONResult(map[string]any{
		"pods":              podMetrics,
		"metrics_available": true,
	})
}

// handleNodesTop handles the nodes_top tool with real metrics API implementation.
func (t *Toolset) handleNodesTop(ctx context.Context, args struct {
	Context string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Check if metrics API is available
	if !t.metricsAvailable(clientSet) {
		return mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "MetricsUnavailable",
				"message": "Metrics server is not available",
				"details": "The metrics.k8s.io API is not available. Ensure metrics-server is installed and running.",
			},
			"metrics_available": false,
		})
	}

	// Get node metrics
	metrics, err := clientSet.Metrics.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "MetricsError",
				"message": fmt.Sprintf("Failed to retrieve node metrics: %v", err),
				"details": err.Error(),
			},
			"metrics_available": false,
		})
	}

	// Get node allocatable resources for percentage calculation
	nodes, err := clientSet.Typed.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list nodes: %w", err)), nil
	}

	// Create a map of node allocatable resources
	allocatableMap := make(map[string]map[string]int64)
	for _, node := range nodes.Items {
		allocatableMap[node.Name] = map[string]int64{
			"cpu":    node.Status.Allocatable.Cpu().MilliValue(),
			"memory": node.Status.Allocatable.Memory().Value(),
		}
	}

	// Process metrics
	nodeMetrics := make([]map[string]any, 0, len(metrics.Items))
	for _, metric := range metrics.Items {
		cpu := metric.Usage.Cpu().MilliValue()
		memory := metric.Usage.Memory().Value()

		nodeData := map[string]any{
			"name":   metric.Name,
			"cpu_m":  cpu,
			"memory": memory,
		}

		// Calculate percentages if allocatable is available
		if allocatable, ok := allocatableMap[metric.Name]; ok {
			if allocatable["cpu"] > 0 {
				cpuPercent := float64(cpu) / float64(allocatable["cpu"]) * 100
				nodeData["cpu_percent"] = cpuPercent
			}
			if allocatable["memory"] > 0 {
				memoryPercent := float64(memory) / float64(allocatable["memory"]) * 100
				nodeData["memory_percent"] = memoryPercent
			}
			nodeData["cpu_allocatable_m"] = allocatable["cpu"]
			nodeData["memory_allocatable"] = allocatable["memory"]
		}

		nodeMetrics = append(nodeMetrics, nodeData)
	}

	return mcpHelpers.NewJSONResult(map[string]any{
		"nodes":             nodeMetrics,
		"metrics_available": true,
	})
}
