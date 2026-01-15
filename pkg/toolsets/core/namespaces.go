package core

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// handleNamespacesList handles the namespaces_list tool.
func (t *Toolset) handleNamespacesList(ctx context.Context, args struct {
	Context string `json:"context"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	namespaces, err := clientSet.Typed.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list namespaces: %w", err)), nil
	}

	nsList := make([]map[string]any, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		nsList = append(nsList, map[string]any{
			"name":       ns.Name,
			"status":     string(ns.Status.Phase),
			"created_at": ns.CreationTimestamp.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	return mcpHelpers.NewJSONResult(map[string]any{"namespaces": nsList})
}
