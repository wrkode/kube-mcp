package kubevirt

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
)

// KubeVirtToolsetTestSuite tests KubeVirt toolset functionality.
type KubeVirtToolsetTestSuite struct {
	suite.Suite
	provider *mockProvider
}

// SetupTest sets up the test.
func (s *KubeVirtToolsetTestSuite) SetupTest() {
	s.provider = &mockProvider{}
}

// TestCRDDetection tests CRD detection logic.
func (s *KubeVirtToolsetTestSuite) TestCRDDetection() {
	// Test with nil discovery (should be disabled)
	toolset := NewToolset(s.provider, nil)
	s.False(toolset.IsEnabled(), "Toolset should be disabled when discovery is nil")

	// Note: Full CRD detection testing requires a real CRDDiscovery instance
	// which needs a real ClientSet. This is better tested in integration tests.
	// Unit tests here verify the basic structure and behavior.
}

// TestCRDDetectionWithOptionalCRDs tests detection with optional CRDs.
// Note: This requires a real CRDDiscovery instance, better tested in integration tests.
func (s *KubeVirtToolsetTestSuite) TestCRDDetectionWithOptionalCRDs() {
	// Test basic structure - full CRD detection is tested in integration tests
	toolset := NewToolset(s.provider, nil)
	s.False(toolset.IsEnabled(), "Toolset should be disabled without discovery")
}

// TestToolsetName tests toolset name.
func (s *KubeVirtToolsetTestSuite) TestToolsetName() {
	toolset := NewToolset(s.provider, nil)
	s.Equal("kubevirt", toolset.Name(), "Toolset name should be kubevirt")
}

// TestToolsWhenDisabled tests that tools return empty when disabled.
func (s *KubeVirtToolsetTestSuite) TestToolsWhenDisabled() {
	toolset := NewToolset(s.provider, nil)
	s.False(toolset.IsEnabled(), "Toolset should be disabled")

	tools := toolset.Tools()
	s.Empty(tools, "Tools should be empty when disabled")
}

// Mock implementations
type mockProvider struct{}

func (m *mockProvider) GetClientSet(context string) (*kubernetes.ClientSet, error) {
	return nil, nil
}

func (m *mockProvider) ListContexts() ([]string, error) {
	return []string{}, nil
}

func (m *mockProvider) GetCurrentContext() (string, error) {
	return "", nil
}

// TestKubeVirtToolsetTestSuite runs the KubeVirt toolset test suite.
func TestKubeVirtToolsetTestSuite(t *testing.T) {
	suite.Run(t, new(KubeVirtToolsetTestSuite))
}
