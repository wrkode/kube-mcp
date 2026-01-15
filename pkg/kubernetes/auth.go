package kubernetes

import (
	"context"
	"fmt"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// TokenReviewer validates Bearer tokens using Kubernetes TokenReview API.
type TokenReviewer struct {
	clientSet *ClientSet
}

// NewTokenReviewer creates a new TokenReviewer.
func NewTokenReviewer(clientSet *ClientSet) *TokenReviewer {
	return &TokenReviewer{
		clientSet: clientSet,
	}
}

// ValidateToken validates a Bearer token using TokenReview.
// Returns the authenticated user info if valid, or an error if invalid.
func (t *TokenReviewer) ValidateToken(ctx context.Context, token string) (*authenticationv1.UserInfo, error) {
	tr := &authenticationv1.TokenReview{
		Spec: authenticationv1.TokenReviewSpec{
			Token: token,
		},
	}

	result, err := t.clientSet.Typed.AuthenticationV1().TokenReviews().Create(ctx, tr, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create token review: %w", err)
	}

	if !result.Status.Authenticated {
		return nil, fmt.Errorf("token is not authenticated: %s", result.Status.Error)
	}

	return &result.Status.User, nil
}

// CredentialSelector selects credentials for Kubernetes API calls.
type CredentialSelector struct {
	// BearerToken is the Bearer token to use if provided.
	BearerToken string

	// UseServiceAccount indicates whether to use service account credentials.
	UseServiceAccount bool
}

// SelectCredentials selects the appropriate REST config based on credentials.
// If BearerToken is provided, it takes precedence.
// Otherwise, uses the original config (which may have service account credentials).
func SelectCredentials(originalConfig *rest.Config, selector *CredentialSelector) *rest.Config {
	if selector == nil {
		return originalConfig
	}

	// Create a copy of the config
	config := *originalConfig

	// If Bearer token is provided, use it
	if selector.BearerToken != "" {
		config.BearerToken = selector.BearerToken
		config.BearerTokenFile = "" // Clear BearerTokenFile if token is provided
	}

	return &config
}
