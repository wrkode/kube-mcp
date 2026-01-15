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
	"github.com/wrkode/kube-mcp/pkg/toolsets/autoscaling"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutoscalingTestSuite tests Autoscaling toolset functionality.
type AutoscalingTestSuite struct {
	EnvtestSuite
	toolset *autoscaling.Toolset
}

// SetupTest sets up the test.
func (s *AutoscalingTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()

	crdDiscovery := kubernetes.NewCRDDiscovery(s.clientSet, 5*time.Minute)
	err := crdDiscovery.DiscoverCRDs(s.ctx)
	s.Require().NoError(err, "Failed to discover CRDs")

	s.toolset = autoscaling.NewToolset(s.provider, crdDiscovery)
	s.Require().True(s.toolset.IsEnabled(), "Autoscaling toolset should be enabled")

	logger := observability.NewLogger(observability.LogLevelInfo, false)
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry)
	s.toolset.SetObservability(logger, metrics)
	s.toolset.SetRBACAuthorizer(nil, false)
}

// TestHPAList tests autoscaling.hpa_list operation.
func (s *AutoscalingTestSuite) TestHPAList() {
	ctx := context.Background()
	namespace := "default"

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-hpa",
			Namespace: namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind: "Deployment",
				Name: "test-deployment",
			},
			MinReplicas: int32Ptr(1),
			MaxReplicas: 10,
		},
		Status: autoscalingv2.HorizontalPodAutoscalerStatus{
			CurrentReplicas: 2,
			DesiredReplicas: 2,
		},
	}
	_, err := s.clientSet.Typed.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(ctx, hpa, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create HPA")

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

	result, err := s.toolset.TestHandleHPAList(ctx, args)
	s.Require().NoError(err, "HPA list should succeed")
	s.Require().NotNil(result, "Result should not be nil")
	s.Require().False(result.IsError, "Result should not be an error")

	var resultData map[string]interface{}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "Result should be TextContent")
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "Should parse result JSON")

	items, ok := resultData["items"].([]interface{})
	s.Require().True(ok, "Result should have items array")
	s.Require().Greater(len(items), 0, "Should have at least one HPA")
}

// TestAutoscalingTestSuite runs the autoscaling test suite.
func TestAutoscalingTestSuite(t *testing.T) {
	suite.Run(t, new(AutoscalingTestSuite))
}
