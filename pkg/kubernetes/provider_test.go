package kubernetes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// ProviderTestSuite tests Kubernetes provider functionality.
type ProviderTestSuite struct {
	suite.Suite
	tempDir string
}

// SetupTest sets up the test.
func (s *ProviderTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "kubeconfig-test-*")
	s.Require().NoError(err, "Failed to create temp directory")
	s.tempDir = tempDir
}

// TearDownTest cleans up after the test.
func (s *ProviderTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// TestDefaultContextSelection tests default context selection.
func (s *ProviderTestSuite) TestDefaultContextSelection() {
	// Create a test kubeconfig with multiple contexts
	kubeconfig := api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			"cluster1": {
				Server: "https://cluster1.example.com",
			},
			"cluster2": {
				Server: "https://cluster2.example.com",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"user1": {},
			"user2": {},
		},
		Contexts: map[string]*api.Context{
			"context1": {
				Cluster:  "cluster1",
				AuthInfo: "user1",
			},
			"context2": {
				Cluster:  "cluster2",
				AuthInfo: "user2",
			},
		},
		CurrentContext: "context1",
	}

	kubeconfigPath := filepath.Join(s.tempDir, "config")
	err := clientcmd.WriteToFile(kubeconfig, kubeconfigPath)
	s.Require().NoError(err, "Failed to write kubeconfig")

	factory := NewClientFactory(100, 200, 0)
	provider, err := NewKubeconfigProvider(factory, kubeconfigPath)
	s.Require().NoError(err, "Failed to create provider")

	// Test GetCurrentContext
	currentCtx, err := provider.GetCurrentContext()
	s.Require().NoError(err, "Failed to get current context")
	s.Equal("context1", currentCtx, "Current context should be context1")

	// Test ListContexts
	contexts, err := provider.ListContexts()
	s.Require().NoError(err, "Failed to list contexts")
	s.Contains(contexts, "context1", "Should contain context1")
	s.Contains(contexts, "context2", "Should contain context2")
}

// TestWithContext tests WithContext helper.
func (s *ProviderTestSuite) TestWithContext() {
	// Create a test kubeconfig
	kubeconfig := api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			"cluster1": {
				Server: "https://cluster1.example.com",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"user1": {},
		},
		Contexts: map[string]*api.Context{
			"context1": {
				Cluster:  "cluster1",
				AuthInfo: "user1",
			},
		},
		CurrentContext: "context1",
	}

	kubeconfigPath := filepath.Join(s.tempDir, "config")
	err := clientcmd.WriteToFile(kubeconfig, kubeconfigPath)
	s.Require().NoError(err, "Failed to write kubeconfig")

	factory := NewClientFactory(100, 200, 0)
	provider, err := NewKubeconfigProvider(factory, kubeconfigPath)
	s.Require().NoError(err, "Failed to create provider")

	// Test WithContext with valid context
	clusterClient, err := provider.WithContext("context1")
	s.Require().NoError(err, "WithContext should succeed")
	s.Require().NotNil(clusterClient, "ClusterClient should not be nil")
	s.Equal("context1", clusterClient.GetContext(), "Context should match")

	// Test WithContext with empty context (should use default)
	clusterClient, err = provider.WithContext("")
	s.Require().NoError(err, "WithContext with empty context should succeed")
	s.Require().NotNil(clusterClient, "ClusterClient should not be nil")

	// Test WithContext with unknown context
	_, err = provider.WithContext("unknown-context")
	s.Require().Error(err, "WithContext with unknown context should fail")
}

// TestUnknownContextError tests error handling for unknown contexts.
func (s *ProviderTestSuite) TestUnknownContextError() {
	kubeconfig := api.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*api.Cluster{
			"cluster1": {
				Server: "https://cluster1.example.com",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"user1": {},
		},
		Contexts: map[string]*api.Context{
			"context1": {
				Cluster:  "cluster1",
				AuthInfo: "user1",
			},
		},
		CurrentContext: "context1",
	}

	kubeconfigPath := filepath.Join(s.tempDir, "config")
	err := clientcmd.WriteToFile(kubeconfig, kubeconfigPath)
	s.Require().NoError(err, "Failed to write kubeconfig")

	factory := NewClientFactory(100, 200, 0)
	provider, err := NewKubeconfigProvider(factory, kubeconfigPath)
	s.Require().NoError(err, "Failed to create provider")

	// Try to get client set for unknown context
	_, err = provider.GetClientSet("unknown-context")
	s.Require().Error(err, "Should error on unknown context")
	s.Contains(err.Error(), "unknown-context", "Error should mention the context name")
}

// TestProviderTestSuite runs the provider test suite.
func TestProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ProviderTestSuite))
}

