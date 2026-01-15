package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/wrkode/kube-mcp/pkg/config"
	"github.com/wrkode/kube-mcp/pkg/kubernetes"
	"golang.org/x/oauth2"
)

// OAuthMiddleware provides OAuth2/OIDC authentication middleware.
type OAuthMiddleware struct {
	config        *config.OAuth2Config
	verifier      TokenVerifier
	k8sVerifier   *KubernetesTokenVerifier
	validateToken bool
}

// TokenVerifier verifies OAuth tokens.
type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) error
}

// NewOAuthMiddleware creates a new OAuth middleware.
func NewOAuthMiddleware(cfg *config.OAuth2Config, clientSet *kubernetes.ClientSet, securityCfg *config.SecurityConfig) (*OAuthMiddleware, error) {
	var verifier TokenVerifier
	var err error

	switch cfg.Provider {
	case "oidc":
		verifier, err = NewOIDCVerifier(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create OIDC verifier: %w", err)
		}
	case "oauth2":
		verifier, err = NewOAuth2Verifier(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create OAuth2 verifier: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", cfg.Provider)
	}

	var k8sVerifier *KubernetesTokenVerifier
	validateToken := true
	if securityCfg != nil {
		validateToken = securityCfg.ValidateToken
		if validateToken && clientSet != nil {
			k8sVerifier = NewKubernetesTokenVerifier(clientSet)
		}
	}

	return &OAuthMiddleware{
		config:        cfg,
		verifier:      verifier,
		k8sVerifier:   k8sVerifier,
		validateToken: validateToken,
	}, nil
}

// Middleware returns an HTTP middleware function for OAuth authentication.
func (m *OAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Verify token with OAuth provider (if configured)
		if m.verifier != nil {
			if err := m.verifier.VerifyToken(r.Context(), token); err != nil {
				http.Error(w, fmt.Sprintf("Token verification failed: %v", err), http.StatusUnauthorized)
				return
			}
		}

		// Validate token with Kubernetes TokenReview if enabled
		if m.validateToken && m.k8sVerifier != nil {
			if err := m.k8sVerifier.VerifyToken(r.Context(), token); err != nil {
				http.Error(w, fmt.Sprintf("Kubernetes token validation failed: %v", err), http.StatusUnauthorized)
				return
			}
		}

		// Store token in context for use by tool handlers
		ctx := context.WithValue(r.Context(), "bearer_token", token)
		r = r.WithContext(ctx)

		// Token is valid, proceed
		next.ServeHTTP(w, r)
	})
}

// OIDCVerifier verifies OIDC tokens.
type OIDCVerifier struct {
	provider *oauth2.Config
}

// NewOIDCVerifier creates a new OIDC verifier.
func NewOIDCVerifier(cfg *config.OAuth2Config) (*OIDCVerifier, error) {
	// For OIDC, we would typically use a library like github.com/coreos/go-oidc
	// For now, this is a placeholder implementation
	return &OIDCVerifier{}, nil
}

// VerifyToken verifies an OIDC token.
func (v *OIDCVerifier) VerifyToken(ctx context.Context, token string) error {
	// TODO: Implement OIDC token verification
	// This would use the OIDC provider to verify the token
	return nil
}

// OAuth2Verifier verifies OAuth2 tokens.
type OAuth2Verifier struct {
	config *oauth2.Config
}

// NewOAuth2Verifier creates a new OAuth2 verifier.
func NewOAuth2Verifier(cfg *config.OAuth2Config) (*OAuth2Verifier, error) {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.IssuerURL + "/auth",
			TokenURL: cfg.IssuerURL + "/token",
		},
		RedirectURL: cfg.RedirectURL,
	}

	return &OAuth2Verifier{
		config: oauth2Config,
	}, nil
}

// VerifyToken verifies an OAuth2 token.
func (v *OAuth2Verifier) VerifyToken(ctx context.Context, token string) error {
	// For OAuth2, we would verify the token with the provider
	// This is a placeholder implementation
	return nil
}
