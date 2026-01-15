package autoscaling

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHandleHPAList is a test helper that exposes handleHPAList for testing.
func (t *Toolset) TestHandleHPAList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	return t.handleHPAList(ctx, args)
}

// TestHandleHPAExplain is a test helper that exposes handleHPAExplain for testing.
func (t *Toolset) TestHandleHPAExplain(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}) (*mcp.CallToolResult, error) {
	return t.handleHPAExplain(ctx, args)
}
