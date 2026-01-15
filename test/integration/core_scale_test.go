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

// CoreScaleTestSuite tests scale subresource operations.
type CoreScaleTestSuite struct {
	EnvtestSuite
	toolset *core.Toolset
}

// SetupTest sets up the test.
func (s *CoreScaleTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()
	s.toolset = core.NewToolset(s.provider)
}

// TestResourcesScale tests resources_scale operation.
func (s *CoreScaleTestSuite) TestResourcesScale() {
	ctx := context.Background()
	namespace := "test-ns-scale"

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
			Name:      "test-deployment-scale",
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

	// Test resources_scale - get current replicas (replicas = nil means get only)
	args := struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Replicas  *int   `json:"replicas"` // nil = get-only, 0 = scale to zero, >0 = scale to that number
		DryRun    bool   `json:"dry_run"`
		Context   string `json:"context"`
	}{
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
		Name:      "test-deployment-scale",
		Namespace: namespace,
		Replicas:  nil, // nil means get current scale
		DryRun:    false,
		Context:   "",
	}

	result, err := s.toolset.TestHandleResourcesScale(ctx, args)
	s.Require().NoError(err, "resources_scale get should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Parse result
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var scaleData map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &scaleData)
	s.Require().NoError(err, "should parse JSON result")

	replicas, ok := scaleData["replicas"]
	s.Require().True(ok, "replicas should be present")
	replicasFloat, ok := replicas.(float64)
	s.Require().True(ok, "replicas should be a number")
	s.Equal(float64(2), replicasFloat, "current replicas should be 2")

	// Test resources_scale - update replicas
	replicasValue := 5
	args.Replicas = &replicasValue
	result, err = s.toolset.TestHandleResourcesScale(ctx, args)
	s.Require().NoError(err, "resources_scale update should succeed")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify deployment was scaled
	deploy, err := s.clientSet.Typed.AppsV1().Deployments(namespace).Get(ctx, "test-deployment-scale", metav1.GetOptions{})
	s.Require().NoError(err, "deployment should exist")
	s.Equal(int32(5), *deploy.Spec.Replicas, "replicas should be updated to 5")
}

// TestResourcesScaleNonScalable tests that scaling a non-scalable resource returns an error.
func (s *CoreScaleTestSuite) TestResourcesScaleNonScalable() {
	ctx := context.Background()
	namespace := "test-ns-scale-error"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a ConfigMap (non-scalable resource)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: namespace,
		},
		Data: map[string]string{
			"key": "value",
		},
	}
	_, err = s.clientSet.Typed.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create configmap")

	// Test resources_scale on non-scalable resource
	replicas := 3
	args := struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Replicas  *int   `json:"replicas"` // nil = get-only, 0 = scale to zero, >0 = scale to that number
		DryRun    bool   `json:"dry_run"`
		Context   string `json:"context"`
	}{
		Group:     "",
		Version:   "v1",
		Kind:      "ConfigMap",
		Name:      "test-configmap",
		Namespace: namespace,
		Replicas:  &replicas,
		DryRun:    false,
		Context:   "",
	}

	result, err := s.toolset.TestHandleResourcesScale(ctx, args)
	// Should return an error result
	s.Require().NotNil(result, "result should not be nil")
	s.True(result.IsError, "result should be an error for non-scalable resource")
}

// TestResourcesScaleDryRun tests resources_scale with dry-run.
func (s *CoreScaleTestSuite) TestResourcesScaleDryRun() {
	ctx := context.Background()
	namespace := "test-ns-scale-dryrun"

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
			Name:      "test-deployment-scale-dryrun",
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

	// Test resources_scale with dry-run
	replicasValue := 5
	args := struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Replicas  *int   `json:"replicas"`
		DryRun    bool   `json:"dry_run"`
		Context   string `json:"context"`
	}{
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
		Name:      "test-deployment-scale-dryrun",
		Namespace: namespace,
		Replicas:  &replicasValue,
		DryRun:    true,
		Context:   "",
	}

	result, err := s.toolset.TestHandleResourcesScale(ctx, args)
	s.Require().NoError(err, "resources_scale dry-run should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify deployment was NOT scaled (dry-run)
	deploy, err := s.clientSet.Typed.AppsV1().Deployments(namespace).Get(ctx, "test-deployment-scale-dryrun", metav1.GetOptions{})
	s.Require().NoError(err, "deployment should exist")
	s.Equal(int32(2), *deploy.Spec.Replicas, "replicas should NOT be updated after dry-run")
}

// TestCoreScaleSuite runs the core scale test suite.
func TestCoreScaleSuite(t *testing.T) {
	suite.Run(t, new(CoreScaleTestSuite))
}
