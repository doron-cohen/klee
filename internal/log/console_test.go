package log_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	internallog "github.com/doron-cohen/klee/internal/log"
	"github.com/stretchr/testify/require"
)

func TestConsoleHandlerJSON(t *testing.T) {
	var buf bytes.Buffer
	h := internallog.NewConsoleHandler(&buf, slog.LevelDebug, "json", true)
	logger := slog.New(h)
	logger.Info("hello", "key", "value")

	var record map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &record))
	require.Equal(t, "INFO", record["level"])
	require.Equal(t, "hello", record["msg"])
	require.Equal(t, "value", record["key"])
}

func TestConsoleHandlerPrettyNoColor(t *testing.T) {
	var buf bytes.Buffer
	h := internallog.NewConsoleHandler(&buf, slog.LevelDebug, "pretty", true)
	logger := slog.New(h)
	logger.Info("hello world")

	out := buf.String()
	require.Contains(t, out, "hello world")
	// no ANSI escape codes when noColor=true
	require.NotContains(t, out, "\x1b[")
}

func TestConsoleHandlerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	h := internallog.NewConsoleHandler(&buf, slog.LevelWarn, "json", true)
	logger := slog.New(h)

	logger.Info("should be filtered")
	logger.Warn("should appear")

	require.NotContains(t, buf.String(), "should be filtered")
	require.Contains(t, buf.String(), "should appear")
}

func TestConsoleHandlerUnknownFormatFallsToPretty(t *testing.T) {
	var buf bytes.Buffer
	h := internallog.NewConsoleHandler(&buf, slog.LevelDebug, "unknown", true)
	logger := slog.New(h)
	logger.Info("test message")

	out := buf.String()
	require.Contains(t, out, "test message")
	require.False(t, strings.HasPrefix(strings.TrimSpace(out), "{"), "expected pretty output, got JSON")
}
