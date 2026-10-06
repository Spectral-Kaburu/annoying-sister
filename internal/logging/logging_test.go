package logging

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"  debug  ", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFor(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	baseLogger := slog.New(handler)

	compLogger := For(baseLogger, ComponentIdle)
	compLogger.Info("test message")

	output := buf.String()
	require.Contains(t, output, "component=idle_watcher")
	require.Contains(t, output, `msg="test message"`)
}

func TestInit(t *testing.T) {
	t.Setenv("NUDGE_LOG_LEVEL", "debug")
	t.Setenv("NUDGE_LOG_FORMAT", "text")

	logger := Init()
	require.NotNil(t, logger)
}
