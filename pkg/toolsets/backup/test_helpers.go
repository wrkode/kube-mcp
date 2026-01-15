package backup

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHandleBackupsList is a test helper that exposes handleBackupsList for testing.
func (t *Toolset) TestHandleBackupsList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	return t.handleBackupsList(ctx, args)
}

// TestHandleBackupGet is a test helper that exposes handleBackupGet for testing.
func (t *Toolset) TestHandleBackupGet(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Raw       bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	return t.handleBackupGet(ctx, args)
}
