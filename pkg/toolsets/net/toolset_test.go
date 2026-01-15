package net

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/config"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
)

// NetToolsetTestSuite tests the Network toolset.
type NetToolsetTestSuite struct {
	suite.Suite
	provider *mockProvider
}

// SetupTest sets up the test.
func (s *NetToolsetTestSuite) SetupTest() {
	s.provider = &mockProvider{}
}

// TestCRDDetection tests CRD detection logic.
func (s *NetToolsetTestSuite) TestCRDDetection() {
	// Toolset should always be enabled (NetworkPolicy is native)
	netConfig := config.NetConfig{}
	toolset := NewToolset(s.provider, nil, netConfig)
	s.True(toolset.IsEnabled(), "Toolset should always be enabled (NetworkPolicy is native)")
	s.Equal("net", toolset.Name(), "Toolset name should be net")
}

// TestToolsetName tests toolset name.
func (s *NetToolsetTestSuite) TestToolsetName() {
	toolset := &Toolset{}
	s.Equal("net", toolset.Name(), "Toolset name should be net")
}

// TestToolsWhenDisabled tests that NetworkPolicy tools are still returned (always available).
func (s *NetToolsetTestSuite) TestToolsWhenDisabled() {
	netConfig := config.NetConfig{}
	toolset := NewToolset(s.provider, nil, netConfig)
	tools := toolset.Tools()
	s.NotEmpty(tools, "NetworkPolicy tools should always be available")
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

// TestNetToolsetTestSuite runs the network toolset test suite.
func TestNetToolsetTestSuite(t *testing.T) {
	suite.Run(t, new(NetToolsetTestSuite))
}
