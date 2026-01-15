package certs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	"github.com/wrkode/kube-mcp/pkg/observability"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Toolset implements the Cert-Manager toolset.
type Toolset struct {
	provider         kubernetes.ClientProvider
	discovery        *kubernetes.CRDDiscovery
	logger           *observability.Logger
	metrics          *observability.Metrics
	rbacAuthorizer   kubernetes.RBACAuthorizer
	requireRBAC      bool
	enabled          bool
	certificateGVR   schema.GroupVersionResource
	issuerGVR        schema.GroupVersionResource
	clusterIssuerGVR schema.GroupVersionResource
	certRequestGVR   schema.GroupVersionResource
	orderGVR         schema.GroupVersionResource
	challengeGVR     schema.GroupVersionResource
	hasCertificate   bool
	hasIssuer        bool
	hasClusterIssuer bool
	hasCertRequest   bool
	hasOrder         bool
	hasChallenge     bool
}

// NewToolset creates a new Cert-Manager toolset with CRD detection.
func NewToolset(provider kubernetes.ClientProvider, discovery *kubernetes.CRDDiscovery) *Toolset {
	enabled := false
	var certificateGVR, issuerGVR, clusterIssuerGVR, certRequestGVR, orderGVR, challengeGVR schema.GroupVersionResource
	hasCertificate, hasIssuer, hasClusterIssuer, hasCertRequest, hasOrder, hasChallenge := false, false, false, false, false, false

	if discovery != nil {
		// Certificate is required
		if gvr, ok := discovery.GetGVR(CertificateGVK); ok {
			enabled = true
			hasCertificate = true
			certificateGVR = gvr
		}

		// Optional CRDs
		if gvr, ok := discovery.GetGVR(IssuerGVK); ok {
			hasIssuer = true
			issuerGVR = gvr
		}
		if gvr, ok := discovery.GetGVR(ClusterIssuerGVK); ok {
			hasClusterIssuer = true
			clusterIssuerGVR = gvr
		}
		if gvr, ok := discovery.GetGVR(CertificateRequestGVK); ok {
			hasCertRequest = true
			certRequestGVR = gvr
		}
		if gvr, ok := discovery.GetGVR(OrderGVK); ok {
			hasOrder = true
			orderGVR = gvr
		}
		if gvr, ok := discovery.GetGVR(ChallengeGVK); ok {
			hasChallenge = true
			challengeGVR = gvr
		}
	}

	return &Toolset{
		provider:         provider,
		discovery:        discovery,
		enabled:          enabled,
		certificateGVR:   certificateGVR,
		issuerGVR:        issuerGVR,
		clusterIssuerGVR: clusterIssuerGVR,
		certRequestGVR:   certRequestGVR,
		orderGVR:         orderGVR,
		challengeGVR:     challengeGVR,
		hasCertificate:   hasCertificate,
		hasIssuer:        hasIssuer,
		hasClusterIssuer: hasClusterIssuer,
		hasCertRequest:   hasCertRequest,
		hasOrder:         hasOrder,
		hasChallenge:     hasChallenge,
	}
}

// SetObservability sets the observability components for the toolset.
func (t *Toolset) SetObservability(logger *observability.Logger, metrics *observability.Metrics) {
	t.logger = logger
	t.metrics = metrics
}

// SetRBACAuthorizer sets the RBAC authorizer for the toolset.
func (t *Toolset) SetRBACAuthorizer(authorizer kubernetes.RBACAuthorizer, requireRBAC bool) {
	t.rbacAuthorizer = authorizer
	t.requireRBAC = requireRBAC
}

// IsEnabled returns whether the Cert-Manager toolset is enabled.
func (t *Toolset) IsEnabled() bool {
	return t.enabled
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "certs"
}

// unmarshalArgs unmarshals args from map[string]interface{} to the target struct type.
func unmarshalArgs[T any](args any) (T, error) {
	var result T
	if args == nil {
		return result, nil
	}

	if typed, ok := args.(T); ok {
		return typed, nil
	}

	if argsMap, ok := args.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(argsMap)
		if err != nil {
			return result, err
		}
		err = json.Unmarshal(jsonData, &result)
		return result, err
	}

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
				"message": "Cert-Manager toolset is not enabled",
				"details": "Required CRD (cert-manager.io/v1/Certificate) not detected in cluster",
			},
		})
		return result, err
	}
	return nil, nil
}

// checkRBAC performs an RBAC check before an operation.
func (t *Toolset) checkRBAC(ctx context.Context, clientSet *kubernetes.ClientSet, verb string, gvr schema.GroupVersionResource, namespace string) (*mcp.CallToolResult, error) {
	if !t.requireRBAC || t.rbacAuthorizer == nil {
		return nil, nil
	}

	user := ""
	allowed, err := t.rbacAuthorizer.Allowed(ctx, user, verb, gvr, namespace)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to check RBAC: %w", err)), nil
	}

	if !allowed {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"code":    "KubernetesError",
				"message": fmt.Sprintf("Forbidden: user does not have permission to %s %s/%s in namespace %s", verb, gvr.Group, gvr.Resource, namespace),
				"details": map[string]any{
					"verb":      verb,
					"group":     gvr.Group,
					"resource":  gvr.Resource,
					"namespace": namespace,
					"reason":    "Forbidden",
				},
			},
		})
		return result, err
	}

	return nil, nil
}
