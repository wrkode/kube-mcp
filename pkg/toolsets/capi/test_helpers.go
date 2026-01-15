package capi

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHandleClustersList is a test helper that exposes handleClustersList for testing.
func (t *Toolset) TestHandleClustersList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
}) (*mcp.CallToolResult, error) {
	return t.handleClustersList(ctx, args)
}

// TestHandleClusterGet is a test helper that exposes handleClusterGet for testing.
func (t *Toolset) TestHandleClusterGet(ctx context.Context, args struct {
	Context   string `json:"context"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Raw       bool   `json:"raw"`
}) (*mcp.CallToolResult, error) {
	return t.handleClusterGet(ctx, args)
}

// TestHandleMachinesList is a test helper that exposes handleMachinesList for testing.
func (t *Toolset) TestHandleMachinesList(ctx context.Context, args struct {
	Context          string `json:"context"`
	ClusterNamespace string `json:"cluster_namespace"`
	ClusterName      string `json:"cluster_name"`
	Limit            int    `json:"limit"`
	Continue         string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	return t.handleMachinesList(ctx, args)
}

// TestHandleMachineDeploymentsList is a test helper that exposes handleMachineDeploymentsList for testing.
func (t *Toolset) TestHandleMachineDeploymentsList(ctx context.Context, args struct {
	Context          string `json:"context"`
	ClusterNamespace string `json:"cluster_namespace"`
	ClusterName      string `json:"cluster_name"`
}) (*mcp.CallToolResult, error) {
	return t.handleMachineDeploymentsList(ctx, args)
}

// TestHandleScaleMachineDeployment is a test helper that exposes handleScaleMachineDeployment for testing.
func (t *Toolset) TestHandleScaleMachineDeployment(ctx context.Context, args struct {
	Context   string `json:"context"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Replicas  int    `json:"replicas"`
	Confirm   bool   `json:"confirm"`
}) (*mcp.CallToolResult, error) {
	return t.handleScaleMachineDeployment(ctx, args)
}
