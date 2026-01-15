package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/config"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	"github.com/wrkode/kube-mcp/pkg/observability"
	"github.com/wrkode/kube-mcp/pkg/toolsets/net"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetTestSuite tests Network toolset functionality.
type NetTestSuite struct {
	EnvtestSuite
	toolset *net.Toolset
}

// SetupTest sets up the test.
func (s *NetTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()

	crdDiscovery := kubernetes.NewCRDDiscovery(s.clientSet, 5*time.Minute)
	err := crdDiscovery.DiscoverCRDs(s.ctx)
	s.Require().NoError(err, "Failed to discover CRDs")

	// Create toolset with empty net config (no Hubble)
	netConfig := config.NetConfig{}
	s.toolset = net.NewToolset(s.provider, crdDiscovery, netConfig)
	s.Require().True(s.toolset.IsEnabled(), "Network toolset should be enabled")

	logger := observability.NewLogger(observability.LogLevelInfo, false)
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry)
	s.toolset.SetObservability(logger, metrics)
	s.toolset.SetRBACAuthorizer(nil, false)
}

// TestNetworkPoliciesList tests net.networkpolicies_list operation.
func (s *NetTestSuite) TestNetworkPoliciesList() {
	ctx := context.Background()
	namespace := "default"

	np := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-np",
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
		},
	}
	_, err := s.clientSet.Typed.NetworkingV1().NetworkPolicies(namespace).Create(ctx, np, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create NetworkPolicy")

	args := struct {
		Context       string `json:"context"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
	}{
		Namespace: namespace,
		Limit:     10,
	}

	result, err := s.toolset.TestHandleNetworkPoliciesList(ctx, args)
	s.Require().NoError(err, "NetworkPolicies list should succeed")
	s.Require().NotNil(result, "Result should not be nil")
	s.Require().False(result.IsError, "Result should not be an error")

	var resultData map[string]interface{}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "Result should be TextContent")
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "Should parse result JSON")

	items, ok := resultData["items"].([]interface{})
	s.Require().True(ok, "Result should have items array")
	s.Require().Greater(len(items), 0, "Should have at least one NetworkPolicy")
}

// TestNetTestSuite runs the network test suite.
func TestNetTestSuite(t *testing.T) {
	suite.Run(t, new(NetTestSuite))
}
