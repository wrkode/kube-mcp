package core

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHandlePodsList is a test helper that exposes handlePodsList for testing.
func (t *Toolset) TestHandlePodsList(ctx context.Context, args struct {
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	FieldSelector string `json:"field_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
	Context       string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handlePodsList(ctx, args)
}

// TestHandlePodsGet is a test helper that exposes handlePodsGet for testing.
func (t *Toolset) TestHandlePodsGet(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handlePodsGet(ctx, args)
}

// TestHandlePodsDelete is a test helper that exposes handlePodsDelete for testing.
func (t *Toolset) TestHandlePodsDelete(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handlePodsDelete(ctx, args)
}

// TestHandlePodsLogs is a test helper that exposes handlePodsLogs for testing.
func (t *Toolset) TestHandlePodsLogs(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Container string `json:"container"`
	TailLines *int   `json:"tail_lines"`
	Since     string `json:"since"`
	SinceTime string `json:"since_time"`
	Previous  bool   `json:"previous"`
	Follow    bool   `json:"follow"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handlePodsLogs(ctx, args)
}

// TestHandleNamespacesList is a test helper that exposes handleNamespacesList for testing.
func (t *Toolset) TestHandleNamespacesList(ctx context.Context, args struct {
	Context string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleNamespacesList(ctx, args)
}

// TestHandleResourcesApply is a test helper that exposes handleResourcesApply for testing.
func (t *Toolset) TestHandleResourcesApply(ctx context.Context, args struct {
	Manifest     map[string]any `json:"manifest"`
	FieldManager string         `json:"field_manager"`
	DryRun       bool           `json:"dry_run"`
	Context      string         `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleResourcesApply(ctx, args)
}

// TestHandleResourcesPatch is a test helper that exposes handleResourcesPatch for testing.
func (t *Toolset) TestHandleResourcesPatch(ctx context.Context, args struct {
	Group        string      `json:"group"`
	Version      string      `json:"version"`
	Kind         string      `json:"kind"`
	Name         string      `json:"name"`
	Namespace    string      `json:"namespace"`
	PatchType    string      `json:"patch_type"`
	PatchData    interface{} `json:"patch_data"` // object for merge/strategic, array for json patch
	FieldManager string      `json:"field_manager"`
	DryRun       bool        `json:"dry_run"`
	Context      string      `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleResourcesPatch(ctx, args)
}

// TestHandleResourcesDiff is a test helper that exposes handleResourcesDiff for testing.
func (t *Toolset) TestHandleResourcesDiff(ctx context.Context, args struct {
	Group      string                 `json:"group"`
	Version    string                 `json:"version"`
	Kind       string                 `json:"kind"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	Manifest   map[string]interface{} `json:"manifest"`
	DiffFormat string                 `json:"diff_format"`
	Context    string                 `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleResourcesDiff(ctx, args)
}

// TestHandleResourcesValidate is a test helper that exposes handleResourcesValidate for testing.
func (t *Toolset) TestHandleResourcesValidate(ctx context.Context, args struct {
	Manifest     map[string]interface{} `json:"manifest"`
	SchemaVersion string                `json:"schema_version"`
	Context      string                 `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleResourcesValidate(ctx, args)
}

// TestHandleResourcesRelationships is a test helper that exposes handleResourcesRelationships for testing.
func (t *Toolset) TestHandleResourcesRelationships(ctx context.Context, args struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Direction string `json:"direction"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleResourcesRelationships(ctx, args)
}

// TestHandleConfigMapsGetData is a test helper that exposes handleConfigMapsGetData for testing.
func (t *Toolset) TestHandleConfigMapsGetData(ctx context.Context, args struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Keys      []string `json:"keys"`
	Context   string   `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleConfigMapsGetData(ctx, args)
}

// TestHandleConfigMapsSetData is a test helper that exposes handleConfigMapsSetData for testing.
func (t *Toolset) TestHandleConfigMapsSetData(ctx context.Context, args struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"`
	Merge     bool              `json:"merge"`
	Context   string            `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleConfigMapsSetData(ctx, args)
}

// TestHandleSecretsGetData is a test helper that exposes handleSecretsGetData for testing.
func (t *Toolset) TestHandleSecretsGetData(ctx context.Context, args struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Keys      []string `json:"keys"`
	Decode    bool     `json:"decode"`
	Context   string   `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleSecretsGetData(ctx, args)
}

// TestHandleSecretsSetData is a test helper that exposes handleSecretsSetData for testing.
func (t *Toolset) TestHandleSecretsSetData(ctx context.Context, args struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"`
	Merge     bool              `json:"merge"`
	Encode    bool              `json:"encode"`
	Context   string            `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleSecretsSetData(ctx, args)
}

// TestHandleResourcesScale is a test helper that exposes handleResourcesScale for testing.
func (t *Toolset) TestHandleResourcesScale(ctx context.Context, args struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  *int   `json:"replicas"` // nil = get-only, 0 = scale to zero, >0 = scale to that number
	DryRun    bool   `json:"dry_run"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleResourcesScale(ctx, args)
}

// TestHandlePodsPortForward is a test helper that exposes handlePodsPortForward for testing.
func (t *Toolset) TestHandlePodsPortForward(ctx context.Context, args struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	LocalPort int    `json:"local_port"`
	PodPort   int    `json:"pod_port"`
	Container string `json:"container"`
	Context   string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handlePodsPortForward(ctx, args)
}

// TestHandleResourcesWatch is a test helper that exposes handleResourcesWatch for testing.
func (t *Toolset) TestHandleResourcesWatch(ctx context.Context, args struct {
	Group         string `json:"group"`
	Version       string `json:"version"`
	Kind          string `json:"kind"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	FieldSelector string `json:"field_selector"`
	Timeout       int    `json:"timeout"`
	Context       string `json:"context"`
}) (*mcp.CallToolResult, error) {
	return t.handleResourcesWatch(ctx, args)
}
