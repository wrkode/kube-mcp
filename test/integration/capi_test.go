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
	"github.com/wrkode/kube-mcp/pkg/toolsets/capi"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CAPITestSuite tests CAPI toolset functionality.
type CAPITestSuite struct {
	EnvtestSuite
	toolset *capi.Toolset
}

// SetupTest sets up the test.
func (s *CAPITestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()

	// Create CRD discovery
	crdDiscovery := kubernetes.NewCRDDiscovery(s.clientSet, 5*time.Minute)
	err := crdDiscovery.DiscoverCRDs(s.ctx)
	s.Require().NoError(err, "Failed to discover CRDs")

	// Create toolset
	s.toolset = capi.NewToolset(s.provider, crdDiscovery)
	s.Require().True(s.toolset.IsEnabled(), "CAPI toolset should be enabled")

	// Set observability
	logger := observability.NewLogger(observability.LogLevelInfo, false)
	// Use a test-specific registry to avoid duplicate registration in parallel tests
	registry := prometheus.NewRegistry()
	metrics := observability.NewMetrics(registry)
	s.toolset.SetObservability(logger, metrics)

	// Disable RBAC for tests
	s.toolset.SetRBACAuthorizer(nil, false)
}

// TestClustersList tests capi.clusters_list operation.
func (s *CAPITestSuite) TestClustersList() {
	ctx := context.Background()
	namespace := "capi-system"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Cluster
	clusterGVR := schema.GroupVersionResource{
		Group:    "cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "clusters",
	}

	cluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "Cluster",
			"metadata": map[string]interface{}{
				"name":      "test-cluster",
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"controlPlaneRef": map[string]interface{}{
					"apiVersion": "controlplane.cluster.x-k8s.io/v1beta1",
					"kind":       "KubeadmControlPlane",
					"name":       "test-cluster-control-plane",
					"namespace":  namespace,
				},
				"infrastructureRef": map[string]interface{}{
					"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta1",
					"kind":       "AWSCluster",
					"name":       "test-cluster",
					"namespace":  namespace,
				},
				"topology": map[string]interface{}{
					"version": "v1.28.0",
				},
			},
			"status": map[string]interface{}{
				"ready":               true,
				"controlPlaneReady":   true,
				"infrastructureReady": true,
				"conditions": []interface{}{
					map[string]interface{}{
						"type":               "Ready",
						"status":             "True",
						"reason":             "AllReplicasReady",
						"message":            "Cluster is ready",
						"lastTransitionTime": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	createdCluster, err := createUnstructured(ctx, s.clientSet, clusterGVR, namespace, cluster)
	s.Require().NoError(err, "Failed to create Cluster")

	// Test clusters_list
	args := struct {
		Context       string `json:"context"`
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
	}{
		Context:   "",
		Namespace: namespace,
	}

	result, err := s.toolset.TestHandleClustersList(ctx, args)
	s.Require().NoError(err, "clusters_list should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result contains the cluster
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	// Parse JSON from text content
	var resultData map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "should parse JSON result")

	items, ok := resultData["items"].([]any)
	s.Require().True(ok, "items should be a slice")
	s.Require().Len(items, 1, "should have one cluster")

	clusterData, ok := items[0].(map[string]any)
	s.Require().True(ok, "cluster should be a map")
	s.Equal("test-cluster", clusterData["name"], "cluster name should match")
	s.Equal(namespace, clusterData["namespace"], "cluster namespace should match")
	s.Equal(true, clusterData["ready"], "cluster should be ready")
	s.Equal(true, clusterData["control_plane_ready"], "cluster control plane should be ready")
	s.Equal(true, clusterData["infrastructure_ready"], "cluster infrastructure should be ready")
	s.Equal("v1.28.0", clusterData["kubernetes_version"], "cluster version should match")

	_ = createdCluster // Use createdCluster to avoid unused variable warning
}

// TestClusterGet tests capi.cluster_get operation.
func (s *CAPITestSuite) TestClusterGet() {
	ctx := context.Background()
	namespace := "capi-system-get"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Cluster
	clusterGVR := schema.GroupVersionResource{
		Group:    "cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "clusters",
	}

	cluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "Cluster",
			"metadata": map[string]interface{}{
				"name":      "test-cluster-get",
				"namespace": namespace,
			},
			"status": map[string]interface{}{
				"ready": true,
				"conditions": []interface{}{
					map[string]interface{}{
						"type":               "Ready",
						"status":             "True",
						"reason":             "AllReplicasReady",
						"message":            "Cluster is ready",
						"lastTransitionTime": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	_, err = createUnstructured(ctx, s.clientSet, clusterGVR, namespace, cluster)
	s.Require().NoError(err, "Failed to create Cluster")

	// Test cluster_get
	args := struct {
		Context   string `json:"context"`
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Raw       bool   `json:"raw"`
	}{
		Context:   "",
		Namespace: namespace,
		Name:      "test-cluster-get",
		Raw:       false,
	}

	result, err := s.toolset.TestHandleClusterGet(ctx, args)
	s.Require().NoError(err, "cluster_get should succeed")
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
	s.Equal("test-cluster-get", summary["name"], "cluster name should match")
	s.Equal(true, summary["ready"], "cluster should be ready")
}

// TestMachineDeploymentsList tests capi.machinedeployments_list operation.
func (s *CAPITestSuite) TestMachineDeploymentsList() {
	ctx := context.Background()
	namespace := "capi-system-md"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Cluster first
	clusterGVR := schema.GroupVersionResource{
		Group:    "cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "clusters",
	}

	cluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "Cluster",
			"metadata": map[string]interface{}{
				"name":      "test-cluster-md",
				"namespace": namespace,
			},
		},
	}

	_, err = createUnstructured(ctx, s.clientSet, clusterGVR, namespace, cluster)
	s.Require().NoError(err, "Failed to create Cluster")

	// Create a MachineDeployment
	mdGVR := schema.GroupVersionResource{
		Group:    "cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "machinedeployments",
	}

	md := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "MachineDeployment",
			"metadata": map[string]interface{}{
				"name":      "test-md",
				"namespace": namespace,
				"labels": map[string]interface{}{
					"cluster.x-k8s.io/cluster-name": "test-cluster-md",
				},
			},
			"spec": map[string]interface{}{
				"replicas": int64(3),
			},
			"status": map[string]interface{}{
				"readyReplicas":   int64(3),
				"updatedReplicas": int64(3),
				"paused":          false,
				"conditions": []interface{}{
					map[string]interface{}{
						"type":               "Ready",
						"status":             "True",
						"reason":             "AllReplicasReady",
						"message":            "MachineDeployment is ready",
						"lastTransitionTime": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	_, err = createUnstructured(ctx, s.clientSet, mdGVR, namespace, md)
	s.Require().NoError(err, "Failed to create MachineDeployment")

	// Test machinedeployments_list
	args := struct {
		Context          string `json:"context"`
		ClusterNamespace string `json:"cluster_namespace"`
		ClusterName      string `json:"cluster_name"`
	}{
		Context:          "",
		ClusterNamespace: namespace,
		ClusterName:      "test-cluster-md",
	}

	result, err := s.toolset.TestHandleMachineDeploymentsList(ctx, args)
	s.Require().NoError(err, "machinedeployments_list should succeed")
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
	s.Require().Len(items, 1, "should have one machine deployment")

	mdData, ok := items[0].(map[string]any)
	s.Require().True(ok, "machine deployment should be a map")
	s.Equal("test-md", mdData["name"], "machine deployment name should match")
	s.Equal(float64(3), mdData["replicas_desired"], "replicas_desired should match")
	s.Equal(float64(3), mdData["replicas_ready"], "replicas_ready should match")
}

// TestScaleMachineDeployment tests capi.scale_machinedeployment operation.
func (s *CAPITestSuite) TestScaleMachineDeployment() {
	ctx := context.Background()
	namespace := "capi-system-scale"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a MachineDeployment
	mdGVR := schema.GroupVersionResource{
		Group:    "cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "machinedeployments",
	}

	md := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "MachineDeployment",
			"metadata": map[string]interface{}{
				"name":      "test-md-scale",
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"replicas": int64(2),
			},
			"status": map[string]interface{}{
				"readyReplicas":   int64(2),
				"updatedReplicas": int64(2),
			},
		},
	}

	_, err = createUnstructured(ctx, s.clientSet, mdGVR, namespace, md)
	s.Require().NoError(err, "Failed to create MachineDeployment")

	// Test scale_machinedeployment
	args := struct {
		Context   string `json:"context"`
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Replicas  int    `json:"replicas"`
		Confirm   bool   `json:"confirm"`
	}{
		Context:   "",
		Namespace: namespace,
		Name:      "test-md-scale",
		Replicas:  5,
		Confirm:   true,
	}

	result, err := s.toolset.TestHandleScaleMachineDeployment(ctx, args)
	s.Require().NoError(err, "scale_machinedeployment should succeed")
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
	s.Equal(float64(5), summary["replicas_desired"], "replicas_desired should be updated to 5")

	// Verify the actual resource was updated
	updated, err := getUnstructured(ctx, s.clientSet, mdGVR, namespace, "test-md-scale")
	s.Require().NoError(err, "Failed to get updated MachineDeployment")
	replicas, found, _ := unstructured.NestedInt64(updated.Object, "spec", "replicas")
	s.Require().True(found, "replicas should be found")
	s.Equal(int64(5), replicas, "replicas should be updated to 5")
}

// TestCAPITestSuite runs the CAPI test suite.
func TestCAPITestSuite(t *testing.T) {
	suite.Run(t, new(CAPITestSuite))
}
