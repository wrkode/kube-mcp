package net

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wrkode/kube-mcp/pkg/config"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	"github.com/wrkode/kube-mcp/pkg/observability"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Toolset implements the Network toolset for NetworkPolicy, Cilium, and Hubble.
type Toolset struct {
	provider                          kubernetes.ClientProvider
	discovery                         *kubernetes.CRDDiscovery
	logger                            *observability.Logger
	metrics                           *observability.Metrics
	rbacAuthorizer                    kubernetes.RBACAuthorizer
	requireRBAC                       bool
	enabled                           bool // Always true (NetworkPolicy is native)
	hasCilium                         bool
	hasHubble                         bool
	ciliumNetworkPolicyGVR            schema.GroupVersionResource
	ciliumClusterwideNetworkPolicyGVR schema.GroupVersionResource
	hubbleClient                      *http.Client
	hubbleAPIURL                      string
}

// NewToolset creates a new Network toolset with CRD detection.
func NewToolset(provider kubernetes.ClientProvider, discovery *kubernetes.CRDDiscovery, netConfig config.NetConfig) *Toolset {
	enabled := true // NetworkPolicy is always available
	hasCilium := false
	hasHubble := netConfig.HubbleAPIURL != ""
	var ciliumNetworkPolicyGVR, ciliumClusterwideNetworkPolicyGVR schema.GroupVersionResource

	if discovery != nil {
		// Check for Cilium CRDs
		if gvr, ok := discovery.GetGVR(CiliumNetworkPolicyGVK); ok {
			hasCilium = true
			ciliumNetworkPolicyGVR = gvr
		}
		if gvr, ok := discovery.GetGVR(CiliumClusterwideNetworkPolicyGVK); ok {
			hasCilium = true
			ciliumClusterwideNetworkPolicyGVR = gvr
		}
	}

	// Setup Hubble client if configured
	var hubbleClient *http.Client
	if hasHubble {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: netConfig.HubbleInsecure,
			},
		}

		// Load CA cert if provided
		if netConfig.HubbleCAFile != "" {
			caCert, err := os.ReadFile(netConfig.HubbleCAFile)
			if err == nil {
				caCertPool := x509.NewCertPool()
				if caCertPool.AppendCertsFromPEM(caCert) {
					transport.TLSClientConfig.RootCAs = caCertPool
				}
			}
		}

		timeout := netConfig.HubbleTimeout.Duration()
		if timeout == 0 {
			timeout = 10 * time.Second
		}

		hubbleClient = &http.Client{
			Transport: transport,
			Timeout:   timeout,
		}
	}

	return &Toolset{
		provider:                          provider,
		discovery:                         discovery,
		enabled:                           enabled,
		hasCilium:                         hasCilium,
		hasHubble:                         hasHubble,
		ciliumNetworkPolicyGVR:            ciliumNetworkPolicyGVR,
		ciliumClusterwideNetworkPolicyGVR: ciliumClusterwideNetworkPolicyGVR,
		hubbleClient:                      hubbleClient,
		hubbleAPIURL:                      netConfig.HubbleAPIURL,
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

// IsEnabled returns whether the Network toolset is enabled.
func (t *Toolset) IsEnabled() bool {
	return t.enabled
}

// Name returns the toolset name.
func (t *Toolset) Name() string {
	return "net"
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
