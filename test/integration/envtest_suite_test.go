package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// EnvtestSuite provides shared setup for integration tests.
type EnvtestSuite struct {
	suite.Suite
	envtest    *envtest.Environment
	restConfig *rest.Config
	clientSet  *kubernetes.ClientSet
	provider   kubernetes.ClientProvider
	ctx        context.Context
	cancel     context.CancelFunc
}

// SetupSuite sets up the test environment before all tests.
func (s *EnvtestSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Configure envtest
	// Load CRDs from all subdirectories
	crdBasePath := filepath.Join("..", "..", "testdata", "crds")
	s.envtest = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join(crdBasePath, "capi"),
			filepath.Join(crdBasePath, "flux"),
			filepath.Join(crdBasePath, "policy"),
			filepath.Join(crdBasePath, "rollouts"),
			filepath.Join(crdBasePath, "certs"),
			filepath.Join(crdBasePath, "autoscaling"),
			filepath.Join(crdBasePath, "backup"),
			filepath.Join(crdBasePath, "net"),
		},
		ErrorIfCRDPathMissing: false,
		UseExistingCluster:    nil, // Use embedded etcd + API server
	}

	// Start the test environment
	cfg, err := s.envtest.Start()
	s.Require().NoError(err, "Failed to start test environment")
	s.restConfig = cfg

	// Create client factory
	factory := kubernetes.NewClientFactory(100, 200, 30*time.Second)

	// Create client set
	clientSet, err := factory.CreateClientSet(cfg)
	s.Require().NoError(err, "Failed to create client set")
	s.clientSet = clientSet

	// Create a kubeconfig provider for testing
	// We'll use a simple in-memory provider that wraps our test client
	provider := &testProvider{
		clientSet: clientSet,
	}
	s.provider = provider
}

// TearDownSuite tears down the test environment after all tests.
func (s *EnvtestSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}
	if s.envtest != nil {
		s.NoError(s.envtest.Stop(), "Failed to stop test environment")
	}
}

// SetupTest is called before each test.
func (s *EnvtestSuite) SetupTest() {
	// Each test can override this if needed
}

// TearDownTest is called after each test.
func (s *EnvtestSuite) TearDownTest() {
	// Cleanup can be added here if needed
}

// testProvider is a simple provider for testing that always returns the same client set.
type testProvider struct {
	clientSet *kubernetes.ClientSet
}

func (p *testProvider) GetClientSet(context string) (*kubernetes.ClientSet, error) {
	return p.clientSet, nil
}

func (p *testProvider) ListContexts() ([]string, error) {
	return []string{"test-context"}, nil
}

func (p *testProvider) GetCurrentContext() (string, error) {
	return "test-context", nil
}

func (p *testProvider) WithContext(ctx string) (kubernetes.ClusterClient, error) {
	return &testClusterClient{clientSet: p.clientSet, context: ctx}, nil
}

// testClusterClient implements ClusterClient for testing.
type testClusterClient struct {
	clientSet *kubernetes.ClientSet
	context   string
}

func (c *testClusterClient) GetClientSet() *kubernetes.ClientSet {
	return c.clientSet
}

func (c *testClusterClient) GetContext() string {
	return c.context
}

// TestEnvtestSuite runs the envtest suite.
func TestEnvtestSuite(t *testing.T) {
	suite.Run(t, new(EnvtestSuite))
}
