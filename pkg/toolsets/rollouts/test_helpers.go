package rollouts

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHandleRolloutsList is a test helper that exposes handleRolloutsList for testing.
func (t *Toolset) TestHandleRolloutsList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	return t.handleRolloutsList(ctx, args)
}

// TestHandleRolloutGetStatus is a test helper that exposes handleRolloutGetStatus for testing.
func (t *Toolset) TestHandleRolloutGetStatus(ctx context.Context, args struct {
	Context   string `json:"context"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Raw       bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	return t.handleRolloutGetStatus(ctx, args)
}

// TestHandleRolloutPromote is a test helper that exposes handleRolloutPromote for testing.
func (t *Toolset) TestHandleRolloutPromote(ctx context.Context, args struct {
	Context   string `json:"context"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	return t.handleRolloutPromote(ctx, args)
}
