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
	"github.com/wrkode/kube-mcp/pkg/toolsets/certs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CertsTestSuite tests Cert-Manager toolset functionality.
type CertsTestSuite struct {
	EnvtestSuite
	toolset *certs.Toolset
}

// SetupTest sets up the test.
func (s *CertsTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()

	crdDiscovery := kubernetes.NewCRDDiscovery(s.clientSet, 5*time.Minute)
	err := crdDiscovery.DiscoverCRDs(s.ctx)
	s.Require().NoError(err, "Failed to discover CRDs")

	s.toolset = certs.NewToolset(s.provider, crdDiscovery)
	s.Require().True(s.toolset.IsEnabled(), "Certs toolset should be enabled")

	logger := observability.NewLogger(observability.LogLevelInfo, false)
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry)
	s.toolset.SetObservability(logger, metrics)
	s.toolset.SetRBACAuthorizer(nil, false)
}

// TestCertificatesList tests certs.certificates_list operation.
func (s *CertsTestSuite) TestCertificatesList() {
	ctx := context.Background()
	namespace := "cert-manager"

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	certGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}

	cert := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      "test-cert",
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"secretName": "test-secret",
				"dnsNames":   []string{"example.com"},
			},
			"status": map[string]interface{}{
				"conditions": []map[string]interface{}{
					{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}
	_, err = s.clientSet.Dynamic.Resource(certGVR).Namespace(namespace).Create(ctx, cert, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create Certificate")

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

	result, err := s.toolset.TestHandleCertificatesList(ctx, args)
	s.Require().NoError(err, "Certificates list should succeed")
	s.Require().NotNil(result, "Result should not be nil")
	s.Require().False(result.IsError, "Result should not be an error")

	var resultData map[string]interface{}
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "Result should be TextContent")
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "Should parse result JSON")

	items, ok := resultData["items"].([]interface{})
	s.Require().True(ok, "Result should have items array")
	s.Require().Greater(len(items), 0, "Should have at least one certificate")
}

// TestCertsTestSuite runs the certs test suite.
func TestCertsTestSuite(t *testing.T) {
	suite.Run(t, new(CertsTestSuite))
}
