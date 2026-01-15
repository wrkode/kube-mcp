package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/toolsets/helm"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelmTestSuite tests Helm operations.
// Note: Full Helm integration requires a real Kubernetes cluster or more complex envtest setup.
// This test provides basic structure and can be extended with actual Helm chart operations.
type HelmTestSuite struct {
	EnvtestSuite
	toolset *helm.Toolset
}

// SetupTest sets up the test.
func (s *HelmTestSuite) SetupTest() {
	s.EnvtestSuite.SetupTest()

	// Create a temporary directory for Helm settings
	tempDir, err := os.MkdirTemp("", "helm-test-*")
	s.Require().NoError(err, "Failed to create temp directory")

	// Create Helm settings
	settings := cli.New()
	settings.RepositoryCache = filepath.Join(tempDir, "repository")
	settings.RepositoryConfig = filepath.Join(tempDir, "repositories.yaml")

	// Create Helm toolset
	s.toolset = helm.NewToolset(s.provider, settings)
	s.toolset.SetObservability(nil, nil) // Can add logger/metrics if needed
}

// TestHelmReleasesList tests helm_releases_list operation.
// This is a basic test that verifies the tool can be called.
// Full Helm operations require Helm charts and releases to be installed.
func (s *HelmTestSuite) TestHelmReleasesList() {
	ctx := context.Background()
	namespace := "test-ns-helm"

	// Create namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.clientSet.Typed.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err, "Failed to create namespace")

	// Note: Actual Helm install/list/uninstall tests would require:
	// 1. Helm charts available
	// 2. Tiller/Helm 3 setup in envtest (complex)
	// 3. Or using a real cluster for integration tests
	//
	// For now, this test verifies the toolset can be created and tools registered.
	// Full Helm integration tests should be run against a real cluster or
	// with a more sophisticated test setup that includes Helm server components.

	// Verify toolset is created
	s.Require().NotNil(s.toolset, "Helm toolset should be created")
	s.Equal("helm", s.toolset.Name(), "Toolset name should be helm")

	// Verify tools are defined
	tools := s.toolset.Tools()
	s.Require().Greater(len(tools), 0, "Helm toolset should have tools")
}

// TestHelmToolsetRegistration tests that Helm toolset can be registered.
func (s *HelmTestSuite) TestHelmToolsetRegistration() {
	// This test verifies the toolset structure is correct
	// Actual Helm operations require Helm server components which are not
	// easily testable with envtest alone.

	s.Require().NotNil(s.toolset, "Toolset should be created")

	// Verify observability can be set
	s.toolset.SetObservability(nil, nil)

	// Verify tools method works
	tools := s.toolset.Tools()
	s.Require().NotNil(tools, "Tools should not be nil")
}

// TestHelmTestSuite runs the Helm test suite.
func TestHelmTestSuite(t *testing.T) {
	suite.Run(t, new(HelmTestSuite))
}
