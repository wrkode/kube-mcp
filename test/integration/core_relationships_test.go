package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/toolsets/core"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CoreRelationshipsTestSuite tests resource relationships operations.
type CoreRelationshipsTestSuite struct {
	EnvtestSuite
	toolset *core.Toolset
}

// SetupTest sets up the test.
func (s *CoreRelationshipsTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()
	s.toolset = core.NewToolset(s.provider)
}

// TestResourcesRelationships tests finding resource relationships.
func (s *CoreRelationshipsTestSuite) TestResourcesRelationships() {
	ctx := context.Background()
	namespace := "test-ns-rel"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment-rel",
			Namespace: namespace,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
	deployment, err = s.clientSet.Typed.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create deployment")

	// Create a Pod that will be owned by the Deployment (via ReplicaSet)
	// Note: In real scenarios, pods are created by ReplicaSets, not directly
	// For this test, we'll test the relationships tool on the deployment itself

	// Test relationships for deployment (should find dependents - pods via ReplicaSet)
	args := struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Direction string `json:"direction"`
		Context   string `json:"context"`
	}{
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
		Name:      "test-deployment-rel",
		Namespace: namespace,
		Direction: "both",
		Context:   "",
	}

	result, err := s.toolset.TestHandleResourcesRelationships(ctx, args)
	s.Require().NoError(err, "resources_relationships should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result structure
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var relData map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &relData)
	s.Require().NoError(err, "should parse JSON result")
	s.Equal("test-deployment-rel", relData["name"])
	s.Equal(namespace, relData["namespace"])
	s.Equal("Deployment", relData["kind"])

	// Should have owners and dependents arrays (even if empty)
	_, hasOwners := relData["owners"]
	s.True(hasOwners, "result should have owners field")
	_, hasDependents := relData["dependents"]
	s.True(hasDependents, "result should have dependents field")
}

// TestResourcesRelationshipsOwnersOnly tests finding only owners.
func (s *CoreRelationshipsTestSuite) TestResourcesRelationshipsOwnersOnly() {
	ctx := context.Background()
	namespace := "test-ns-rel-owners"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Pod (pods typically have owners like ReplicaSets)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-rel",
			Namespace: namespace,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "nginx:latest",
				},
			},
		},
	}
	_, err = s.clientSet.Typed.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create pod")

	// Test relationships for pod (owners only)
	args := struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Direction string `json:"direction"`
		Context   string `json:"context"`
	}{
		Group:     "",
		Version:   "v1",
		Kind:      "Pod",
		Name:      "test-pod-rel",
		Namespace: namespace,
		Direction: "owners",
		Context:   "",
	}

	result, err := s.toolset.TestHandleResourcesRelationships(ctx, args)
	s.Require().NoError(err, "resources_relationships should succeed")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok)
	var relData map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &relData)
	s.Require().NoError(err)

	// Should have owners but not dependents
	_, hasOwners := relData["owners"]
	s.True(hasOwners, "result should have owners field")
	_, hasDependents := relData["dependents"]
	s.False(hasDependents, "result should NOT have dependents field when direction is 'owners'")
}

// TestCoreRelationshipsSuite runs the relationships test suite.
func TestCoreRelationshipsSuite(t *testing.T) {
	suite.Run(t, new(CoreRelationshipsTestSuite))
}

