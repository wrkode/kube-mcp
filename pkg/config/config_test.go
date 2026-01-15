package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

// ConfigTestSuite tests configuration loading and merging.
type ConfigTestSuite struct {
	suite.Suite
	tempDir string
}

// SetupTest sets up the test.
func (s *ConfigTestSuite) SetupTest() {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "kube-mcp-config-test-*")
	s.Require().NoError(err, "Failed to create temp directory")
	s.tempDir = tempDir
}

// TearDownTest cleans up after the test.
func (s *ConfigTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// TestLoadBaseConfig tests loading a base configuration file.
func (s *ConfigTestSuite) TestLoadBaseConfig() {
	// Create base config file
	baseConfig := `
[server]
log_level = "debug"
transports = ["stdio"]

[kubernetes]
qps = 100
burst = 200
`
	basePath := filepath.Join(s.tempDir, "config.toml")
	err := os.WriteFile(basePath, []byte(baseConfig), 0644)
	s.Require().NoError(err, "Failed to write base config")

	loader := NewLoader(basePath, "")
	cfg, err := loader.Load()
	s.Require().NoError(err, "Failed to load config")
	s.Require().NotNil(cfg, "Config should not be nil")
	s.Equal("debug", cfg.Server.LogLevel, "Log level should match")
	s.Equal([]string{"stdio"}, cfg.Server.Transports, "Transports should match")
	s.Equal(float32(100), cfg.Kubernetes.QPS, "QPS should match")
	s.Equal(200, cfg.Kubernetes.Burst, "Burst should match")
}

// TestLoadWithDropIn tests loading with drop-in files.
func (s *ConfigTestSuite) TestLoadWithDropIn() {
	// Create base config
	baseConfig := `
[server]
log_level = "info"
transports = ["stdio"]

[kubernetes]
qps = 50
`
	basePath := filepath.Join(s.tempDir, "config.toml")
	err := os.WriteFile(basePath, []byte(baseConfig), 0644)
	s.Require().NoError(err, "Failed to write base config")

	// Create drop-in directory
	confDPath := filepath.Join(s.tempDir, "conf.d")
	err = os.MkdirAll(confDPath, 0755)
	s.Require().NoError(err, "Failed to create conf.d directory")

	// Create drop-in file
	dropInConfig := `
[server]
log_level = "debug"

[kubernetes]
burst = 200
`
	dropInPath := filepath.Join(confDPath, "override.toml")
	err = os.WriteFile(dropInPath, []byte(dropInConfig), 0644)
	s.Require().NoError(err, "Failed to write drop-in config")

	loader := NewLoader(basePath, confDPath)
	cfg, err := loader.Load()
	s.Require().NoError(err, "Failed to load config")
	s.Require().NotNil(cfg, "Config should not be nil")

	// Drop-in should override base
	s.Equal("debug", cfg.Server.LogLevel, "Log level should be overridden")
	s.Equal(float32(50), cfg.Kubernetes.QPS, "QPS should remain from base")
	s.Equal(200, cfg.Kubernetes.Burst, "Burst should be from drop-in")
}

// TestDeepMerge tests deep merging of nested structures.
func (s *ConfigTestSuite) TestDeepMerge() {
	baseConfig := `
[server]
log_level = "info"

[server.http]
address = "0.0.0.0:8080"
`
	basePath := filepath.Join(s.tempDir, "config.toml")
	err := os.WriteFile(basePath, []byte(baseConfig), 0644)
	s.Require().NoError(err, "Failed to write base config")

	dropInConfig := `
[server.http]
address = "0.0.0.0:9090"
`
	confDPath := filepath.Join(s.tempDir, "conf.d")
	err = os.MkdirAll(confDPath, 0755)
	s.Require().NoError(err, "Failed to create conf.d directory")
	dropInPath := filepath.Join(confDPath, "override.toml")
	err = os.WriteFile(dropInPath, []byte(dropInConfig), 0644)
	s.Require().NoError(err, "Failed to write drop-in config")

	loader := NewLoader(basePath, confDPath)
	cfg, err := loader.Load()
	s.Require().NoError(err, "Failed to load config")
	s.Require().NotNil(cfg, "Config should not be nil")

	// Deep merge should update address
	s.Equal("0.0.0.0:9090", cfg.Server.HTTP.Address, "Address should be overridden")
}

// TestMissingConfigFile tests handling of missing config file.
func (s *ConfigTestSuite) TestMissingConfigFile() {
	loader := NewLoader("/nonexistent/config.toml", "")
	cfg, err := loader.Load()
	s.Require().NoError(err, "Missing config file should not error")
	s.Require().NotNil(cfg, "Config should still be created with defaults")
}

// TestConfigTestSuite runs the config test suite.
func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
