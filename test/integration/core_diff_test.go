package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/toolsets/core"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CoreDiffTestSuite tests diff operations.
type CoreDiffTestSuite struct {
	EnvtestSuite
	toolset *core.Toolset
}

// SetupTest sets up the test.
func (s *CoreDiffTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()
	s.toolset = core.NewToolset(s.provider)
}

// TestResourcesDiff tests resources_diff operation with changes.
func (s *CoreDiffTestSuite) TestResourcesDiff() {
	ctx := context.Background()
	namespace := "test-ns-diff"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment-diff",
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
	_, err = s.clientSet.Typed.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create deployment")

	// Create desired manifest with changes (different replicas and label)
	desiredDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment-diff",
			Namespace: namespace,
			Labels:    map[string]string{"app": "test", "version": "v2"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(5),
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

	desiredMap, err := toMap(desiredDeployment)
	s.Require().NoError(err, "Failed to convert desired deployment to map")

	// Test resources_diff
	args := struct {
		Group      string                 `json:"group"`
		Version    string                 `json:"version"`
		Kind       string                 `json:"kind"`
		Name       string                 `json:"name"`
		Namespace  string                 `json:"namespace"`
		Manifest   map[string]interface{} `json:"manifest"`
		DiffFormat string                 `json:"diff_format"`
		Context    string                 `json:"context"`
	}{
		Group:      "apps",
		Version:    "v1",
		Kind:       "Deployment",
		Name:       "test-deployment-diff",
		Namespace:  namespace,
		Manifest:   desiredMap,
		DiffFormat: "unified",
		Context:    "",
	}

	result, err := s.toolset.TestHandleResourcesDiff(ctx, args)
	s.Require().NoError(err, "resources_diff should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify diff output
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")
	s.Require().NotEmpty(textContent.Text, "diff output should not be empty")

	// Verify diff contains expected information
	diffText := textContent.Text
	s.True(strings.Contains(diffText, "Current") || strings.Contains(diffText, "Desired"), "diff should contain state information")
}

// TestResourcesDiffNoChanges tests resources_diff when there are no changes.
func (s *CoreDiffTestSuite) TestResourcesDiffNoChanges() {
	ctx := context.Background()
	namespace := "test-ns-diff-nochange"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment-nochange",
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
	_, err = s.clientSet.Typed.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create deployment")

	// Create desired manifest identical to current (except metadata fields)
	desiredDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment-nochange",
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

	desiredMap, err := toMap(desiredDeployment)
	s.Require().NoError(err, "Failed to convert desired deployment to map")

	// Test resources_diff
	args := struct {
		Group      string                 `json:"group"`
		Version    string                 `json:"version"`
		Kind       string                 `json:"kind"`
		Name       string                 `json:"name"`
		Namespace  string                 `json:"namespace"`
		Manifest   map[string]interface{} `json:"manifest"`
		DiffFormat string                 `json:"diff_format"`
		Context    string                 `json:"context"`
	}{
		Group:      "apps",
		Version:    "v1",
		Kind:       "Deployment",
		Name:       "test-deployment-nochange",
		Namespace:  namespace,
		Manifest:   desiredMap,
		DiffFormat: "unified",
		Context:    "",
	}

	result, err := s.toolset.TestHandleResourcesDiff(ctx, args)
	s.Require().NoError(err, "resources_diff should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify diff output indicates no changes
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")
	diffText := textContent.Text
	s.True(
		strings.Contains(diffText, "No differences") || strings.Contains(diffText, "No meaningful differences"),
		"diff should indicate no changes",
	)
}

// TestResourcesDiffJSONFormat tests resources_diff with JSON format.
func (s *CoreDiffTestSuite) TestResourcesDiffJSONFormat() {
	ctx := context.Background()
	namespace := "test-ns-diff-json"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment-json",
			Namespace: namespace,
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
	_, err = s.clientSet.Typed.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create deployment")

	// Create desired manifest with changes
	desiredDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment-json",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(5),
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

	desiredMap, err := toMap(desiredDeployment)
	s.Require().NoError(err, "Failed to convert desired deployment to map")

	// Test resources_diff with JSON format
	args := struct {
		Group      string                 `json:"group"`
		Version    string                 `json:"version"`
		Kind       string                 `json:"kind"`
		Name       string                 `json:"name"`
		Namespace  string                 `json:"namespace"`
		Manifest   map[string]interface{} `json:"manifest"`
		DiffFormat string                 `json:"diff_format"`
		Context    string                 `json:"context"`
	}{
		Group:      "apps",
		Version:    "v1",
		Kind:       "Deployment",
		Name:       "test-deployment-json",
		Namespace:  namespace,
		Manifest:   desiredMap,
		DiffFormat: "json",
		Context:    "",
	}

	result, err := s.toolset.TestHandleResourcesDiff(ctx, args)
	s.Require().NoError(err, "resources_diff should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify JSON diff output
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")
	s.Require().NotEmpty(textContent.Text, "JSON diff output should not be empty")
}

// TestResourcesDiffNotFound tests resources_diff when resource doesn't exist.
func (s *CoreDiffTestSuite) TestResourcesDiffNotFound() {
	ctx := context.Background()
	namespace := "test-ns-diff-notfound"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create desired manifest for non-existent resource
	desiredDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "non-existent-deployment",
			Namespace: namespace,
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

	desiredMap, err := toMap(desiredDeployment)
	s.Require().NoError(err, "Failed to convert desired deployment to map")

	// Test resources_diff on non-existent resource
	args := struct {
		Group      string                 `json:"group"`
		Version    string                 `json:"version"`
		Kind       string                 `json:"kind"`
		Name       string                 `json:"name"`
		Namespace  string                 `json:"namespace"`
		Manifest   map[string]interface{} `json:"manifest"`
		DiffFormat string                 `json:"diff_format"`
		Context    string                 `json:"context"`
	}{
		Group:      "apps",
		Version:    "v1",
		Kind:       "Deployment",
		Name:       "non-existent-deployment",
		Namespace:  namespace,
		Manifest:   desiredMap,
		DiffFormat: "unified",
		Context:    "",
	}

	result, err := s.toolset.TestHandleResourcesDiff(ctx, args)
	s.Require().NotNil(result, "result should not be nil")
	s.True(result.IsError, "result should be an error for non-existent resource")
}

// TestCoreDiffSuite runs the core diff test suite.
func TestCoreDiffSuite(t *testing.T) {
	suite.Run(t, new(CoreDiffTestSuite))
}

