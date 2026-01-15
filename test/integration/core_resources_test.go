package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/toolsets/core"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CoreResourcesTestSuite tests resource-related operations.
type CoreResourcesTestSuite struct {
	EnvtestSuite
	toolset *core.Toolset
}

// SetupTest sets up the test.
func (s *CoreResourcesTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()
	s.toolset = core.NewToolset(s.provider)
}

// TestResourcesApply tests resources_apply operation with server-side apply.
func (s *CoreResourcesTestSuite) TestResourcesApply() {
	ctx := context.Background()
	namespace := "test-ns-apply"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Deployment manifest
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
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

	// Convert to map[string]any for the tool
	deploymentMap, err := toMap(deployment)
	s.Require().NoError(err, "Failed to convert deployment to map")

	// Test resources_apply - create
	args := struct {
		Manifest     map[string]any `json:"manifest"`
		FieldManager string         `json:"field_manager"`
		DryRun       bool           `json:"dry_run"`
		Context      string         `json:"context"`
	}{
		Manifest:     deploymentMap,
		FieldManager: "kube-mcp-test",
		DryRun:       false,
		Context:      "",
	}

	result, err := s.toolset.TestHandleResourcesApply(ctx, args)
	s.Require().NoError(err, "resources_apply should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify deployment was created
	deploy, err := s.clientSet.Typed.AppsV1().Deployments(namespace).Get(ctx, "test-deployment", metav1.GetOptions{})
	s.Require().NoError(err, "deployment should exist")
	s.Equal(int32(2), *deploy.Spec.Replicas, "replicas should match")

	// Test resources_apply - update (idempotent)
	deployment.Spec.Replicas = int32Ptr(3)
	deploymentMap, err = toMap(deployment)
	s.Require().NoError(err, "Failed to convert updated deployment to map")

	args.Manifest = deploymentMap
	result, err = s.toolset.TestHandleResourcesApply(ctx, args)
	s.Require().NoError(err, "resources_apply update should succeed")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify deployment was updated
	deploy, err = s.clientSet.Typed.AppsV1().Deployments(namespace).Get(ctx, "test-deployment", metav1.GetOptions{})
	s.Require().NoError(err, "deployment should still exist")
	s.Equal(int32(3), *deploy.Spec.Replicas, "replicas should be updated")
}

// TestResourcesApplyDryRun tests resources_apply with dry-run.
func (s *CoreResourcesTestSuite) TestResourcesApplyDryRun() {
	ctx := context.Background()
	namespace := "test-ns-apply-dryrun"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a Deployment manifest
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment-dryrun",
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

	deploymentMap, err := toMap(deployment)
	s.Require().NoError(err, "Failed to convert deployment to map")

	// Test resources_apply with dry-run
	args := struct {
		Manifest     map[string]any `json:"manifest"`
		FieldManager string         `json:"field_manager"`
		DryRun       bool           `json:"dry_run"`
		Context      string         `json:"context"`
	}{
		Manifest:     deploymentMap,
		FieldManager: "kube-mcp-test",
		DryRun:       true,
		Context:      "",
	}

	result, err := s.toolset.TestHandleResourcesApply(ctx, args)
	s.Require().NoError(err, "resources_apply dry-run should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify deployment was NOT created (dry-run)
	_, err = s.clientSet.Typed.AppsV1().Deployments(namespace).Get(ctx, "test-deployment-dryrun", metav1.GetOptions{})
	s.Require().Error(err, "deployment should NOT exist after dry-run")
}

// TestCoreResourcesSuite runs the core resources test suite.
func TestCoreResourcesSuite(t *testing.T) {
	suite.Run(t, new(CoreResourcesTestSuite))
}
