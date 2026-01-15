package policy

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHandleViolationsList is a test helper that exposes handleViolationsList for testing.
func (t *Toolset) TestHandleViolationsList(ctx context.Context, args struct {
	Context   string `json:"context"`
	Namespace string `json:"namespace"`
	Engine    string `json:"engine"`
	Limit     int    `json:"limit"`
	Continue  string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	return t.handleViolationsList(ctx, args)
}
