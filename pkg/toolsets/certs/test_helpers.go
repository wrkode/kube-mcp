package certs

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHandleCertificatesList is a test helper that exposes handleCertificatesList for testing.
func (t *Toolset) TestHandleCertificatesList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	return t.handleCertificatesList(ctx, args)
}

// TestHandleCertificateGet is a test helper that exposes handleCertificateGet for testing.
func (t *Toolset) TestHandleCertificateGet(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Raw       bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	return t.handleCertificateGet(ctx, args)
}
