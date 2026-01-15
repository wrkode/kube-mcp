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

// CoreValidateTestSuite tests validation operations.
type CoreValidateTestSuite struct {
	EnvtestSuite
	toolset *core.Toolset
}

// SetupTest sets up the test.
func (s *CoreValidateTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()
	s.toolset = core.NewToolset(s.provider)
}

// TestResourcesValidateValid tests validation of a valid manifest.
func (s *CoreValidateTestSuite) TestResourcesValidateValid() {
	ctx := context.Background()
	namespace := "test-ns-validate"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Create a valid deployment manifest
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment-validate",
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

	deploymentMap, err := toMap(deployment)
	s.Require().NoError(err, "Failed to convert deployment to map")

	// Test validation
	args := struct {
		Manifest     map[string]interface{} `json:"manifest"`
		SchemaVersion string                `json:"schema_version"`
		Context      string                 `json:"context"`
	}{
		Manifest:     deploymentMap,
		SchemaVersion: "",
		Context:      "",
	}

	result, err := s.toolset.TestHandleResourcesValidate(ctx, args)
	s.Require().NoError(err, "validation should succeed")
	s.Require().NotNil(result, "result should not be nil")
	s.Require().False(result.IsError, "result should not be an error")

	// Verify validation result
	s.Require().Len(result.Content, 1, "result should have one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	s.Require().True(ok, "result content should be TextContent")

	var validateData map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &validateData)
	s.Require().NoError(err, "should parse JSON result")
	s.Equal(true, validateData["valid"], "validation should be valid")
}

// TestResourcesValidateInvalid tests validation of an invalid manifest.
func (s *CoreValidateTestSuite) TestResourcesValidateInvalid() {
	ctx := context.Background()

	// Create an invalid manifest (missing required fields)
	invalidManifest := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		// Missing metadata.name
		"spec": map[string]interface{}{
			"replicas": 2,
		},
	}

	args := struct {
		Manifest     map[string]interface{} `json:"manifest"`
		SchemaVersion string                `json:"schema_version"`
		Context      string                 `json:"context"`
	}{
		Manifest:     invalidManifest,
		SchemaVersion: "",
		Context:      "",
	}

	result, _ := s.toolset.TestHandleResourcesValidate(ctx, args)
	s.Require().NotNil(result, "result should not be nil")
	s.True(result.IsError, "result should be an error for invalid manifest")
}

// TestResourcesValidateMissingNamespace tests validation when namespace doesn't exist.
func (s *CoreValidateTestSuite) TestResourcesValidateMissingNamespace() {
	ctx := context.Background()

	// Create a valid deployment manifest with non-existent namespace
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "non-existent-namespace",
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

	deploymentMap, err := toMap(deployment)
	s.Require().NoError(err, "Failed to convert deployment to map")

	args := struct {
		Manifest     map[string]interface{} `json:"manifest"`
		SchemaVersion string                `json:"schema_version"`
		Context      string                 `json:"context"`
	}{
		Manifest:     deploymentMap,
		SchemaVersion: "",
		Context:      "",
	}

	result, err := s.toolset.TestHandleResourcesValidate(ctx, args)
	s.Require().NoError(err, "validation should not return error even if namespace doesn't exist")
	s.Require().NotNil(result, "result should not be nil")
	s.True(result.IsError, "result should be an error for non-existent namespace")
}

// TestCoreValidateSuite runs the core validate test suite.
func TestCoreValidateSuite(t *testing.T) {
	suite.Run(t, new(CoreValidateTestSuite))
}

