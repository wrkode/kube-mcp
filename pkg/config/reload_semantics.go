package config

// ReloadableSettings defines which configuration settings can be reloaded at runtime.
// Settings not listed here require a server restart.
type ReloadableSettings struct {
	// Logging settings
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`

	// Toolset enabling/disabling
	ToolsetsEnabled map[string]bool `toml:"toolsets_enabled"`

	// Kiali settings (if enabled)
	KialiEnabled bool   `toml:"kiali_enabled"`
	KialiURL     string `toml:"kiali_url"`
	KialiToken   string `toml:"kiali_token"`
	KialiTimeout string `toml:"kiali_timeout"`

	// KubeVirt settings (if enabled)
	KubeVirtEnabled bool `toml:"kubevirt_enabled"`

	// OAuth settings
	OAuthValidateToken  bool `toml:"oauth_validate_token"`
	OAuthPropagateToken bool `toml:"oauth_propagate_token"`

	// Rate limiting
	RateLimitEnabled bool `toml:"rate_limit_enabled"`
	RateLimitRPS     int  `toml:"rate_limit_rps"`

	// Metrics
	MetricsEnabled bool `toml:"metrics_enabled"`
}

// RestartRequiredSettings defines which settings require a server restart.
// These include:
// - Transport configuration (stdio/http)
// - Ports and addresses
// - Kubeconfig paths and provider type
// - TLS certificates
// - OAuth provider URLs and client credentials
type RestartRequiredSettings struct {
	// Server configuration
	Transports []string `toml:"transports"`
	HTTPPort   int      `toml:"http_port"`

	// Kubernetes provider
	ProviderType   string `toml:"provider_type"`
	KubeconfigPath string `toml:"kubeconfig_path"`
	Context        string `toml:"context"`

	// TLS
	TLSCertFile string `toml:"tls_cert_file"`
	TLSKeyFile  string `toml:"tls_key_file"`

	// OAuth
	OAuthProvider string `toml:"oauth_provider"`
	OAuthURL      string `toml:"oauth_url"`
	OAuthClientID string `toml:"oauth_client_id"`
}

// CanReload checks if a configuration change can be applied at runtime.
func CanReload(setting string) bool {
	reloadable := map[string]bool{
		"log_level":             true,
		"log_format":            true,
		"toolsets_enabled":      true,
		"kiali_enabled":         true,
		"kiali_url":             true,
		"kiali_token":           true,
		"kiali_timeout":         true,
		"kubevirt_enabled":      true,
		"oauth_validate_token":  true,
		"oauth_propagate_token": true,
		"rate_limit_enabled":    true,
		"rate_limit_rps":        true,
		"metrics_enabled":       true,
	}
	return reloadable[setting]
}

// RequiresRestart checks if a configuration change requires a server restart.
func RequiresRestart(setting string) bool {
	return !CanReload(setting)
}
