package policy

import (
	"context"
	"encoding/json"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Toolset implements the Policy toolset for Kyverno and Gatekeeper.
type Toolset struct {
	provider               kubernetes.ClientProvider
	discovery              *kubernetes.CRDDiscovery
	logger                 *observability.Logger
	metrics                *observability.Metrics
	enabled                bool
	hasKyverno             bool
	hasGatekeeper          bool
	clusterPolicyGVR       schema.GroupVersionResource
	policyGVR              schema.GroupVersionResource
	policyReportGVR        schema.GroupVersionResource
	clusterPolicyReportGVR schema.GroupVersionResource
	constraintTemplateGVR  schema.GroupVersionResource
}

// NewToolset creates a new Policy toolset with CRD detection.
func NewToolset(provider kubernetes.ClientProvider, discovery *kubernetes.CRDDiscovery) *Toolset {
	enabled := false
	hasKyverno := false
	hasGatekeeper := false
	var clusterPolicyGVR, policyGVR, policyReportGVR, clusterPolicyReportGVR, constraintTemplateGVR schema.GroupVersionResource

	if discovery != nil {
		// Check for Kyverno ClusterPolicy or Policy
		if gvr, ok := discovery.GetGVR(KyvernoClusterPolicyGVK); ok {
			enabled = true
			hasKyverno = true
			clusterPolicyGVR = gvr
		}
		if gvr, ok := discovery.GetGVR(KyvernoPolicyGVK); ok {
			enabled = true
			hasKyverno = true
			policyGVR = gvr
		}

		// Check for PolicyReport (optional but enhances functionality)
		if gvr, ok := discovery.GetGVR(KyvernoPolicyReportGVK); ok {
			policyReportGVR = gvr
		}
		if gvr, ok := discovery.GetGVR(KyvernoClusterPolicyReportGVK); ok {
			clusterPolicyReportGVR = gvr
		}

		// Check for Gatekeeper ConstraintTemplate
		if gvr, ok := discovery.GetGVR(GatekeeperConstraintTemplateGVK); ok {
			enabled = true
			hasGatekeeper = true
			constraintTemplateGVR = gvr
		}
	}

	return &Toolset{
		provider:               provider,
		discovery:              discovery,
		enabled:                enabled,
		hasKyverno:             hasKyverno,
		hasGatekeeper:          hasGatekeeper,
		clusterPolicyGVR:       clusterPolicyGVR,
		policyGVR:              policyGVR,
		policyReportGVR:        policyReportGVR,
		clusterPolicyReportGVR: clusterPolicyReportGVR,
		constraintTemplateGVR:  constraintTemplateGVR,
	}
}

// SetObservability sets the observability components for the toolset.
func (t *Toolset) SetObservability(logger *observability.Logger, metrics *observability.Metrics) {
	t.logger = logger
	t.metrics = metrics
}

// IsEnabled returns whether the Policy toolset is enabled.
func (t *Toolset) IsEnabled() bool {
	return t.enabled
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "policy"
}

// unmarshalArgs unmarshals args from map[string]interface{} to the target struct type.
func unmarshalArgs[T any](args any) (T, error) {
	var result T
	if args == nil {
		return result, nil
	}

	// If args is already the correct type, return it
	if typed, ok := args.(T); ok {
		return typed, nil
	}

	// If args is a map, unmarshal it
	if argsMap, ok := args.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(argsMap)
		if err != nil {
			return result, err
		}
		err = json.Unmarshal(jsonData, &result)
		return result, err
	}

	// Try to marshal and unmarshal as a fallback
	jsonData, err := json.Marshal(args)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(jsonData, &result)
	return result, err
}

// wrapToolHandler wraps a tool handler with observability.
func (t *Toolset) wrapToolHandler(
	toolName string,
	handler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error),
	getCluster func(args any) string,
) func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
	if t.logger == nil || t.metrics == nil {
		return handler
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		start := time.Now()
		cluster := getCluster(args)
		if cluster == "" {
			cluster = "default"
		}

		defer func() {
			if r := recover(); r != nil {
				t.logger.Error(ctx, "Panic in tool handler",
					"tool", toolName,
					"panic", r,
					"cluster", cluster,
				)
			}
		}()

		result, out, err := handler(ctx, req, args)

		duration := time.Since(start)
		t.logger.LogToolInvocation(ctx, toolName, cluster, duration, err)
		success := err == nil && (result == nil || !result.IsError)
		t.metrics.RecordToolCall(toolName, cluster, success, duration.Seconds())

		return result, out, err
	}
}

// checkFeatureEnabled checks if the toolset is enabled and returns an error if not.
func (t *Toolset) checkFeatureEnabled() (*mcp.CallToolResult, error) {
	if !t.enabled {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureNotInstalled",
				"message": "Policy toolset is not enabled",
				"details": "Required Policy CRDs (Kyverno ClusterPolicy/Policy or Gatekeeper ConstraintTemplate) not detected in cluster",
			},
		})
		return result, err
	}
	return nil, nil
}
