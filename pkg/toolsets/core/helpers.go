package core

import (
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
)

// getClusterClient gets a cluster client for the given context.
// If context is empty, uses the provider's default context.
func (t *Toolset) getClusterClient(context string) (*kubernetes.ClientSet, error) {
	return t.provider.GetClientSet(context)
}

// getContextOrDefault returns the context or the default if empty.
func (t *Toolset) getContextOrDefault(context string) string {
	if context == "" {
		defaultCtx, err := t.provider.GetCurrentContext()
		if err != nil {
			return ""
		}
		return defaultCtx
	}
	return context
}
