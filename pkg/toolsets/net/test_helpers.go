package net

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHandleNetworkPoliciesList is a test helper that exposes handleNetworkPoliciesList for testing.
func (t *Toolset) TestHandleNetworkPoliciesList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	return t.handleNetworkPoliciesList(ctx, args)
}

// TestHandleNetworkPolicyExplain is a test helper that exposes handleNetworkPolicyExplain for testing.
func (t *Toolset) TestHandleNetworkPolicyExplain(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}) (*mcp.CallToolResult, error) {
	return t.handleNetworkPolicyExplain(ctx, args)
}
