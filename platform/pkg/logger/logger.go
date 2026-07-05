// Package logger provides a shared JSON structured logger built on log/slog.
// Every service tags its logs with a "service" attribute and honours the
// LOG_LEVEL env var (debug|info|warn|error, default info).
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// New returns a JSON logger writing to stdout, tagged with the service name.
func New(service string) *slog.Logger {
	return NewWithWriter(os.Stdout, service)
}

// NewWithWriter is like New but writes to w. Useful for tests.
func NewWithWriter(w io.Writer, service string) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: levelFromEnv()})
	return slog.New(handler).With("service", service)
}

func levelFromEnv() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
