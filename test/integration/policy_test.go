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
	"github.com/wrkode/kube-mcp/pkg/toolsets/policy"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PolicyTestSuite tests Policy toolset functionality.
type PolicyTestSuite struct {
	EnvtestSuite
	toolset *policy.Toolset
}

// SetupTest sets up the test.
func (s *PolicyTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()

	// Create CRD discovery
	crdDiscovery := kubernetes.NewCRDDiscovery(s.clientSet, 5*time.Minute)
	err := crdDiscovery.DiscoverCRDs(s.ctx)
	s.Require().NoError(err, "Failed to discover CRDs")

	// Create toolset
	s.toolset = policy.NewToolset(s.provider, crdDiscovery)
	s.Require().True(s.toolset.IsEnabled(), "Policy toolset should be enabled")

	// Set observability
	logger := observability.NewLogger(observability.LogLevelInfo, false)
	// Use a test-specific registry to avoid duplicate registration in parallel tests
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry)
	s.toolset.SetObservability(logger, metrics)
}

// TestViolationsList tests policy.violations_list operation.
func (s *PolicyTestSuite) TestViolationsList() {
	ctx := context.Background()
	namespace := "policy-system"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a PolicyReport
	policyReportGVR := schema.GroupVersionResource{
		Group:    "wgpolicyk8s.io",
		Version:  "v1alpha2",
		Resource: "policyreports",
	}

	policyReport := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "wgpolicyk8s.io/v1alpha2",
			"kind":       "PolicyReport",
			"metadata": map[string]interface{}{
				"name":      "test-policyreport",
				"namespace": namespace,
			},
			"results": []interface{}{
				map[string]interface{}{
					"policy":    "require-labels",
					"rule":      "check-labels",
					"message":   "validation error: required label 'app' is missing",
					"severity":  "error",
					"timestamp": time.Now().Format(time.RFC3339),
					"resources": []interface{}{
						map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"name":       "test-pod",
							"namespace":  namespace,
						},
					},
				},
				map[string]interface{}{
					"policy":    "require-labels",
					"rule":      "check-annotations",
					"message":   "validation passed",
					"severity":  "info",
					"timestamp": time.Now().Format(time.RFC3339),
					"resources": []interface{}{
						map[string]interface{}{
							"kind":       "Pod",
							"apiVersion": "v1",
							"name":       "test-pod-2",
							"namespace":  namespace,
						},
					},
				},
			},
		},
	}

	_, err = createUnstructured(ctx, s.clientSet, policyReportGVR, namespace, policyReport)
	s.Require().NoError(err, "Failed to create PolicyReport")

	// Test violations_list
	args := struct {
		Context   string `json:"context"`
		Namespace string `json:"namespace"`
		Engine    string `json:"engine"`
		Limit     int    `json:"limit"`
		Continue  string `json:"continue"`
	}{
		Context:   "",
		Namespace: namespace,
		Engine:    "kyverno",
		Limit:     0, // No limit for this test
		Continue:  "",
	}

	result, err := s.toolset.TestHandleViolationsList(ctx, args)
	s.Require().NoError(err, "violations_list should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var resultData map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "should parse JSON result")

	items, ok := resultData["items"].([]any)
	s.Require().True(ok, "items should be a slice")
	s.Require().GreaterOrEqual(len(items), 1, "should have at least one violation")

	// Verify violation structure
	if len(items) > 0 {
		violation, ok := items[0].(map[string]any)
		s.Require().True(ok, "violation should be a map")
		s.Equal("kyverno", violation["engine"], "engine should be kyverno")
		s.Equal("require-labels", violation["policy"], "policy should match")
	}
}

// TestPolicyTestSuite runs the Policy test suite.
func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyTestSuite))
}
