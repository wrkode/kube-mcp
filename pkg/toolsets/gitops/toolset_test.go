package gitops

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
)

// GitOpsToolsetTestSuite tests GitOps toolset functionality.
type GitOpsToolsetTestSuite struct {
	suite.Suite
	provider *mockProvider
}

// SetupTest sets up the test.
func (s *GitOpsToolsetTestSuite) SetupTest() {
	s.provider = &mockProvider{}
}

// TestCRDDetection tests CRD detection logic.
func (s *GitOpsToolsetTestSuite) TestCRDDetection() {
	// Test with nil discovery (should be disabled)
	toolset := NewToolset(s.provider, nil)
	s.False(toolset.IsEnabled(), "Toolset should be disabled when discovery is nil")
}

// TestToolsetName tests toolset name.
func (s *GitOpsToolsetTestSuite) TestToolsetName() {
	toolset := NewToolset(s.provider, nil)
	s.Equal("gitops", toolset.Name(), "Toolset name should be gitops")
}

// TestToolsWhenDisabled tests that tools return empty when disabled.
func (s *GitOpsToolsetTestSuite) TestToolsWhenDisabled() {
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

// TestGitOpsToolsetTestSuite runs the GitOps toolset test suite.
func TestGitOpsToolsetTestSuite(t *testing.T) {
	suite.Run(t, new(GitOpsToolsetTestSuite))
}
