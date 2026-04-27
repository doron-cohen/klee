package log_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	internallog "github.com/doron-cohen/klee/internal/log"
	"github.com/stretchr/testify/require"
)

func TestFileHandlerWritesJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	h := internallog.NewFileHandler(internallog.FileOptions{
		Path:       path,
		Level:      slog.LevelDebug,
		Format:     "json",
		MaxSizeMB:  100,
		MaxBackups: 3,
		MaxAgeDays: 28,
	})
	logger := slog.New(h)
	logger.Info("file log message", "key", "val")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var record map[string]any
	require.NoError(t, json.Unmarshal(data, &record))
	require.Equal(t, "INFO", record["level"])
	require.Equal(t, "file log message", record["msg"])
	require.Equal(t, "val", record["key"])
}

func TestFileHandlerLevelFiltering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	h := internallog.NewFileHandler(internallog.FileOptions{
		Path:   path,
		Level:  slog.LevelWarn,
		Format: "json",
	})
	logger := slog.New(h)
	logger.Info("filtered")
	logger.Warn("kept")

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	require.NotContains(t, content, "filtered")
	require.Contains(t, content, "kept")
}

func TestFileHandlerTextFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")
	h := internallog.NewFileHandler(internallog.FileOptions{
		Path:   path,
		Level:  slog.LevelDebug,
		Format: "text",
	})
	logger := slog.New(h)
	logger.Info("text format test")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), "text format test")
}
