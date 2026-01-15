package rollouts

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RolloutKind represents the type of progressive delivery resource.
type RolloutKind string

const (
	RolloutKindRollout RolloutKind = "Rollout"
	RolloutKindCanary  RolloutKind = "Canary"
)

// RolloutPhase represents the phase/status of a rollout.
type RolloutPhase string

const (
	RolloutPhaseProgressing RolloutPhase = "Progressing"
	RolloutPhasePaused      RolloutPhase = "Paused"
	RolloutPhaseDegraded    RolloutPhase = "Degraded"
	RolloutPhaseHealthy     RolloutPhase = "Healthy"
	RolloutPhaseUnknown     RolloutPhase = "Unknown"
)

// RolloutSummary represents a normalized progressive delivery resource summary.
type RolloutSummary struct {
	Kind        string       `json:"kind"`
	Name        string       `json:"name"`
	Namespace   string       `json:"namespace"`
	Phase       RolloutPhase `json:"phase"`
	Ready       bool         `json:"ready"`
	Message     string       `json:"message,omitempty"`
	LastUpdated *string      `json:"last_updated,omitempty"` // RFC3339 string
	Revision    string       `json:"revision,omitempty"`
}

// RolloutStatus represents detailed status information.
type RolloutStatus struct {
	RolloutSummary
	Revisions      []string                 `json:"revisions,omitempty"`
	CurrentStep    *int                     `json:"current_step,omitempty"`
	TotalSteps     *int                     `json:"total_steps,omitempty"`
	TrafficWeight  *int                     `json:"traffic_weight,omitempty"` // 0-100
	Paused         bool                     `json:"paused"`
	AnalysisStatus map[string]interface{}   `json:"analysis_status,omitempty"`
	Conditions     []map[string]interface{} `json:"conditions,omitempty"`
}

// GVKs for Progressive Delivery CRDs
var (
	// Argo Rollouts CRD
	RolloutGVK = schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "Rollout",
	}
	// Flagger Canary CRD
	CanaryGVK = schema.GroupVersionKind{
		Group:   "flagger.app",
		Version: "v1beta1",
		Kind:    "Canary",
	}
)

// Argo Rollouts annotations (documented stable annotations)
const (
	// Argo Rollouts promote annotation
	ArgoRolloutsPromoteAnnotation = "rollouts.argoproj.io/promote"
	// Argo Rollouts abort annotation
	ArgoRolloutsAbortAnnotation = "rollouts.argoproj.io/abort"
	// Argo Rollouts retry annotation
	ArgoRolloutsRetryAnnotation = "rollouts.argoproj.io/retry"
)
