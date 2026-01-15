package gitops

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// AppKind represents the type of GitOps application.
type AppKind string

const (
	AppKindKustomization AppKind = "Kustomization"
	AppKindHelmRelease   AppKind = "HelmRelease"
	AppKindApplication   AppKind = "Application"
)

// AppStatus represents the status of a GitOps application.
type AppStatus string

const (
	AppStatusReady       AppStatus = "Ready"
	AppStatusProgressing AppStatus = "Progressing"
	AppStatusDegraded    AppStatus = "Degraded"
	AppStatusUnknown     AppStatus = "Unknown"
)

// AppSummary represents a normalized GitOps application summary.
type AppSummary struct {
	Kind        string    `json:"kind"`
	Name        string    `json:"name"`
	Namespace   string    `json:"namespace"`
	Ready       bool      `json:"ready"`
	Status      AppStatus `json:"status"`
	LastUpdated *string   `json:"last_updated,omitempty"` // RFC3339 string
	Revision    string    `json:"revision,omitempty"`
	Artifact    string    `json:"artifact,omitempty"`
	Message     string    `json:"message,omitempty"`
}

// AppCondition represents a normalized condition.
type AppCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
	Time    string `json:"time,omitempty"` // RFC3339 string
}

// AppDetails represents detailed information about a GitOps application.
type AppDetails struct {
	AppSummary
	Conditions []AppCondition `json:"conditions,omitempty"`
}

// GVKs for GitOps CRDs
var (
	// Flux CRDs
	KustomizationGVK = schema.GroupVersionKind{
		Group:   "kustomize.toolkit.fluxcd.io",
		Version: "v1",
		Kind:    "Kustomization",
	}
	HelmReleaseGVK = schema.GroupVersionKind{
		Group:   "helm.toolkit.fluxcd.io",
		Version: "v2",
		Kind:    "HelmRelease",
	}
	// Argo CD CRD
	ApplicationGVK = schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Application",
	}
)

// Flux reconcile annotation keys (documented stable annotations)
const (
	// Flux Kustomization reconcile annotation
	FluxKustomizationReconcileAnnotation = "kustomize.toolkit.fluxcd.io/reconcile"
	// Flux HelmRelease reconcile annotation
	FluxHelmReleaseReconcileAnnotation = "helm.toolkit.fluxcd.io/reconcile"
)

// Argo CD refresh annotation (best-effort refresh trigger)
const (
	ArgoCDRefreshAnnotation = "argocd.argoproj.io/refresh"
)
