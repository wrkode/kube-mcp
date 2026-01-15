package autoscaling

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// normalizeHPASummary normalizes an HPA into a summary.
func (t *Toolset) normalizeHPASummary(hpa *autoscalingv2.HorizontalPodAutoscaler) HPASummary {
	summary := HPASummary{
		Name:      hpa.Name,
		Namespace: hpa.Namespace,
	}

	if hpa.Spec.MinReplicas != nil {
		summary.MinReplicas = hpa.Spec.MinReplicas
	}
	summary.MaxReplicas = &hpa.Spec.MaxReplicas

	if hpa.Status.CurrentReplicas > 0 {
		current := int32(hpa.Status.CurrentReplicas)
		summary.CurrentReplicas = &current
	}
	if hpa.Status.DesiredReplicas > 0 {
		desired := int32(hpa.Status.DesiredReplicas)
		summary.DesiredReplicas = &desired
	}

	// Extract metrics
	for _, metric := range hpa.Spec.Metrics {
		hpaMetric := HPAMetric{
			Type: string(metric.Type),
		}

		switch metric.Type {
		case autoscalingv2.ResourceMetricSourceType:
			if metric.Resource != nil {
				hpaMetric.Name = string(metric.Resource.Name)
				if metric.Resource.Target.Type == autoscalingv2.UtilizationMetricType {
					if metric.Resource.Target.AverageUtilization != nil {
						avg := fmt.Sprintf("%d%%", *metric.Resource.Target.AverageUtilization)
						hpaMetric.TargetAvg = &avg
					}
				}
			}
		case autoscalingv2.PodsMetricSourceType:
			if metric.Pods != nil {
				if metric.Pods.Target.Type == autoscalingv2.AverageValueMetricType {
					avg := metric.Pods.Target.AverageValue.String()
					hpaMetric.TargetAvg = &avg
				}
			}
		case autoscalingv2.ObjectMetricSourceType:
			if metric.Object != nil {
				hpaMetric.Name = metric.Object.Metric.Name
			}
		case autoscalingv2.ExternalMetricSourceType:
			if metric.External != nil {
				hpaMetric.Name = metric.External.Metric.Name
			}
		}

		summary.Metrics = append(summary.Metrics, hpaMetric)
	}

	// Extract current metric values from status
	for i, metricStatus := range hpa.Status.CurrentMetrics {
		if i < len(summary.Metrics) {
			switch metricStatus.Type {
			case autoscalingv2.ResourceMetricSourceType:
				if metricStatus.Resource != nil && metricStatus.Resource.Current.AverageUtilization != nil {
					avg := fmt.Sprintf("%d%%", *metricStatus.Resource.Current.AverageUtilization)
					summary.Metrics[i].CurrentAvg = &avg
				}
			case autoscalingv2.PodsMetricSourceType:
				if metricStatus.Pods != nil {
					avg := metricStatus.Pods.Current.AverageValue.String()
					summary.Metrics[i].CurrentAvg = &avg
				}
			}
		}
	}

	// Extract conditions
	for _, cond := range hpa.Status.Conditions {
		condMap := map[string]any{
			"type":   cond.Type,
			"status": cond.Status,
			"reason": cond.Reason,
		}
		if cond.Message != "" {
			condMap["message"] = cond.Message
		}
		if !cond.LastTransitionTime.IsZero() {
			condMap["last_transition_time"] = cond.LastTransitionTime.Format(time.RFC3339)
		}
		summary.Conditions = append(summary.Conditions, condMap)
	}

	// Extract last scale time
	if hpa.Status.LastScaleTime != nil {
		lastScaleTime := hpa.Status.LastScaleTime.Format(time.RFC3339)
		summary.LastScaleTime = &lastScaleTime
	}

	return summary
}

// handleHPAList handles the autoscaling.hpa_list tool.
func (t *Toolset) handleHPAList(ctx context.Context, args struct {
	Context       string `json:"context"`
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"label_selector"`
	Limit         int    `json:"limit"`
	Continue      string `json:"continue"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: args.LabelSelector,
	}
	if args.Limit > 0 {
		listOptions.Limit = int64(args.Limit)
	}
	if args.Continue != "" {
		listOptions.Continue = args.Continue
	}

	var list *autoscalingv2.HorizontalPodAutoscalerList
	if args.Namespace != "" {
		list, err = clientSet.Typed.AutoscalingV2().HorizontalPodAutoscalers(args.Namespace).List(ctx, listOptions)
	} else {
		list, err = clientSet.Typed.AutoscalingV2().HorizontalPodAutoscalers("").List(ctx, listOptions)
	}

	if err != nil {
		// Fallback to v2beta2 if v2 not available
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to list HPAs: %w", err)), nil
	}

	var hpas []HPASummary
	for _, item := range list.Items {
		hpas = append(hpas, t.normalizeHPASummary(&item))
	}

	resultData := map[string]any{
		"items": hpas,
	}
	if list.Continue != "" {
		resultData["continue"] = list.Continue
		if list.RemainingItemCount != nil {
			resultData["remaining_item_count"] = *list.RemainingItemCount
		}
	}

	result, err := mcpHelpers.NewJSONResult(resultData)
	return result, err
}

// handleHPAExplain handles the autoscaling.hpa_explain tool.
func (t *Toolset) handleHPAExplain(ctx context.Context, args struct {
	Context   string `json:"context"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}) (*mcp.CallToolResult, error) {
	clientSet, err := t.provider.GetClientSet(args.Context)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get client set: %w", err)), nil
	}

	hpa, err := clientSet.Typed.AutoscalingV2().HorizontalPodAutoscalers(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to get HPA: %w", err)), nil
	}

	summary := t.normalizeHPASummary(hpa)
	result, err := mcpHelpers.NewJSONResult(summary)
	return result, err
}
