package capi

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
)

// CAPIToolsetTestSuite tests CAPI toolset functionality.
type CAPIToolsetTestSuite struct {
	suite.Suite
	provider *mockProvider
}

// SetupTest sets up the test.
func (s *CAPIToolsetTestSuite) SetupTest() {
	s.provider = &mockProvider{}
}

// TestCRDDetection tests CRD detection logic.
func (s *CAPIToolsetTestSuite) TestCRDDetection() {
	// Test with nil discovery (should be disabled)
	toolset := NewToolset(s.provider, nil)
	s.False(toolset.IsEnabled(), "Toolset should be disabled when discovery is nil")
}

// TestToolsetName tests toolset name.
func (s *CAPIToolsetTestSuite) TestToolsetName() {
	toolset := NewToolset(s.provider, nil)
	s.Equal("capi", toolset.Name(), "Toolset name should be capi")
}

// TestToolsWhenDisabled tests that tools return empty when disabled.
func (s *CAPIToolsetTestSuite) TestToolsWhenDisabled() {
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

// TestCAPIToolsetTestSuite runs the CAPI toolset test suite.
func TestCAPIToolsetTestSuite(t *testing.T) {
	suite.Run(t, new(CAPIToolsetTestSuite))
}
