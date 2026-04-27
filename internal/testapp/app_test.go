package testapp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doron-cohen/klee"
	"github.com/doron-cohen/klee/internal/testapp"
	"github.com/doron-cohen/klee/kleetest"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// run creates a fresh app, loads config from args, and runs with kleetest.
func run(t *testing.T, args ...string) *kleetest.Result {
	t.Helper()
	app := testapp.NewApp()
	require.NoError(t, app.LoadConfig(klee.ConfigOptions[testapp.Config]{
		FlagArgs: append([]string{"app"}, args...),
	}))
	return kleetest.Run(t, app, args...)
}

// runWithFile writes a YAML config to a temp file and runs with --config.
func runWithFile(t *testing.T, yaml string, args ...string) *kleetest.Result {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o644))
	return run(t, append([]string{"--config", path}, args...)...)
}

// --- version ---

func TestVersion(t *testing.T) {
	result := run(t, "version")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "1.0.0-test")
	result.Stderr.Empty(t)
}

func TestVersionShort(t *testing.T) {
	result := run(t, "version", "--short")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "1.0.0")
	result.Stderr.Empty(t)
}

// --- config loading ---

func TestConfigDefaults(t *testing.T) {
	result := run(t, "echo")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "host=localhost")
	result.Stdout.Contains(t, "port=8080")
}

func TestConfigFromFile(t *testing.T) {
	result := runWithFile(t, "host: myserver\nport: 9090\n", "echo")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "host=myserver")
	result.Stdout.Contains(t, "port=9090")
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("TESTAPP_HOST", "envhost")
	result := run(t, "echo")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "host=envhost")
}

func TestConfigEnvOverridesFile(t *testing.T) {
	t.Setenv("TESTAPP_HOST", "envhost")
	result := runWithFile(t, "host: filehost\n", "echo")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "host=envhost")
}

func TestConfigFilePartialOverride(t *testing.T) {
	result := runWithFile(t, "host: myserver\n", "echo")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "host=myserver")
	result.Stdout.Contains(t, "port=8080") // default still applied
}

// --- AfterLoad ---

func TestAfterLoad(t *testing.T) {
	app := testapp.NewApp()
	afterLoadCalled := false

	require.NoError(t, app.LoadConfig(klee.ConfigOptions[testapp.Config]{
		FlagArgs: []string{"app"},
		AfterLoad: func(cfg *testapp.Config, _ *cli.Command) error {
			afterLoadCalled = true
			cfg.Host = "from-after-load"
			return nil
		},
	}))

	result := kleetest.Run(t, app, "echo")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "host=from-after-load")
	require.True(t, afterLoadCalled)
}

// --- run flags ---

func TestRunFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{"no flags", []string{"flags"}, "debug=false quiet=false json=false no-color=false"},
		{"debug", []string{"--debug", "flags"}, "debug=true"},
		{"quiet", []string{"--quiet", "flags"}, "quiet=true"},
		{"json", []string{"--json", "flags"}, "json=true"},
		{"no-color", []string{"--no-color", "flags"}, "no-color=true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := run(t, tt.args...)
			result.ExitCode.Equals(t, 0)
			result.Stdout.Contains(t, tt.wantOut)
		})
	}
}

// --- logging ---

func TestLogCommandWritesToStderr(t *testing.T) {
	result := run(t, "log")
	result.ExitCode.Equals(t, 0)
	result.Stderr.Contains(t, "log command ran")
}

func TestLogLevelFlagFiltersOutput(t *testing.T) {
	// Default level is info, so info messages appear
	result := run(t, "log")
	result.ExitCode.Equals(t, 0)
	result.Stderr.Contains(t, "log command ran")
}

func TestLogFormatFlagJSON(t *testing.T) {
	result := run(t, "--log-format", "json", "log")
	result.ExitCode.Equals(t, 0)
	result.Stderr.Contains(t, `"msg":"log command ran"`)
}

// --- error rendering ---

func TestErrors(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantErr  string
	}{
		{"user error message", []string{"fail", "user"}, 2, "bad input"},
		{"user error hint", []string{"fail", "user"}, 2, "check your input"},
		{"internal error message", []string{"fail", "internal"}, 3, "unexpected failure"},
		{"internal error debug hint", []string{"fail", "internal"}, 3, "--debug"},
		{"plain error", []string{"fail", "plain"}, 1, "plain error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := run(t, tt.args...)
			result.ExitCode.Equals(t, tt.wantCode)
			result.Stderr.Contains(t, tt.wantErr)
		})
	}
}

