package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handlePodsPortForward handles the pods_port_forward tool.
// Note: Full port forwarding requires HTTP transport (Streamable HTTP) for long-lived connections.
// This implementation validates the request and returns connection setup information.
func (t *Toolset) handlePodsPortForward(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	LocalPort int    `json:"local_port"`
	PodPort   int    `json:"pod_port"`
	Container string `json:"container"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	// Validate ports
	if args.LocalPort <= 0 || args.LocalPort > 65535 {
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid local_port: must be between 1 and 65535")), nil
	}
	if args.PodPort <= 0 || args.PodPort > 65535 {
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid pod_port: must be between 1 and 65535")), nil
	}

	// Check RBAC
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	if rbacResult, rbacErr := t.checkRBAC(ctx, clientSet, "get", gvr, args.Namespace); rbacErr != nil || rbacResult != nil {
		if rbacResult != nil {
			return rbacResult, nil
		}
		return mcpHelpers.NewErrorResult(rbacErr), nil
	}

	// Verify pod exists and get pod details
	pod, err := clientSet.Typed.CoreV1().Pods(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get pod: %w", err)), nil
	}

	// Check if pod is ready for port forwarding
	if pod.Status.Phase != corev1.PodRunning {
		return mcpHelpers.NewErrorResult(fmt.Errorf("pod is not running (current phase: %s)", pod.Status.Phase)), nil
	}

	// Return port forward setup information
	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"status":          "configured",
		"local_port":      args.LocalPort,
		"pod_port":        args.PodPort,
		"pod":             args.Name,
		"namespace":       args.Namespace,
		"connection":      fmt.Sprintf("localhost:%d -> %s/%s:%d", args.LocalPort, args.Namespace, args.Name, args.PodPort),
		"note":            "Port forwarding setup validated. For active port forwarding, use HTTP transport endpoint or kubectl port-forward directly.",
		"kubectl_command": fmt.Sprintf("kubectl port-forward %s/%s %d:%d", args.Namespace, args.Name, args.LocalPort, args.PodPort),
	})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", err)), nil
	}
	return result, nil
}

// handleResourcesWatch handles the resources_watch tool.
func (t *Toolset) handleResourcesWatch(ctx context.Context, args struct {
	Group         string `json:"group"`
	Version       string `json:"version"`
	Kind          string `json:"kind"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	FieldSelector string `json:"field_selector"`
	Timeout       int    `json:"timeout"` // Timeout in seconds (0 = no timeout, use context)
	Context       string `json:"context"`
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

	// Build watch options
	watchOptions := metav1.ListOptions{}
	if args.LabelSelector != "" {
		watchOptions.LabelSelector = args.LabelSelector
	}
	if args.FieldSelector != "" {
		watchOptions.FieldSelector = args.FieldSelector
	}

	// Set timeout if specified
	if args.Timeout > 0 {
		watchOptions.TimeoutSeconds = int64Ptr(int64(args.Timeout))
	}

	// Create watch
	watcher, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Watch(ctx, watchOptions)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create watch: %w", err)), nil
	}
	defer watcher.Stop()

	// Collect events (with timeout)
	events := make([]map[string]any, 0)
	timeout := time.Duration(args.Timeout) * time.Second
	if args.Timeout == 0 {
		timeout = 30 * time.Second // Default 30 second watch
	}

	watchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Watch for changes (watch API returns initial state as ADDED events, so we don't need to list separately)
	for {
		select {
		case <-watchCtx.Done():
			// Timeout or context cancelled
			result, err := mcpHelpers.NewJSONResult(map[string]any{
				"events":      events,
				"event_count": len(events),
				"gvk":         gvk.String(),
				"namespace":   args.Namespace,
				"note":        "Watch completed. For continuous streaming, use HTTP transport endpoint.",
			})
			if err != nil {
				return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", err)), nil
			}
			return result, nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				// Channel closed
				result, err := mcpHelpers.NewJSONResult(map[string]any{
					"events":      events,
					"event_count": len(events),
					"gvk":         gvk.String(),
					"namespace":   args.Namespace,
					"note":        "Watch channel closed.",
				})
				if err != nil {
					return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", err)), nil
				}
				return result, nil
			}

			// Process event
			eventData := map[string]any{
				"type": string(event.Type),
			}

			// Convert object to map for JSON serialization
			if obj, ok := event.Object.(metav1.Object); ok {
				eventData["name"] = obj.GetName()
				eventData["namespace"] = obj.GetNamespace()
				eventData["uid"] = string(obj.GetUID())
			}

			// Try to get object as unstructured for full data
			if unstructuredObj, ok := event.Object.(*unstructured.Unstructured); ok {
				eventData["object"] = unstructuredObj.Object
			} else {
				// Fallback: marshal to JSON and unmarshal to map
				jsonBytes, err := json.Marshal(event.Object)
				if err == nil {
					var objMap map[string]any
					if err := json.Unmarshal(jsonBytes, &objMap); err == nil {
						eventData["object"] = objMap
					} else {
						eventData["object"] = map[string]any{
							"kind": event.Object.GetObjectKind().GroupVersionKind().Kind,
						}
					}
				} else {
					eventData["object"] = map[string]any{
						"kind": event.Object.GetObjectKind().GroupVersionKind().Kind,
					}
				}
			}

			events = append(events, eventData)

			// Limit events to prevent memory issues
			if len(events) >= 1000 {
				result, err := mcpHelpers.NewJSONResult(map[string]any{
					"events":      events,
					"event_count": len(events),
					"gvk":         gvk.String(),
					"namespace":   args.Namespace,
					"note":        "Event limit reached (1000 events). Watch stopped.",
				})
				if err != nil {
					return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create result: %w", err)), nil
				}
				return result, nil
			}
		}
	}
}

// int64Ptr returns a pointer to an int64.
func int64Ptr(i int64) *int64 {
	return &i
}
