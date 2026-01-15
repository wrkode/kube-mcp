package core

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// handleEventsList handles the events_list tool.
func (t *Toolset) handleEventsList(ctx context.Context, args struct {
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	events, err := clientSet.Typed.CoreV1().Events(args.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list events: %w", err)), nil
	}

	eventList := make([]map[string]any, 0, len(events.Items))
	for _, event := range events.Items {
		eventList = append(eventList, map[string]any{
			"name":          event.Name,
			"namespace":     event.Namespace,
			"type":          event.Type,
			"reason":        event.Reason,
			"message":       event.Message,
			"involved_kind": event.InvolvedObject.Kind,
			"involved_name": event.InvolvedObject.Name,
			"first_seen":    event.FirstTimestamp.Time.Format("2006-01-02T15:04:05Z"),
			"last_seen":     event.LastTimestamp.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	return mcpHelpers.NewJSONResult(map[string]any{"events": eventList})
}
