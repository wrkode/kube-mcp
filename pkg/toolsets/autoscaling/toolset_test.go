package autoscaling

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
)

// AutoscalingToolsetTestSuite tests the Autoscaling toolset.
type AutoscalingToolsetTestSuite struct {
	suite.Suite
	provider *mockProvider
}

// SetupTest sets up the test.
func (s *AutoscalingToolsetTestSuite) SetupTest() {
	s.provider = &mockProvider{}
}

// TestCRDDetection tests CRD detection logic.
func (s *AutoscalingToolsetTestSuite) TestCRDDetection() {
	// Toolset should always be enabled (HPA is native)
	toolset := NewToolset(s.provider, nil)
	s.True(toolset.IsEnabled(), "Toolset should always be enabled (HPA is native)")
	s.Equal("autoscaling", toolset.Name(), "Toolset name should be autoscaling")
}

// TestToolsetName tests toolset name.
func (s *AutoscalingToolsetTestSuite) TestToolsetName() {
	toolset := NewToolset(s.provider, nil)
	s.Equal("autoscaling", toolset.Name(), "Toolset name should be autoscaling")
}

// TestToolsWhenDisabled tests that HPA tools are still returned (always available).
func (s *AutoscalingToolsetTestSuite) TestToolsWhenDisabled() {
	toolset := NewToolset(s.provider, nil)
	tools := toolset.Tools()
	s.NotEmpty(tools, "HPA tools should always be available")
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

// TestAutoscalingToolsetTestSuite runs the autoscaling toolset test suite.
func TestAutoscalingToolsetTestSuite(t *testing.T) {
	suite.Run(t, new(AutoscalingToolsetTestSuite))
}
