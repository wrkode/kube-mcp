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
	"github.com/wrkode/kube-mcp/pkg/toolsets/gitops"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GitOpsTestSuite tests GitOps toolset functionality.
type GitOpsTestSuite struct {
	EnvtestSuite
	toolset *gitops.Toolset
}

// SetupTest sets up the test.
func (s *GitOpsTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()

	// Create CRD discovery
	crdDiscovery := kubernetes.NewCRDDiscovery(s.clientSet, 5*time.Minute)
	err := crdDiscovery.DiscoverCRDs(s.ctx)
	s.Require().NoError(err, "Failed to discover CRDs")

	// Create toolset
	s.toolset = gitops.NewToolset(s.provider, crdDiscovery)
	s.Require().True(s.toolset.IsEnabled(), "GitOps toolset should be enabled")

	// Set observability
	logger := observability.NewLogger(observability.LogLevelInfo, false)
	// Use a test-specific registry to avoid duplicate registration in parallel tests
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry)
	s.toolset.SetObservability(logger, metrics)

	// Disable RBAC for tests
	s.toolset.SetRBACAuthorizer(nil, false)
}

// TestAppsList tests gitops.apps_list operation.
func (s *GitOpsTestSuite) TestAppsList() {
	ctx := context.Background()
	namespace := "flux-system"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Kustomization
	kustomizationGVR := schema.GroupVersionResource{
		Group:    "kustomize.toolkit.fluxcd.io",
		Version:  "v1",
		Resource: "kustomizations",
	}

	kustomization := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kustomize.toolkit.fluxcd.io/v1",
			"kind":       "Kustomization",
			"metadata": map[string]interface{}{
				"name":      "test-kustomization",
				"namespace": namespace,
			},
			"status": map[string]interface{}{
				"lastAppliedRevision": "main/abc123",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":               "Ready",
						"status":             "True",
						"message":            "Applied revision: main/abc123",
						"lastTransitionTime": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	_, err = createUnstructured(ctx, s.clientSet, kustomizationGVR, namespace, kustomization)
	s.Require().NoError(err, "Failed to create Kustomization")

	// Create a HelmRelease
	helmReleaseGVR := schema.GroupVersionResource{
		Group:    "helm.toolkit.fluxcd.io",
		Version:  "v2",
		Resource: "helmreleases",
	}

	helmRelease := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "helm.toolkit.fluxcd.io/v2",
			"kind":       "HelmRelease",
			"metadata": map[string]interface{}{
				"name":      "test-helmrelease",
				"namespace": namespace,
			},
			"status": map[string]interface{}{
				"lastAppliedRevision": "1.0.0",
				"artifact": map[string]interface{}{
					"revision": "1.0.0",
				},
				"conditions": []interface{}{
					map[string]interface{}{
						"type":               "Ready",
						"status":             "True",
						"message":            "Release reconciliation succeeded",
						"lastTransitionTime": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	_, err = createUnstructured(ctx, s.clientSet, helmReleaseGVR, namespace, helmRelease)
	s.Require().NoError(err, "Failed to create HelmRelease")

	// Test apps_list
	args := struct {
		Context       string   `json:"context"`
		Namespace     string   `json:"namespace"`
		LabelSelector string   `json:"label_selector"`
		Kinds         []string `json:"kinds"`
		Limit         int      `json:"limit"`
		Continue      string   `json:"continue"`
	}{
		Context:   "",
		Namespace: namespace,
		Kinds:     []string{},
		Limit:     0,
		Continue:  "",
	}

	result, err := s.toolset.TestHandleAppsList(ctx, args)
	s.Require().NoError(err, "apps_list should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result contains both apps
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var resultData map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "should parse JSON result")

	items, ok := resultData["items"].([]any)
	s.Require().True(ok, "items should be a slice")
	s.Require().GreaterOrEqual(len(items), 2, "should have at least two apps")

	// Verify Kustomization
	foundKustomization := false
	foundHelmRelease := false
	for _, item := range items {
		itemData, ok := item.(map[string]any)
		s.Require().True(ok, "item should be a map")
		if itemData["kind"] == "Kustomization" && itemData["name"] == "test-kustomization" {
			foundKustomization = true
			s.Equal("Ready", itemData["status"], "Kustomization status should be Ready")
			s.Equal(true, itemData["ready"], "Kustomization should be ready")
		}
		if itemData["kind"] == "HelmRelease" && itemData["name"] == "test-helmrelease" {
			foundHelmRelease = true
			s.Equal("Ready", itemData["status"], "HelmRelease status should be Ready")
			s.Equal(true, itemData["ready"], "HelmRelease should be ready")
		}
	}
	s.True(foundKustomization, "should find Kustomization")
	s.True(foundHelmRelease, "should find HelmRelease")
}

// TestAppGet tests gitops.app_get operation.
func (s *GitOpsTestSuite) TestAppGet() {
	ctx := context.Background()
	namespace := "flux-system-get"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Kustomization
	kustomizationGVR := schema.GroupVersionResource{
		Group:    "kustomize.toolkit.fluxcd.io",
		Version:  "v1",
		Resource: "kustomizations",
	}

	kustomization := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kustomize.toolkit.fluxcd.io/v1",
			"kind":       "Kustomization",
			"metadata": map[string]interface{}{
				"name":      "test-kustomization-get",
				"namespace": namespace,
			},
			"status": map[string]interface{}{
				"lastAppliedRevision": "main/def456",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":               "Ready",
						"status":             "True",
						"message":            "Applied revision: main/def456",
						"lastTransitionTime": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	_, err = createUnstructured(ctx, s.clientSet, kustomizationGVR, namespace, kustomization)
	s.Require().NoError(err, "Failed to create Kustomization")

	// Test app_get
	args := struct {
		Context   string `json:"context"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Raw       bool   `json:"raw"`
	}{
		Context:   "",
		Kind:      "Kustomization",
		Name:      "test-kustomization-get",
		Namespace: namespace,
		Raw:       false,
	}

	result, err := s.toolset.TestHandleAppGet(ctx, args)
	s.Require().NoError(err, "app_get should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var resultData map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "should parse JSON result")

	summary, ok := resultData["summary"].(map[string]any)
	s.Require().True(ok, "summary should be a map")
	s.Equal("Kustomization", summary["kind"], "kind should match")
	s.Equal("test-kustomization-get", summary["name"], "name should match")
	s.Equal(true, summary["ready"], "should be ready")
	s.Equal("Ready", summary["status"], "status should be Ready")
}

// TestAppReconcile tests gitops.app_reconcile operation.
func (s *GitOpsTestSuite) TestAppReconcile() {
	ctx := context.Background()
	namespace := "flux-system-reconcile"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Kustomization
	kustomizationGVR := schema.GroupVersionResource{
		Group:    "kustomize.toolkit.fluxcd.io",
		Version:  "v1",
		Resource: "kustomizations",
	}

	kustomization := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kustomize.toolkit.fluxcd.io/v1",
			"kind":       "Kustomization",
			"metadata": map[string]interface{}{
				"name":      "test-kustomization-reconcile",
				"namespace": namespace,
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
				},
			},
		},
	}

	_, err = createUnstructured(ctx, s.clientSet, kustomizationGVR, namespace, kustomization)
	s.Require().NoError(err, "Failed to create Kustomization")

	// Test app_reconcile
	args := struct {
		Context   string `json:"context"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Confirm   bool   `json:"confirm"`
	}{
		Context:   "",
		Kind:      "Kustomization",
		Name:      "test-kustomization-reconcile",
		Namespace: namespace,
		Confirm:   true,
	}

	result, err := s.toolset.TestHandleAppReconcile(ctx, args)
	s.Require().NoError(err, "app_reconcile should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var resultData map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "should parse JSON result")

	reconcileResult, ok := resultData["result"].(map[string]any)
	s.Require().True(ok, "result should be a map")
	annotationApplied, ok := reconcileResult["annotation_applied"].(string)
	s.Require().True(ok, "annotation_applied should be present")
	s.Equal("kustomize.toolkit.fluxcd.io/reconcile", annotationApplied, "annotation should match")

	// Verify the annotation was actually added
	updated, err := getUnstructured(ctx, s.clientSet, kustomizationGVR, namespace, "test-kustomization-reconcile")
	s.Require().NoError(err, "Failed to get updated Kustomization")
	annotations := updated.GetAnnotations()
	s.Require().NotNil(annotations, "annotations should not be nil")
	_, found := annotations["kustomize.toolkit.fluxcd.io/reconcile"]
	s.True(found, "reconcile annotation should be present")
}

// TestGitOpsTestSuite runs the GitOps test suite.
func TestGitOpsTestSuite(t *testing.T) {
	suite.Run(t, new(GitOpsTestSuite))
}
