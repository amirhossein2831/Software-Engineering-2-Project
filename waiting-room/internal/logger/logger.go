package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

func New(service string) *slog.Logger {
	return NewWithWriter(os.Stdout, service)
}

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
