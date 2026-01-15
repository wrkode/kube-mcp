package core

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handleResourcesDescribe handles the resources_describe tool.
func (t *Toolset) handleResourcesDescribe(ctx context.Context, args struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
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
	resource, err := clientSet.Dynamic.Resource(gvr).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get resource: %w", err)), nil
	}

	// Get events for this resource
	events, err := t.getResourceEvents(ctx, clientSet, args.Namespace, gvk.Kind, args.Name)
	if err != nil {
		// Don't fail if events can't be fetched, just continue without them
		events = []corev1.Event{}
	}

	// Format describe output
	output := t.formatDescribeOutput(resource, events)

	return mcpHelpers.NewTextResult(output), nil
}

// getResourceEvents fetches events related to a specific resource.
func (t *Toolset) getResourceEvents(ctx context.Context, clientSet *kubernetes.ClientSet, namespace, kind, name string) ([]corev1.Event, error) {
	// List all events in the namespace
	events, err := clientSet.Typed.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Filter events for this resource
	filteredEvents := make([]corev1.Event, 0)
	for _, event := range events.Items {
		if event.InvolvedObject.Kind == kind && event.InvolvedObject.Name == name {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// Sort by last timestamp (newest first), limit to 20 most recent
	sort.Slice(filteredEvents, func(i, j int) bool {
		return filteredEvents[i].LastTimestamp.After(filteredEvents[j].LastTimestamp.Time)
	})
	if len(filteredEvents) > 20 {
		filteredEvents = filteredEvents[:20]
	}

	return filteredEvents, nil
}

// formatDescribeOutput formats a resource and its events into kubectl-style describe output.
func (t *Toolset) formatDescribeOutput(resource *unstructured.Unstructured, events []corev1.Event) string {
	var buf strings.Builder

	// Basic metadata
	name := resource.GetName()
	namespace := resource.GetNamespace()
	kind := resource.GetKind()
	apiVersion := resource.GetAPIVersion()

	buf.WriteString(fmt.Sprintf("Name:         %s\n", name))
	if namespace != "" {
		buf.WriteString(fmt.Sprintf("Namespace:    %s\n", namespace))
	}
	buf.WriteString(fmt.Sprintf("Labels:       %s\n", formatLabels(resource.GetLabels())))
	buf.WriteString(fmt.Sprintf("Annotations:  %s\n", formatAnnotations(resource.GetAnnotations())))
	buf.WriteString(fmt.Sprintf("API Version:  %s\n", apiVersion))
	buf.WriteString(fmt.Sprintf("Kind:         %s\n", kind))

	// Creation timestamp
	if creationTime := resource.GetCreationTimestamp(); !creationTime.IsZero() {
		buf.WriteString(fmt.Sprintf("Created:      %s\n", formatTime(creationTime.Time)))
	}

	buf.WriteString("\n")

	// Resource-specific formatting
	switch kind {
	case "Pod":
		t.formatPodDescribe(&buf, resource)
	case "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet":
		t.formatWorkloadDescribe(&buf, resource)
	case "Service":
		t.formatServiceDescribe(&buf, resource)
	default:
		t.formatGenericDescribe(&buf, resource)
	}

	// Events section
	if len(events) > 0 {
		buf.WriteString("\nEvents:\n")
		buf.WriteString(fmt.Sprintf("  %-8s %-12s %-8s %-12s %s\n", "Type", "Reason", "Age", "From", "Message"))
		buf.WriteString("  " + strings.Repeat("-", 80) + "\n")
		for _, event := range events {
			age := formatAge(event.LastTimestamp.Time)
			buf.WriteString(fmt.Sprintf("  %-8s %-12s %-8s %-12s %s\n",
				event.Type,
				event.Reason,
				age,
				event.Source.Component,
				event.Message))
		}
	}

	return buf.String()
}

// formatPodDescribe formats pod-specific describe output.
func (t *Toolset) formatPodDescribe(buf *strings.Builder, pod *unstructured.Unstructured) {
	// Status
	if status, found, _ := unstructured.NestedString(pod.Object, "status", "phase"); found {
		buf.WriteString(fmt.Sprintf("Status:       %s\n", status))
	}

	// Node
	if nodeName, found, _ := unstructured.NestedString(pod.Object, "spec", "nodeName"); found {
		buf.WriteString(fmt.Sprintf("Node:         %s\n", nodeName))
	}

	// Containers
	if containers, found, _ := unstructured.NestedSlice(pod.Object, "spec", "containers"); found {
		buf.WriteString("\nContainers:\n")
		for _, container := range containers {
			containerMap, ok := container.(map[string]interface{})
			if !ok {
				continue
			}
			containerName, _ := containerMap["name"].(string)
			image, _ := containerMap["image"].(string)
			buf.WriteString(fmt.Sprintf("  %s:\n", containerName))
			buf.WriteString(fmt.Sprintf("    Image:     %s\n", image))

			// Container status
			if statuses, found, _ := unstructured.NestedSlice(pod.Object, "status", "containerStatuses"); found {
				for _, status := range statuses {
					statusMap, ok := status.(map[string]interface{})
					if !ok {
						continue
					}
					if statusMap["name"] == containerName {
						if ready, ok := statusMap["ready"].(bool); ok {
							buf.WriteString(fmt.Sprintf("    Ready:     %v\n", ready))
						}
						if restartCount, ok := statusMap["restartCount"].(int64); ok {
							buf.WriteString(fmt.Sprintf("    Restarts:  %d\n", restartCount))
						}
						if state, ok := statusMap["state"].(map[string]interface{}); ok {
							if running, ok := state["running"].(map[string]interface{}); ok {
								if startedAt, ok := running["startedAt"].(string); ok {
									buf.WriteString(fmt.Sprintf("    Started:   %s\n", startedAt))
								}
							}
						}
						break
					}
				}
			}
		}
	}

	// Conditions
	if conditions, found, _ := unstructured.NestedSlice(pod.Object, "status", "conditions"); found {
		if len(conditions) > 0 {
			buf.WriteString("\nConditions:\n")
			buf.WriteString(fmt.Sprintf("  %-20s %-8s %-8s %s\n", "Type", "Status", "LastProbe", "LastTransition"))
			buf.WriteString("  " + strings.Repeat("-", 80) + "\n")
			for _, condition := range conditions {
				condMap, ok := condition.(map[string]interface{})
				if !ok {
					continue
				}
				condType, _ := condMap["type"].(string)
				status, _ := condMap["status"].(string)
				lastTransition, _ := condMap["lastTransitionTime"].(string)
				message, _ := condMap["message"].(string)
				buf.WriteString(fmt.Sprintf("  %-20s %-8s %-8s %s", condType, status, "", lastTransition))
				if message != "" {
					buf.WriteString(fmt.Sprintf("  %s", message))
				}
				buf.WriteString("\n")
			}
		}
	}
}

// formatWorkloadDescribe formats workload resource (Deployment, ReplicaSet, etc.) describe output.
func (t *Toolset) formatWorkloadDescribe(buf *strings.Builder, resource *unstructured.Unstructured) {
	// Replicas
	if replicas, found, _ := unstructured.NestedInt64(resource.Object, "status", "replicas"); found {
		buf.WriteString(fmt.Sprintf("Replicas:     %d\n", replicas))
	}
	if readyReplicas, found, _ := unstructured.NestedInt64(resource.Object, "status", "readyReplicas"); found {
		buf.WriteString(fmt.Sprintf("Ready:        %d\n", readyReplicas))
	}
	if availableReplicas, found, _ := unstructured.NestedInt64(resource.Object, "status", "availableReplicas"); found {
		buf.WriteString(fmt.Sprintf("Available:    %d\n", availableReplicas))
	}

	// Selector
	if selector, found, _ := unstructured.NestedMap(resource.Object, "spec", "selector", "matchLabels"); found {
		buf.WriteString(fmt.Sprintf("Selector:     %s\n", formatLabels(convertToStringMap(selector))))
	}

	// Conditions
	if conditions, found, _ := unstructured.NestedSlice(resource.Object, "status", "conditions"); found {
		if len(conditions) > 0 {
			buf.WriteString("\nConditions:\n")
			buf.WriteString(fmt.Sprintf("  %-20s %-8s %s\n", "Type", "Status", "Reason"))
			buf.WriteString("  " + strings.Repeat("-", 80) + "\n")
			for _, condition := range conditions {
				condMap, ok := condition.(map[string]interface{})
				if !ok {
					continue
				}
				condType, _ := condMap["type"].(string)
				status, _ := condMap["status"].(string)
				reason, _ := condMap["reason"].(string)
				buf.WriteString(fmt.Sprintf("  %-20s %-8s %s\n", condType, status, reason))
			}
		}
	}
}

// formatServiceDescribe formats service-specific describe output.
func (t *Toolset) formatServiceDescribe(buf *strings.Builder, svc *unstructured.Unstructured) {
	// Type
	if svcType, found, _ := unstructured.NestedString(svc.Object, "spec", "type"); found {
		buf.WriteString(fmt.Sprintf("Type:         %s\n", svcType))
	}

	// Cluster IP
	if clusterIP, found, _ := unstructured.NestedString(svc.Object, "spec", "clusterIP"); found {
		buf.WriteString(fmt.Sprintf("IP:           %s\n", clusterIP))
	}

	// Ports
	if ports, found, _ := unstructured.NestedSlice(svc.Object, "spec", "ports"); found {
		if len(ports) > 0 {
			buf.WriteString("\nPorts:\n")
			buf.WriteString(fmt.Sprintf("  %-10s %-10s %-20s %s\n", "Name", "Port", "TargetPort", "Protocol"))
			buf.WriteString("  " + strings.Repeat("-", 80) + "\n")
			for _, port := range ports {
				portMap, ok := port.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := portMap["name"].(string)
				portNum, _ := portMap["port"].(int64)
				targetPort, _ := portMap["targetPort"]
				protocol, _ := portMap["protocol"].(string)
				if protocol == "" {
					protocol = "TCP"
				}
				buf.WriteString(fmt.Sprintf("  %-10s %-10d %-20v %s\n", name, portNum, targetPort, protocol))
			}
		}
	}

	// Selector
	if selector, found, _ := unstructured.NestedMap(svc.Object, "spec", "selector"); found {
		buf.WriteString(fmt.Sprintf("\nSelector:     %s\n", formatLabels(convertToStringMap(selector))))
	}
}

// formatGenericDescribe formats generic resource describe output.
func (t *Toolset) formatGenericDescribe(buf *strings.Builder, resource *unstructured.Unstructured) {
	// Try to format spec and status sections
	if spec, found, _ := unstructured.NestedMap(resource.Object, "spec"); found && len(spec) > 0 {
		buf.WriteString("\nSpec:\n")
		buf.WriteString(formatMap(spec, 2))
	}

	if status, found, _ := unstructured.NestedMap(resource.Object, "status"); found && len(status) > 0 {
		buf.WriteString("\nStatus:\n")
		buf.WriteString(formatMap(status, 2))
	}
}

// Helper functions for formatting

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "<none>"
	}
	pairs := make([]string, 0, len(labels))
	for k, v := range labels {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ", ")
}

func formatAnnotations(annotations map[string]string) string {
	if len(annotations) == 0 {
		return "<none>"
	}
	pairs := make([]string, 0, len(annotations))
	for k, v := range annotations {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ", ")
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 -0700 MST")
}

func formatAge(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	return fmt.Sprintf("%dd", int(duration.Hours()/24))
}

func convertToStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}

func formatMap(m map[string]interface{}, indent int) string {
	var buf strings.Builder
	prefix := strings.Repeat("  ", indent)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		switch val := v.(type) {
		case map[string]interface{}:
			buf.WriteString(fmt.Sprintf("%s%s:\n", prefix, k))
			buf.WriteString(formatMap(val, indent+1))
		case []interface{}:
			buf.WriteString(fmt.Sprintf("%s%s: [%d items]\n", prefix, k, len(val)))
		default:
			buf.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, k, val))
		}
	}
	return buf.String()
}

