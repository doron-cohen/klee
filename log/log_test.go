package log_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	kleelog "github.com/doron-cohen/klee/log"
	"github.com/stretchr/testify/require"
)

// captureSetup runs Setup with console directed to a buffer instead of os.Stderr.
// It temporarily replaces os.Stderr to capture output.
func captureSetup(t *testing.T, cfg kleelog.Config, opts kleelog.SetupOptions) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stderr
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = orig })

	ctx := context.Background()
	_, logger, setupErr := kleelog.Setup(ctx, cfg, opts)
	if setupErr != nil {
		w.Close()
		return "", setupErr
	}
	logger.Info("test message")

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	return buf.String(), nil
}

func TestSetupNoSinks(t *testing.T) {
	cfg := kleelog.Config{
		Console: kleelog.ConsoleConfig{Enabled: false},
		File:    kleelog.FileConfig{Enabled: false},
	}
	ctx := context.Background()
	_, logger, err := kleelog.Setup(ctx, cfg, kleelog.SetupOptions{})
	require.NoError(t, err)
	require.NotNil(t, logger)
	// Should not panic — writes go to discard
	logger.Info("discarded")
}

func TestSetupConsoleSink(t *testing.T) {
	cfg := kleelog.Config{
		Console: kleelog.ConsoleConfig{Enabled: true, Level: "info", Format: "json"},
		File:    kleelog.FileConfig{Enabled: false},
	}
	out, err := captureSetup(t, cfg, kleelog.SetupOptions{NoColor: true})
	require.NoError(t, err)
	require.Contains(t, out, "test message")
}

func TestSetupFileSink(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	cfg := kleelog.Config{
		Console: kleelog.ConsoleConfig{Enabled: false},
		File: kleelog.FileConfig{
			Enabled: true,
			Path:    path,
			Level:   "debug",
			Format:  "json",
		},
	}
	ctx := context.Background()
	_, logger, err := kleelog.Setup(ctx, cfg, kleelog.SetupOptions{})
	require.NoError(t, err)
	logger.Info("file test")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), "file test")
}

func TestSetupBothSinks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")

	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stderr
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = orig })

	cfg := kleelog.Config{
		Console: kleelog.ConsoleConfig{Enabled: true, Level: "debug", Format: "json"},
		File:    kleelog.FileConfig{Enabled: true, Path: path, Level: "debug", Format: "json"},
	}
	ctx := context.Background()
	_, logger, err := kleelog.Setup(ctx, cfg, kleelog.SetupOptions{NoColor: true})
	require.NoError(t, err)
	logger.Info("both sinks")

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	os.Stderr = orig

	require.Contains(t, buf.String(), "both sinks")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), "both sinks")
}

func TestSetupSetsDefault(t *testing.T) {
	cfg := kleelog.Config{
		Console: kleelog.ConsoleConfig{Enabled: true, Level: "info", Format: "json"},
	}
	ctx := context.Background()
	_, logger, err := kleelog.Setup(ctx, cfg, kleelog.SetupOptions{NoColor: true})
	require.NoError(t, err)
	require.Same(t, logger.Handler(), slog.Default().Handler())
}

func TestFromCtxFallback(t *testing.T) {
	ctx := context.Background()
	logger := kleelog.FromCtx(ctx)
	require.NotNil(t, logger)
	require.Same(t, slog.Default(), logger)
}

func TestFromCtxStored(t *testing.T) {
	ctx := context.Background()
	custom := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx = kleelog.WithCtx(ctx, custom)
	require.Same(t, custom, kleelog.FromCtx(ctx))
}

func TestInvalidLevel(t *testing.T) {
	cfg := kleelog.Config{
		Console: kleelog.ConsoleConfig{Enabled: true, Level: "garbage"},
	}
	ctx := context.Background()
	_, _, err := kleelog.Setup(ctx, cfg, kleelog.SetupOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "garbage")
}

func TestFilePathDefaultsToXDG(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	cfg := kleelog.Config{
		Console: kleelog.ConsoleConfig{Enabled: false},
		File:    kleelog.FileConfig{Enabled: true, Level: "debug", Format: "json"},
	}
	ctx := context.Background()
	_, logger, err := kleelog.Setup(ctx, cfg, kleelog.SetupOptions{AppName: "testapp"})
	require.NoError(t, err)
	logger.Info("xdg path test")
	// If this didn't panic, the path was resolved and the file was opened
}
