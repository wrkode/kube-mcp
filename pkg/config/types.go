package config

import (
	"fmt"
	"time"
)

// Duration is a custom type that wraps time.Duration and supports TOML string parsing.
type Duration time.Duration

// UnmarshalText implements encoding.TextUnmarshaler for TOML parsing.
func (d *Duration) UnmarshalText(text []byte) error {
	duration, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", string(text), err)
	}
	*d = Duration(duration)
	return nil
}

// Duration returns the time.Duration value.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// String implements fmt.Stringer.
func (d Duration) String() string {
	return time.Duration(d).String()
}

// Config represents the complete configuration for kube-mcp.
type Config struct {
	Server     ServerConfig     `toml:"server"`
	Kubernetes KubernetesConfig `toml:"kubernetes"`
	Security   SecurityConfig   `toml:"security"`
	Helm       HelmConfig       `toml:"helm"`
	KubeVirt   KubeVirtConfig   `toml:"kubevirt"`
	Kiali      KialiConfig      `toml:"kiali"`
	Toolsets   ToolsetsConfig   `toml:"toolsets"`
}

// ServerConfig contains server-level configuration.
type ServerConfig struct {
	// Transports enabled for the server
	Transports []string `toml:"transports"` // "stdio", "http"

	// HTTP server configuration
	HTTP HTTPConfig `toml:"http"`

	// Log level: "debug", "info", "warn", "error"
	LogLevel string `toml:"log_level" default:"info"`

	// Normalize tool names by replacing dots with underscores (for n8n compatibility)
	// When enabled, "autoscaling.hpa_explain" becomes "autoscaling_hpa_explain"
	NormalizeToolNames bool `toml:"normalize_tool_names"`
}

// HTTPConfig contains HTTP transport configuration.
type HTTPConfig struct {
	// Address to bind the HTTP server
	Address string `toml:"address" default:"0.0.0.0:8080"`

	// Enable OAuth2/OIDC authentication
	OAuth OAuth2Config `toml:"oauth"`

	// CORS configuration
	CORS CORSConfig `toml:"cors"`
}

// OAuth2Config contains OAuth2/OIDC configuration.
type OAuth2Config struct {
	// Enable OAuth2/OIDC
	Enabled bool `toml:"enabled" default:"false"`

	// Provider: "oidc", "oauth2"
	Provider string `toml:"provider" default:"oidc"`

	// Issuer URL
	IssuerURL string `toml:"issuer_url"`

	// Client ID
	ClientID string `toml:"client_id"`

	// Client secret (can be loaded from env var)
	ClientSecret string `toml:"client_secret"`

	// Scopes
	Scopes []string `toml:"scopes" default:"[\"openid\", \"profile\"]"`

	// Redirect URL
	RedirectURL string `toml:"redirect_url"`
}

// CORSConfig contains CORS configuration.
type CORSConfig struct {
	// Enable CORS
	Enabled bool `toml:"enabled" default:"false"`

	// Allowed origins
	AllowedOrigins []string `toml:"allowed_origins"`

	// Allowed methods
	AllowedMethods []string `toml:"allowed_methods" default:"[\"GET\", \"POST\", \"OPTIONS\"]"`

	// Allowed headers
	AllowedHeaders []string `toml:"allowed_headers" default:"[\"Content-Type\", \"Authorization\"]"`
}

// KubernetesConfig contains Kubernetes client configuration.
type KubernetesConfig struct {
	// Provider strategy: "kubeconfig", "in-cluster", "single"
	Provider string `toml:"provider" default:"kubeconfig"`

	// Path to kubeconfig file (for kubeconfig provider)
	KubeconfigPath string `toml:"kubeconfig_path" default:"~/.kube/config"`

	// Context name (for single-cluster mode)
	Context string `toml:"context"`

	// Namespace default (optional)
	DefaultNamespace string `toml:"default_namespace"`

	// QPS for Kubernetes client
	QPS float32 `toml:"qps" default:"100"`

	// Burst for Kubernetes client
	Burst int `toml:"burst" default:"200"`

	// Timeout for Kubernetes API calls
	Timeout Duration `toml:"timeout" default:"30s"`
}

// SecurityConfig contains security-related configuration.
type SecurityConfig struct {
	// Enable read-only mode (prevents all write operations)
	ReadOnly bool `toml:"read_only" default:"false"`

	// Enable non-destructive mode (allows reads and applies, prevents deletes)
	NonDestructive bool `toml:"non_destructive" default:"false"`

	// List of denied GroupVersionKinds (GVKs) that cannot be accessed
	DeniedGVKs []string `toml:"denied_gvks"`

	// Require RBAC checks before operations
	RequireRBAC bool `toml:"require_rbac" default:"true"`

	// Validate Bearer tokens using Kubernetes TokenReview API
	// If false, tokens are accepted but not validated against Kubernetes
	ValidateToken bool `toml:"validate_token" default:"true"`

	// RBAC cache TTL in seconds
	RBACCacheTTL int `toml:"rbac_cache_ttl" default:"5"`
}

// HelmConfig contains Helm-specific configuration.
type HelmConfig struct {
	// Helm storage driver: "secret", "configmap", "memory"
	StorageDriver string `toml:"storage_driver" default:"secret"`

	// Default namespace for Helm operations
	DefaultNamespace string `toml:"default_namespace" default:"default"`
}

// KubeVirtConfig contains KubeVirt-specific configuration.
type KubeVirtConfig struct {
	// Enable KubeVirt toolset (auto-detected if CRDs exist)
	Enabled bool `toml:"enabled" default:"true"`

	// VirtualMachine CRD group version
	VMGroupVersion string `toml:"vm_group_version" default:"kubevirt.io/v1"`
}

// KialiConfig contains Kiali-specific configuration.
type KialiConfig struct {
	// Enable Kiali integration
	Enabled bool `toml:"enabled" default:"false"`

	// Kiali server URL
	URL string `toml:"url"`

	// TLS configuration
	TLS TLSConfig `toml:"tls"`

	// Authentication token (optional, can use OAuth)
	Token string `toml:"token"`

	// Request timeout
	Timeout Duration `toml:"timeout" default:"30s"`
}

// TLSConfig contains TLS configuration.
type TLSConfig struct {
	// Enable TLS
	Enabled bool `toml:"enabled" default:"false"`

	// CA certificate path
	CAFile string `toml:"ca_file"`

	// Client certificate path
	CertFile string `toml:"cert_file"`

	// Client key path
	KeyFile string `toml:"key_file"`

	// Skip TLS verification (insecure)
	InsecureSkipVerify bool `toml:"insecure_skip_verify" default:"false"`
}

// ToolsetsConfig contains configuration for optional toolsets.
type ToolsetsConfig struct {
	GitOps      GitOpsConfig      `toml:"gitops"`
	Policy      PolicyConfig      `toml:"policy"`
	CAPI        CAPIConfig        `toml:"capi"`
	Rollouts    RolloutsConfig    `toml:"rollouts"`
	Certs       CertsConfig       `toml:"certs"`
	Autoscaling AutoscalingConfig `toml:"autoscaling"`
	Backup      BackupConfig      `toml:"backup"`
	Net         NetConfig         `toml:"net"`
}

// GitOpsConfig contains GitOps toolset configuration.
type GitOpsConfig struct {
	// Enable GitOps toolset (auto-detected if CRDs exist)
	Enabled bool `toml:"enabled" default:"true"`
}

// PolicyConfig contains Policy toolset configuration.
type PolicyConfig struct {
	// Enable Policy toolset (auto-detected if CRDs exist)
	Enabled bool `toml:"enabled" default:"true"`
}

// CAPIConfig contains CAPI toolset configuration.
type CAPIConfig struct {
	// Enable CAPI toolset (auto-detected if CRDs exist)
	Enabled bool `toml:"enabled" default:"true"`
}

// RolloutsConfig contains Progressive Delivery toolset configuration.
type RolloutsConfig struct {
	// Enable Progressive Delivery toolset (auto-detected if CRDs exist)
	Enabled bool `toml:"enabled" default:"true"`
}

// CertsConfig contains Cert-Manager toolset configuration.
type CertsConfig struct {
	// Enable Cert-Manager toolset (auto-detected if CRDs exist)
	Enabled bool `toml:"enabled" default:"true"`
}

// AutoscalingConfig contains Autoscaling toolset configuration.
type AutoscalingConfig struct {
	// Enable Autoscaling toolset (HPA always available, KEDA auto-detected)
	Enabled bool `toml:"enabled" default:"true"`
}

// BackupConfig contains Backup/Restore toolset configuration.
type BackupConfig struct {
	// Enable Backup/Restore toolset (auto-detected if CRDs exist)
	Enabled bool `toml:"enabled" default:"true"`
}

// NetConfig contains Network toolset configuration.
type NetConfig struct {
	// Enable Network toolset (NetworkPolicy always available, Cilium auto-detected)
	Enabled bool `toml:"enabled" default:"true"`

	// Hubble API URL (optional, for flow queries)
	HubbleAPIURL string `toml:"hubble_api_url"`

	// Hubble TLS configuration
	HubbleInsecure bool   `toml:"hubble_insecure" default:"false"`
	HubbleCAFile   string `toml:"hubble_ca_file"`

	// Hubble request timeout
	HubbleTimeout Duration `toml:"hubble_timeout" default:"10s"`
}
