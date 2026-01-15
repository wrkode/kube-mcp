package observability

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics provides Prometheus metrics for kube-mcp.
type Metrics struct {
	toolCallsTotal    *prometheus.CounterVec
	toolLatency       *prometheus.HistogramVec
	httpRequestsTotal *prometheus.CounterVec
	httpLatency       *prometheus.HistogramVec
}

// NewMetrics creates a new metrics collector.
// If registry is nil, uses the default Prometheus registry.
func NewMetrics(registry prometheus.Registerer) *Metrics {
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}

	// Use promauto factory with custom registry
	factory := promauto.With(registry)

	return &Metrics{
		toolCallsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kube_mcp_tool_calls_total",
				Help: "Total number of MCP tool calls",
			},
			[]string{"tool", "context", "success"},
		),
		toolLatency: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kube_mcp_tool_latency_seconds",
				Help:    "Latency of MCP tool calls in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"tool", "context"},
		),
		httpRequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kube_mcp_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpLatency: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kube_mcp_http_latency_seconds",
				Help:    "Latency of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}
}

// RecordToolCall records a tool call metric.
func (m *Metrics) RecordToolCall(tool, context string, success bool, latencySeconds float64) {
	successLabel := "false"
	if success {
		successLabel = "true"
	}
	m.toolCallsTotal.WithLabelValues(tool, context, successLabel).Inc()
	m.toolLatency.WithLabelValues(tool, context).Observe(latencySeconds)
}

// RecordHTTPRequest records an HTTP request metric.
func (m *Metrics) RecordHTTPRequest(method, path string, statusCode int, latencySeconds float64) {
	statusLabel := fmt.Sprintf("%d", statusCode)
	m.httpRequestsTotal.WithLabelValues(method, path, statusLabel).Inc()
	m.httpLatency.WithLabelValues(method, path).Observe(latencySeconds)
}
