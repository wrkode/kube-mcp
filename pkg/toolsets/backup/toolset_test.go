package backup

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
)

// BackupToolsetTestSuite tests the Backup toolset.
type BackupToolsetTestSuite struct {
	suite.Suite
	provider *mockProvider
}

// SetupTest sets up the test.
func (s *BackupToolsetTestSuite) SetupTest() {
	s.provider = &mockProvider{}
}

// TestCRDDetection tests CRD detection logic.
func (s *BackupToolsetTestSuite) TestCRDDetection() {
	// Test with nil discovery (should be disabled)
	toolset := NewToolset(s.provider, nil)
	s.False(toolset.IsEnabled(), "Toolset should be disabled when discovery is nil")
}

// TestToolsetName tests toolset name.
func (s *BackupToolsetTestSuite) TestToolsetName() {
	toolset := NewToolset(s.provider, nil)
	s.Equal("backup", toolset.Name(), "Toolset name should be backup")
}

// TestToolsWhenDisabled tests that tools return empty when disabled.
func (s *BackupToolsetTestSuite) TestToolsWhenDisabled() {
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

// TestBackupToolsetTestSuite runs the backup toolset test suite.
func TestBackupToolsetTestSuite(t *testing.T) {
	suite.Run(t, new(BackupToolsetTestSuite))
}
