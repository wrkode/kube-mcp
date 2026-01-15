package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/toolsets/core"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CorePodsTestSuite tests pod-related operations.
type CorePodsTestSuite struct {
	EnvtestSuite
	toolset *core.Toolset
}

// SetupTest sets up the test.
func (s *CorePodsTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()
	s.toolset = core.NewToolset(s.provider)
}

// TestPodsList tests pods_list operation.
func (s *CorePodsTestSuite) TestPodsList() {
	ctx := context.Background()
	namespace := "test-ns"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
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

	// Test pods_list
	args := struct {
		Namespace     string `json:"namespace"`
		LabelSelector string `json:"label_selector"`
		FieldSelector string `json:"field_selector"`
		Limit         int    `json:"limit"`
		Continue      string `json:"continue"`
		Context       string `json:"context"`
	}{
		Namespace: namespace,
		Context:   "",
	}

	result, err := s.toolset.TestHandlePodsList(ctx, args)
	s.Require().NoError(err, "pods_list should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result contains the pod
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	// Parse JSON from text content
	var resultData map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "should parse JSON result")

	pods, ok := resultData["pods"].([]any)
	s.Require().True(ok, "pods should be a slice")
	s.Require().Len(pods, 1, "should have one pod")
	podData, ok := pods[0].(map[string]any)
	s.Require().True(ok, "pod should be a map")
	s.Equal("test-pod", podData["name"], "pod name should match")
}

// TestPodsGet tests pods_get operation.
func (s *CorePodsTestSuite) TestPodsGet() {
	ctx := context.Background()
	namespace := "test-ns-get"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-get",
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

	// Test pods_get
	args := struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}{
		Name:      "test-pod-get",
		Namespace: namespace,
		Context:   "",
	}

	result, err := s.toolset.TestHandlePodsGet(ctx, args)
	s.Require().NoError(err, "pods_get should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify result
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var resultData map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resultData)
	s.Require().NoError(err, "should parse JSON result")
	s.Equal("test-pod-get", resultData["name"], "pod name should match")
	s.Equal(namespace, resultData["namespace"], "namespace should match")
}

// TestPodsDelete tests pods_delete operation.
func (s *CorePodsTestSuite) TestPodsDelete() {
	ctx := context.Background()
	namespace := "test-ns-delete"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-delete",
			Namespace: namespace,
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

	// Verify pod exists
	_, err = s.clientSet.Typed.CoreV1().Pods(namespace).Get(ctx, "test-pod-delete", metav1.GetOptions{})
	s.Require().NoError(err, "pod should exist")

	// Test pods_delete
	args := struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Context   string `json:"context"`
	}{
		Name:      "test-pod-delete",
		Namespace: namespace,
		Context:   "",
	}

	result, err := s.toolset.TestHandlePodsDelete(ctx, args)
	s.Require().NoError(err, "pods_delete should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify pod is deleted
	_, err = s.clientSet.Typed.CoreV1().Pods(namespace).Get(ctx, "test-pod-delete", metav1.GetOptions{})
	s.Require().Error(err, "pod should be deleted")
}

// TestCorePodsSuite runs the core pods test suite.
func TestCorePodsSuite(t *testing.T) {
	suite.Run(t, new(CorePodsTestSuite))
}
