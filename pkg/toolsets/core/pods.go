package core

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handlePodsList handles the pods_list tool.
func (t *Toolset) handlePodsList(ctx context.Context, args struct {
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	FieldSelector string `json:"field_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
	Context       string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Build list options with selectors and pagination
	listOptions := metav1.ListOptions{}
	if args.LabelSelector != "" {
		listOptions.LabelSelector = args.LabelSelector
	}
	if args.FieldSelector != "" {
		listOptions.FieldSelector = args.FieldSelector
	}
	if args.Limit > 0 {
		listOptions.Limit = int64(args.Limit)
	}
	if args.Continue != "" {
		listOptions.Continue = args.Continue
	}

	pods, err := clientSet.Typed.CoreV1().Pods(args.Namespace).List(ctx, listOptions)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list pods: %w", err)), nil
	}

	podList := make([]map[string]any, 0, len(pods.Items))
	for _, pod := range pods.Items {
		podList = append(podList, map[string]any{
			"name":       pod.Name,
			"namespace":  pod.Namespace,
			"status":     string(pod.Status.Phase),
			"node":       pod.Spec.NodeName,
			"created_at": pod.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	result := map[string]any{"pods": podList}
	if pods.Continue != "" {
		result["continue"] = pods.Continue
		result["has_more"] = true
	}

	return mcpHelpers.NewJSONResult(result)
}

// handlePodsGet handles the pods_get tool.
func (t *Toolset) handlePodsGet(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	pod, err := clientSet.Typed.CoreV1().Pods(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get pod: %w", err)), nil
	}

	podData := map[string]any{
		"name":       pod.Name,
		"namespace":  pod.Namespace,
		"status":     string(pod.Status.Phase),
		"node":       pod.Spec.NodeName,
		"created_at": pod.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
		"labels":     pod.Labels,
		"containers": make([]map[string]any, 0),
	}

	for _, container := range pod.Spec.Containers {
		podData["containers"] = append(podData["containers"].([]map[string]any), map[string]any{
			"name":  container.Name,
			"image": container.Image,
		})
	}

	return mcpHelpers.NewJSONResult(podData)
}

// handlePodsDelete handles the pods_delete tool.
func (t *Toolset) handlePodsDelete(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Check RBAC before deletion
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	if rbacResult, rbacErr := t.checkRBAC(ctx, clientSet, "delete", gvr, args.Namespace); rbacErr != nil || rbacResult != nil {
		if rbacResult != nil {
			return rbacResult, nil
		}
		return mcpHelpers.NewErrorResult(rbacErr), nil
	}

	err = clientSet.Typed.CoreV1().Pods(args.Namespace).Delete(ctx, args.Name, metav1.DeleteOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to delete pod: %w", err)), nil
	}

	return mcpHelpers.NewTextResult(fmt.Sprintf("Pod %s/%s deleted successfully", args.Namespace, args.Name)), nil
}

// handlePodsLogs handles the pods_logs tool.
func (t *Toolset) handlePodsLogs(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Container string `json:"container"`
	TailLines *int   `json:"tail_lines"`
	Since     string `json:"since"`
	SinceTime string `json:"since_time"`
	Previous  bool   `json:"previous"`
	Follow    bool   `json:"follow"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	opts := &corev1.PodLogOptions{}
	if args.Container != "" {
		opts.Container = args.Container
	}
	if args.TailLines != nil {
		tailLines := int64(*args.TailLines)
		opts.TailLines = &tailLines
	}
	if args.Since != "" {
		duration, err := time.ParseDuration(args.Since)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("invalid since duration: %w", err)), nil
		}
		sinceTime := metav1.NewTime(time.Now().Add(-duration))
		opts.SinceTime = &sinceTime
	}
	if args.SinceTime != "" {
		sinceTime, err := time.Parse(time.RFC3339, args.SinceTime)
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("invalid since_time format (expected RFC3339): %w", err)), nil
		}
		opts.SinceTime = &metav1.Time{Time: sinceTime}
	}
	if args.Previous {
		opts.Previous = true
	}
	if args.Follow {
		opts.Follow = true
	}

	req_logs := clientSet.Typed.CoreV1().Pods(args.Namespace).GetLogs(args.Name, opts)
	logs, err := req_logs.Stream(ctx)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get logs: %w", err)), nil
	}
	defer logs.Close()

	// If follow is enabled, we'll read available logs and note that streaming is available via HTTP
	// Note: Full streaming requires HTTP transport (Streamable HTTP) integration
	if args.Follow {
		// Read initial logs (non-blocking read with timeout context)
		logBytes := make([]byte, 0)
		buf := make([]byte, 4096)

		// Use a timeout context to limit initial read
		readCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		done := make(chan bool)
		go func() {
			for {
				select {
				case <-readCtx.Done():
					done <- true
					return
				default:
					n, err := logs.Read(buf)
					if n > 0 {
						logBytes = append(logBytes, buf[:n]...)
					}
					if err != nil {
						done <- true
						return
					}
				}
			}
		}()
		<-done

		// Return initial logs with note about streaming
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"logs":           string(logBytes),
			"follow":         true,
			"streaming_note": "Log following is enabled. For real-time streaming, use HTTP transport endpoint.",
		})
		if err != nil {
			return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", err)), nil
		}
		return result, nil
	}

	// Non-follow mode: read all logs
	logBytes := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			logBytes = append(logBytes, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return mcpHelpers.NewTextResult(string(logBytes)), nil
}

// handlePodsExec handles the pods_exec tool.
func (t *Toolset) handlePodsExec(ctx context.Context, args struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Container string   `json:"container"`
	Command   []string `json:"command"`
	Context   string   `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	req_exec := clientSet.Typed.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(args.Namespace).
		Name(args.Name).
		SubResource("exec")

	opts := &corev1.PodExecOptions{
		Command: args.Command,
		Stdout:  true,
		Stderr:  true,
	}
	if args.Container != "" {
		opts.Container = args.Container
	}

	req_exec = req_exec.VersionedParams(opts, metav1.ParameterCodec)

	exec, err := req_exec.Stream(ctx)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to exec: %w", err)), nil
	}
	defer exec.Close()

	output := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, err := exec.Read(buf)
		if n > 0 {
			output = append(output, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return mcpHelpers.NewTextResult(string(output)), nil
}

// handlePodsTop is now implemented in metrics.go
// This function is kept for backward compatibility but delegates to the metrics implementation
