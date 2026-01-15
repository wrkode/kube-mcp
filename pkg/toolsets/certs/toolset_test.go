package certs

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
)

// CertsToolsetTestSuite tests the Certs toolset.
type CertsToolsetTestSuite struct {
	suite.Suite
	provider *mockProvider
}

// SetupTest sets up the test.
func (s *CertsToolsetTestSuite) SetupTest() {
	s.provider = &mockProvider{}
}

// TestCRDDetection tests CRD detection logic.
func (s *CertsToolsetTestSuite) TestCRDDetection() {
	// Test with nil discovery (should be disabled)
	toolset := NewToolset(s.provider, nil)
	s.False(toolset.IsEnabled(), "Toolset should be disabled when discovery is nil")
}

// TestToolsetName tests toolset name.
func (s *CertsToolsetTestSuite) TestToolsetName() {
	toolset := NewToolset(s.provider, nil)
	s.Equal("certs", toolset.Name(), "Toolset name should be certs")
}

// TestToolsWhenDisabled tests that tools return empty when disabled.
func (s *CertsToolsetTestSuite) TestToolsWhenDisabled() {
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

// TestCertsToolsetTestSuite runs the certs toolset test suite.
func TestCertsToolsetTestSuite(t *testing.T) {
	suite.Run(t, new(CertsToolsetTestSuite))
}
