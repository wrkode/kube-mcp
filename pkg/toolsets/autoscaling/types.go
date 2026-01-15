package autoscaling

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// HPASummary represents a normalized HPA summary.
type HPASummary struct {
	Name            string           `json:"name"`
	Namespace       string           `json:"namespace"`
	CurrentReplicas *int32           `json:"current_replicas,omitempty"`
	DesiredReplicas *int32           `json:"desired_replicas,omitempty"`
	MinReplicas     *int32           `json:"min_replicas,omitempty"`
	MaxReplicas     *int32           `json:"max_replicas,omitempty"`
	Metrics         []HPAMetric      `json:"metrics,omitempty"`
	Conditions      []map[string]any `json:"conditions,omitempty"`
	LastScaleTime   *string          `json:"last_scale_time,omitempty"` // RFC3339 string
}

// HPAMetric represents a metric target and current value.
type HPAMetric struct {
	Type       string  `json:"type"` // "Resource", "Pods", "Object", "External"
	Name       string  `json:"name,omitempty"`
	Target     *int32  `json:"target,omitempty"`
	Current    *int32  `json:"current,omitempty"`
	TargetAvg  *string `json:"target_avg,omitempty"` // For average-based metrics
	CurrentAvg *string `json:"current_avg,omitempty"`
}

// KEDAScaledObjectSummary represents a normalized KEDA ScaledObject summary.
type KEDAScaledObjectSummary struct {
	Name            string   `json:"name"`
	Namespace       string   `json:"namespace"`
	TargetKind      string   `json:"target_kind,omitempty"`
	TargetName      string   `json:"target_name,omitempty"`
	MinReplicas     *int32   `json:"min_replicas,omitempty"`
	MaxReplicas     *int32   `json:"max_replicas,omitempty"`
	CurrentReplicas *int32   `json:"current_replicas,omitempty"`
	Paused          bool     `json:"paused"`
	Triggers        []string `json:"triggers,omitempty"`
	Status          string   `json:"status,omitempty"`
}

// KEDATriggerSummary represents a trigger summary.
type KEDATriggerSummary struct {
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Status   string                 `json:"status,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

// GVKs for KEDA CRDs
var (
	ScaledObjectGVK = schema.GroupVersionKind{
		Group:   "keda.sh",
		Version: "v1alpha1",
		Kind:    "ScaledObject",
	}
	ScaledJobGVK = schema.GroupVersionKind{
		Group:   "keda.sh",
		Version: "v1alpha1",
		Kind:    "ScaledJob",
	}
)

// KEDA pause/resume annotations (documented stable annotations)
const (
	KEDAPauseAnnotation = "autoscaling.keda.sh/paused"
)
