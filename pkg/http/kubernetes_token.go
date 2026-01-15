package http

import (
	"context"
	"fmt"

	"github.com/wrkode/kube-mcp/pkg/kubernetes"
)

// KubernetesTokenVerifier verifies Bearer tokens using Kubernetes TokenReview API.
type KubernetesTokenVerifier struct {
	clientSet *kubernetes.ClientSet
}

// NewKubernetesTokenVerifier creates a new Kubernetes token verifier.
func NewKubernetesTokenVerifier(clientSet *kubernetes.ClientSet) *KubernetesTokenVerifier {
	return &KubernetesTokenVerifier{
		clientSet: clientSet,
	}
}

// VerifyToken verifies a Bearer token using TokenReview.
func (v *KubernetesTokenVerifier) VerifyToken(ctx context.Context, token string) error {
	reviewer := kubernetes.NewTokenReviewer(v.clientSet)
	_, err := reviewer.ValidateToken(ctx, token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}
	return nil
}
