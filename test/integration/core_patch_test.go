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

// CorePatchTestSuite tests patch operations.
type CorePatchTestSuite struct {
	EnvtestSuite
	toolset *core.Toolset
}

// SetupTest sets up the test.
func (s *CorePatchTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()
	s.toolset = core.NewToolset(s.provider)
}

// TestResourcesPatchMerge tests resources_patch with merge patch type.
func (s *CorePatchTestSuite) TestResourcesPatchMerge() {
	ctx := context.Background()
	namespace := "test-ns-patch"

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
			Name:      "test-deployment-patch",
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

	// Test resources_patch - update labels using merge patch
	args := struct {
		Group        string      `json:"group"`
		Version      string      `json:"version"`
		Kind         string      `json:"kind"`
		Name         string      `json:"name"`
		Namespace    string      `json:"namespace"`
		PatchType    string      `json:"patch_type"`
		PatchData    interface{} `json:"patch_data"`
		FieldManager string      `json:"field_manager"`
		DryRun       bool        `json:"dry_run"`
		Context      string      `json:"context"`
	}{
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
		Name:      "test-deployment-patch",
		Namespace: namespace,
		PatchType: "merge",
		PatchData: map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app":     "test",
					"version": "v2",
					"env":     "prod",
				},
			},
		},
		FieldManager: "kube-mcp-test",
		DryRun:       false,
		Context:      "",
	}

	result, err := s.toolset.TestHandleResourcesPatch(ctx, args)
	s.Require().NoError(err, "resources_patch should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify deployment was patched
	deploy, err := s.clientSet.Typed.AppsV1().Deployments(namespace).Get(ctx, "test-deployment-patch", metav1.GetOptions{})
	s.Require().NoError(err, "deployment should exist")
	s.Equal("v2", deploy.Labels["version"], "version label should be updated")
	s.Equal("prod", deploy.Labels["env"], "env label should be added")
	s.Equal("test", deploy.Labels["app"], "app label should remain")
}

// TestResourcesPatchStrategic tests resources_patch with strategic merge patch type.
func (s *CorePatchTestSuite) TestResourcesPatchStrategic() {
	ctx := context.Background()
	namespace := "test-ns-patch-strategic"

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
			Name:      "test-deployment-strategic",
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

	// Test resources_patch - update replicas using strategic merge patch
	args := struct {
		Group        string      `json:"group"`
		Version      string      `json:"version"`
		Kind         string      `json:"kind"`
		Name         string      `json:"name"`
		Namespace    string      `json:"namespace"`
		PatchType    string      `json:"patch_type"`
		PatchData    interface{} `json:"patch_data"`
		FieldManager string      `json:"field_manager"`
		DryRun       bool        `json:"dry_run"`
		Context      string      `json:"context"`
	}{
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
		Name:      "test-deployment-strategic",
		Namespace: namespace,
		PatchType: "strategic",
		PatchData: map[string]interface{}{
			"spec": map[string]interface{}{
				"replicas": 5,
			},
		},
		FieldManager: "kube-mcp-test",
		DryRun:       false,
		Context:      "",
	}

	result, err := s.toolset.TestHandleResourcesPatch(ctx, args)
	s.Require().NoError(err, "resources_patch should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify deployment was patched
	deploy, err := s.clientSet.Typed.AppsV1().Deployments(namespace).Get(ctx, "test-deployment-strategic", metav1.GetOptions{})
	s.Require().NoError(err, "deployment should exist")
	s.Equal(int32(5), *deploy.Spec.Replicas, "replicas should be updated to 5")
}

// TestResourcesPatchDryRun tests resources_patch with dry-run.
func (s *CorePatchTestSuite) TestResourcesPatchDryRun() {
	ctx := context.Background()
	namespace := "test-ns-patch-dryrun"

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
			Name:      "test-deployment-patch-dryrun",
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

	// Test resources_patch with dry-run
	args := struct {
		Group        string      `json:"group"`
		Version      string      `json:"version"`
		Kind         string      `json:"kind"`
		Name         string      `json:"name"`
		Namespace    string      `json:"namespace"`
		PatchType    string      `json:"patch_type"`
		PatchData    interface{} `json:"patch_data"`
		FieldManager string      `json:"field_manager"`
		DryRun       bool        `json:"dry_run"`
		Context      string      `json:"context"`
	}{
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
		Name:      "test-deployment-patch-dryrun",
		Namespace: namespace,
		PatchType: "merge",
		PatchData: map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"version": "v2",
				},
			},
		},
		FieldManager: "kube-mcp-test",
		DryRun:       true,
		Context:      "",
	}

	result, err := s.toolset.TestHandleResourcesPatch(ctx, args)
	s.Require().NoError(err, "resources_patch dry-run should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Parse result to verify dry-run status
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var patchData map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &patchData)
	s.Require().NoError(err, "should parse JSON result")
	s.Equal("dry-run-patched", patchData["status"], "status should be dry-run-patched")
	s.Equal(true, patchData["dry_run"], "dry_run should be true")

	// Verify deployment was NOT patched (dry-run)
	deploy, err := s.clientSet.Typed.AppsV1().Deployments(namespace).Get(ctx, "test-deployment-patch-dryrun", metav1.GetOptions{})
	s.Require().NoError(err, "deployment should exist")
	_, exists := deploy.Labels["version"]
	s.False(exists, "version label should NOT be added after dry-run")
}

// TestResourcesPatchInvalidType tests resources_patch with invalid patch type.
func (s *CorePatchTestSuite) TestResourcesPatchInvalidType() {
	ctx := context.Background()
	namespace := "test-ns-patch-invalid"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Test resources_patch with invalid patch type
	args := struct {
		Group        string      `json:"group"`
		Version      string      `json:"version"`
		Kind         string      `json:"kind"`
		Name         string      `json:"name"`
		Namespace    string      `json:"namespace"`
		PatchType    string      `json:"patch_type"`
		PatchData    interface{} `json:"patch_data"`
		FieldManager string      `json:"field_manager"`
		DryRun       bool        `json:"dry_run"`
		Context      string      `json:"context"`
	}{
		Group:     "apps",
		Version:   "v1",
		Kind:      "Deployment",
		Name:      "non-existent",
		Namespace: namespace,
		PatchType: "invalid",
		PatchData: map[string]interface{}{
			"spec": map[string]interface{}{
				"replicas": 5,
			},
		},
		DryRun:  false,
		Context: "",
	}

	result, err := s.toolset.TestHandleResourcesPatch(ctx, args)
	s.Require().NotNil(result, "result should not be nil")
	s.True(result.IsError, "result should be an error for invalid patch type")
}

// TestCorePatchSuite runs the core patch test suite.
func TestCorePatchSuite(t *testing.T) {
	suite.Run(t, new(CorePatchTestSuite))
}

