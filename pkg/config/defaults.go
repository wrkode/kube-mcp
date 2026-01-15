package config

import (
	"time"
)

// applyDefaults applies default values to the configuration.
func (l *Loader) applyDefaults(cfg *Config) {
	// Server defaults
	if len(cfg.Server.Transports) == 0 {
		cfg.Server.Transports = []string{"stdio"}
	}
	if cfg.Server.LogLevel == "" {
		cfg.Server.LogLevel = "info"
	}

	// HTTP defaults
	if cfg.Server.HTTP.Address == "" {
		cfg.Server.HTTP.Address = "0.0.0.0:8080"
	}
	if len(cfg.Server.HTTP.CORS.AllowedMethods) == 0 {
		cfg.Server.HTTP.CORS.AllowedMethods = []string{"GET", "POST", "OPTIONS"}
	}
	if len(cfg.Server.HTTP.CORS.AllowedHeaders) == 0 {
		cfg.Server.HTTP.CORS.AllowedHeaders = []string{"Content-Type", "Authorization"}
	}
	if len(cfg.Server.HTTP.OAuth.Scopes) == 0 {
		cfg.Server.HTTP.OAuth.Scopes = []string{"openid", "profile"}
	}
	if cfg.Server.HTTP.OAuth.Provider == "" {
		cfg.Server.HTTP.OAuth.Provider = "oidc"
	}

	// Kubernetes defaults
	if cfg.Kubernetes.Provider == "" {
		cfg.Kubernetes.Provider = "kubeconfig"
	}
	if cfg.Kubernetes.KubeconfigPath == "" {
		cfg.Kubernetes.KubeconfigPath = "~/.kube/config"
	}
	if cfg.Kubernetes.QPS == 0 {
		cfg.Kubernetes.QPS = 100
	}
	if cfg.Kubernetes.Burst == 0 {
		cfg.Kubernetes.Burst = 200
	}
	if cfg.Kubernetes.Timeout == 0 {
		cfg.Kubernetes.Timeout = Duration(30 * time.Second)
	}

	// Security defaults
	if cfg.Security.RequireRBAC {
		cfg.Security.RequireRBAC = true
	}

	// Helm defaults
	if cfg.Helm.StorageDriver == "" {
		cfg.Helm.StorageDriver = "secret"
	}
	if cfg.Helm.DefaultNamespace == "" {
		cfg.Helm.DefaultNamespace = "default"
	}

	// KubeVirt defaults
	if cfg.KubeVirt.VMGroupVersion == "" {
		cfg.KubeVirt.VMGroupVersion = "kubevirt.io/v1"
	}
	if !cfg.KubeVirt.Enabled {
		cfg.KubeVirt.Enabled = true // Auto-enable, will be disabled if CRDs not found
	}

	// Kiali defaults
	if cfg.Kiali.Timeout == 0 {
		cfg.Kiali.Timeout = Duration(30 * time.Second)
	}

	// Network defaults
	if cfg.Toolsets.Net.HubbleTimeout == 0 {
		cfg.Toolsets.Net.HubbleTimeout = Duration(10 * time.Second)
	}
}
