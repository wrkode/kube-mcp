package gitops

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHandleAppsList is a test helper that exposes handleAppsList for testing.
func (t *Toolset) TestHandleAppsList(ctx context.Context, args struct {
	Context       string   `json:"context"`
	Namespace     string   `json:"namespace"`
	LabelSelector string   `json:"label_selector"`
	Kinds         []string `json:"kinds"`
	Limit         int      `json:"limit"`
	Continue      string   `json:"continue"`
}) (*mcp.CallToolResult, error) {
	return t.handleAppsList(ctx, args)
}

// TestHandleAppGet is a test helper that exposes handleAppGet for testing.
func (t *Toolset) TestHandleAppGet(ctx context.Context, args struct {
	Context   string `json:"context"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Raw       bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	return t.handleAppGet(ctx, args)
}

// TestHandleAppReconcile is a test helper that exposes handleAppReconcile for testing.
func (t *Toolset) TestHandleAppReconcile(ctx context.Context, args struct {
	Context   string `json:"context"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	return t.handleAppReconcile(ctx, args)
}
