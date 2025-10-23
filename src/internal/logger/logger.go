package logger

import (
	"io"
	"os"
	"strings"

	"log/slog"
)

// Logger wraps the structured logger
type Logger struct {
	*slog.Logger
}

// New creates a new logger instance
func New(level, format string) *Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewWithWriter creates a new logger instance with a custom writer
func NewWithWriter(w io.Writer, level, format string) *Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: logLevel,
		})
	} else {
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: logLevel,
		})
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{Logger: l.Logger.With(args...)}
}

// WithField creates a new logger with a single additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(key, value)}
}
