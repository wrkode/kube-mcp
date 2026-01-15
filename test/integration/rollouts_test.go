package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	"github.com/wrkode/kube-mcp/pkg/observability"
	"github.com/wrkode/kube-mcp/pkg/toolsets/rollouts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RolloutsTestSuite tests Progressive Delivery toolset functionality.
type RolloutsTestSuite struct {
	EnvtestSuite
	toolset *rollouts.Toolset
}

// SetupTest sets up the test.
func (s *RolloutsTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()

	// Create CRD discovery
	crdDiscovery := kubernetes.NewCRDDiscovery(s.clientSet, 5*time.Minute)
	err := crdDiscovery.DiscoverCRDs(s.ctx)
	s.Require().NoError(err, "Failed to discover CRDs")

	// Create toolset
	s.toolset = rollouts.NewToolset(s.provider, crdDiscovery)
	s.Require().True(s.toolset.IsEnabled(), "Rollouts toolset should be enabled")

	// Set observability
	logger := observability.NewLogger(observability.LogLevelInfo, false)
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry)
	s.toolset.SetObservability(logger, metrics)

	// Disable RBAC for tests
	s.toolset.SetRBACAuthorizer(nil, false)
}

// TestRolloutsList tests rollouts.list operation.
func (s *RolloutsTestSuite) TestRolloutsList() {
	ctx := context.Background()
	namespace := "rollouts-system"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Rollout
	rolloutGVR := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "rollouts",
	}

	rollout := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "Rollout",
			"metadata": map[string]interface{}{
				"name":      "test-rollout",
				"namespace": namespace,
			},
			"status": map[string]interface{}{
				"phase": "Progressing",
			},
		},
	}
	_, err = s.clientSet.Dynamic.Resource(rolloutGVR).Namespace(namespace).Create(ctx, rollout, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create Rollout")

	// Test list
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

	result, err := s.toolset.TestHandleRolloutsList(ctx, args)
	s.Require().NoError(err, "Rollouts list should succeed")
	s.Require().NotNil(result, "Result should not be nil")
	s.Require().False(result.IsError, "Result should not be an error")

	// Parse result
	var resultData map[string]interface{}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "Result should be TextContent")
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "Should parse result JSON")

	items, ok := resultData["items"].([]interface{})
	s.Require().True(ok, "Result should have items array")
	s.Require().Greater(len(items), 0, "Should have at least one rollout")
}

// TestRolloutsTestSuite runs the rollouts test suite.
func TestRolloutsTestSuite(t *testing.T) {
	suite.Run(t, new(RolloutsTestSuite))
}
