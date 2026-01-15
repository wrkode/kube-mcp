package net

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpHelpers "github.com/wrkode/kube-mcp/pkg/mcp"
)

// handleHubbleFlowsQuery handles the net.hubble_flows_query tool.
func (t *Toolset) handleHubbleFlowsQuery(ctx context.Context, args struct {
	Context      string `json:"context"`
	Namespace    string `json:"namespace,omitempty"`
	Pod          string `json:"pod,omitempty"`
	Verdict      string `json:"verdict,omitempty"`
	Protocol     string `json:"protocol,omitempty"`
	SinceSeconds int    `json:"since_seconds,omitempty"`
	Limit        int    `json:"limit,omitempty"`
}) (*mcp.CallToolResult, error) {
	if !t.hasHubble {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "FeatureDisabled",
				"message": "Hubble API not configured",
				"details": "hubble_api_url must be set in configuration",
			},
		})
		return result, err
	}

	if t.hubbleClient == nil {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "ExternalServiceUnavailable",
				"message": "Hubble client not initialized",
			},
		})
		return result, err
	}

	// Build query URL
	queryURL := fmt.Sprintf("%s/api/v1/flows", t.hubbleAPIURL)
	u, err := url.Parse(queryURL)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("invalid Hubble API URL: %w", err)), nil
	}

	q := u.Query()
	if args.Namespace != "" {
		q.Set("namespace", args.Namespace)
	}
	if args.Pod != "" {
		q.Set("pod", args.Pod)
	}
	if args.Verdict != "" {
		q.Set("verdict", args.Verdict)
	}
	if args.Protocol != "" {
		q.Set("protocol", args.Protocol)
	}
	if args.SinceSeconds > 0 {
		q.Set("since", strconv.Itoa(args.SinceSeconds))
	}
	if args.Limit > 0 {
		q.Set("limit", strconv.Itoa(args.Limit))
	}
	u.RawQuery = q.Encode()

	// Make request
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to create request: %w", err)), nil
	}

	resp, err := t.hubbleClient.Do(req)
	if err != nil {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "ExternalServiceUnavailable",
				"message": fmt.Sprintf("Failed to query Hubble API: %v", err),
			},
		})
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result, err := mcpHelpers.NewJSONResult(map[string]any{
			"error": map[string]any{
				"type":    "ExternalServiceUnavailable",
				"message": fmt.Sprintf("Hubble API returned status %d", resp.StatusCode),
			},
		})
		return result, err
	}

	// Parse response (best-effort, Hubble API format may vary)
	var hubbleResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&hubbleResponse); err != nil {
		return mcpHelpers.NewErrorResult(fmt.Errorf("failed to parse Hubble response: %w", err)), nil
	}

	// Normalize flows (best-effort)
	var flows []HubbleFlow
	if flowsData, ok := hubbleResponse["flows"].([]interface{}); ok {
		for _, flowData := range flowsData {
			if flowMap, ok := flowData.(map[string]interface{}); ok {
				flow := HubbleFlow{}
				if timeStr, ok := flowMap["time"].(string); ok {
					flow.Time = timeStr
				}
				if verdict, ok := flowMap["verdict"].(string); ok {
					flow.Verdict = verdict
				}
				if source, ok := flowMap["source"].(map[string]interface{}); ok {
					flow.Source = &FlowEndpoint{}
					if ns, ok := source["namespace"].(string); ok {
						flow.Source.Namespace = ns
					}
					if pod, ok := source["pod_name"].(string); ok {
						flow.Source.Pod = pod
					}
				}
				if dest, ok := flowMap["destination"].(map[string]interface{}); ok {
					flow.Destination = &FlowEndpoint{}
					if ns, ok := dest["namespace"].(string); ok {
						flow.Destination.Namespace = ns
					}
					if pod, ok := dest["pod_name"].(string); ok {
						flow.Destination.Pod = pod
					}
				}
				if l4, ok := flowMap["l4"].(map[string]interface{}); ok {
					if tcp, ok := l4["TCP"].(map[string]interface{}); ok {
						flow.Protocol = "TCP"
						if port, ok := tcp["destination_port"].(float64); ok {
							p := int32(port)
							flow.Port = &p
						}
					} else if udp, ok := l4["UDP"].(map[string]interface{}); ok {
						flow.Protocol = "UDP"
						if port, ok := udp["destination_port"].(float64); ok {
							p := int32(port)
							flow.Port = &p
						}
					}
				}
				flows = append(flows, flow)
			}
		}
	}

	result, err := mcpHelpers.NewJSONResult(map[string]any{
		"flows": flows,
	})
	return result, err
}
