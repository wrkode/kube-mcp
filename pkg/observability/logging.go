package observability

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// Logger provides structured logging for kube-mcp.
type Logger struct {
	logger *slog.Logger
}

// LogLevel represents a log level.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// NewLogger creates a new logger with the specified level and format.
func NewLogger(level LogLevel, jsonFormat bool) *Logger {
	var logLevel slog.Level
	switch level {
	case LogLevelDebug:
		logLevel = slog.LevelDebug
	case LogLevelInfo:
		logLevel = slog.LevelInfo
	case LogLevelWarn:
		logLevel = slog.LevelWarn
	case LogLevelError:
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if jsonFormat {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// LogToolInvocation logs an MCP tool invocation.
func (l *Logger) LogToolInvocation(ctx context.Context, toolName, cluster string, duration time.Duration, err error) {
	attrs := []any{
		"tool", toolName,
		"cluster", cluster,
		"duration_ms", duration.Milliseconds(),
	}

	if err != nil {
		attrs = append(attrs, "error", err.Error())
		l.logger.ErrorContext(ctx, "Tool invocation failed", attrs...)
	} else {
		l.logger.InfoContext(ctx, "Tool invocation completed", attrs...)
	}
}

// LogHTTPRequest logs an HTTP request.
func (l *Logger) LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	l.logger.InfoContext(ctx, "HTTP request",
		"method", method,
		"path", path,
		"status", statusCode,
		"duration_ms", duration.Milliseconds(),
	)
}

// Debug logs a debug message.
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}
