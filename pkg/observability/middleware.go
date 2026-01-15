package observability

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RecoverPanic recovers from panics in MCP tool handlers.
func RecoverPanic(logger *Logger, toolName string) func() {
	return func() {
		if r := recover(); r != nil {
			logger.Error(context.Background(), "Panic in tool handler",
				"tool", toolName,
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}
}

// RecoverHTTPPanic recovers from panics in HTTP handlers.
func RecoverHTTPPanic(logger *Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				logger.Error(r.Context(), "Panic in HTTP handler",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", p,
					"stack", string(debug.Stack()),
				)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ToolWrapper wraps an MCP tool handler with observability.
func ToolWrapper(
	logger *Logger,
	metrics *Metrics,
	toolName string,
	handler func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error),
) func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
		start := time.Now()
		cluster := "default" // TODO: Extract from request context

		// Recover from panics
		defer RecoverPanic(logger, toolName)()

		// Call the actual handler
		result, out, err := handler(ctx, req, args)

		// Calculate duration
		duration := time.Since(start)

		// Log the invocation
		logger.LogToolInvocation(ctx, toolName, cluster, duration, err)

		// Record metrics
		success := err == nil && (result == nil || !result.IsError)
		metrics.RecordToolCall(toolName, cluster, success, duration.Seconds())

		// If there was an error, wrap it in a normalized error response
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf(`{"error":{"type":"ToolError","message":"%s","details":"%s","cluster":"%s","tool":"%s"}}`,
							err.Error(), err.Error(), cluster, toolName),
					},
				},
			}, nil, nil
		}

		// If result indicates an error, normalize it
		if result != nil && result.IsError {
			// Convert to JSON text content
			errorJSON := fmt.Sprintf(`{"error":{"type":"KubernetesError","message":"%s","cluster":"%s","tool":"%s"}}`,
				extractErrorMessage(result), cluster, toolName)
			result.Content = []mcp.Content{
				&mcp.TextContent{Text: errorJSON},
			}
		}

		return result, out, nil
	}
}

// extractErrorMessage extracts error message from a CallToolResult.
func extractErrorMessage(result *mcp.CallToolResult) string {
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			return textContent.Text
		}
	}
	return "Unknown error"
}

// extractErrorDetails extracts error details from a CallToolResult.
func extractErrorDetails(result *mcp.CallToolResult) string {
	// Try to extract from structured content if available
	if result.StructuredContent != nil {
		return fmt.Sprintf("%v", result.StructuredContent)
	}
	return extractErrorMessage(result)
}

// HTTPMiddleware provides HTTP middleware with observability.
func HTTPMiddleware(logger *Logger, metrics *Metrics, next http.Handler) http.Handler {
	return RecoverHTTPPanic(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log the request
		logger.LogHTTPRequest(r.Context(), r.Method, r.URL.Path, rw.statusCode, duration)

		// Record metrics
		metrics.RecordHTTPRequest(r.Method, r.URL.Path, rw.statusCode, duration.Seconds())
	}))
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
