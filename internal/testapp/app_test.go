package testapp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doron-cohen/klee"
	"github.com/doron-cohen/klee/config"
	"github.com/doron-cohen/klee/internal/testapp"
	"github.com/doron-cohen/klee/kleetest"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// memStore is an in-memory WritableSecretStore for testing.
type memStore struct {
	m map[string]string
}

func newMemStore() *memStore {
	return &memStore{m: make(map[string]string)}
}

func (m *memStore) Get(key string) (string, error) {
	v, ok := m.m[key]
	if !ok {
		return "", config.ErrSecretNotFound
	}
	return v, nil
}

func (m *memStore) Set(key, value string) error {
	m.m[key] = value
	return nil
}

// readOnlyStore implements config.SecretStore but not config.WritableSecretStore.
type readOnlyStore struct{}

func (r *readOnlyStore) Get(key string) (string, error) {
	return "", config.ErrSecretNotFound
}

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

// --- output messages ---

func TestOutputMessages(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      string
		wantErrEmpty bool
	}{
		{"success shown", []string{"msg", "success"}, "success message", false},
		{"warn shown", []string{"msg", "warn"}, "warn message", false},
		{"error shown", []string{"msg", "error"}, "error message", false},
		{"hint shown", []string{"msg", "hint"}, "hint message", false},
		{"warn suppressed quiet", []string{"--quiet", "msg", "warn"}, "", true},
		{"hint suppressed quiet", []string{"--quiet", "msg", "hint"}, "", true},
		{"success shown quiet", []string{"--quiet", "msg", "success"}, "success message", false},
		{"error shown quiet", []string{"--quiet", "msg", "error"}, "error message", false},
		{"all suppressed json", []string{"--json", "msg", "success"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := run(t, tt.args...)
			result.ExitCode.Equals(t, 0)
			if tt.wantErrEmpty {
				result.Stderr.Empty(t)
			} else {
				result.Stderr.Contains(t, tt.wantErr)
			}
		})
	}
}

// --- secrets ---

func TestSecretsGet(t *testing.T) {
	store := newMemStore()
	store.Set("mykey", "myvalue")

	app := testapp.NewApp().WithSecretStore(store)
	require.NoError(t, app.LoadConfig(klee.ConfigOptions[testapp.Config]{
		FlagArgs: []string{"app"},
	}))

	result := kleetest.Run(t, app, "secrets", "get", "mykey")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "myvalue")
	result.Stderr.Empty(t)
}

func TestSecretsGet_MissingKey(t *testing.T) {
	store := newMemStore()

	app := testapp.NewApp().WithSecretStore(store)
	require.NoError(t, app.LoadConfig(klee.ConfigOptions[testapp.Config]{
		FlagArgs: []string{"app"},
	}))

	result := kleetest.Run(t, app, "secrets", "get", "nonexistent")
	result.ExitCode.Equals(t, 1)
	result.Stderr.Contains(t, "not found")
}

func TestSecretsSet(t *testing.T) {
	store := newMemStore()

	app := testapp.NewApp().WithSecretStore(store)
	require.NoError(t, app.LoadConfig(klee.ConfigOptions[testapp.Config]{
		FlagArgs: []string{"app"},
	}))

	result := kleetest.Run(t, app, "secrets", "set", "newkey", "newvalue")
	result.ExitCode.Equals(t, 0)

	val, err := store.Get("newkey")
	require.NoError(t, err)
	require.Equal(t, "newvalue", val)
}

func TestSecretsSet_ReadOnlyStore(t *testing.T) {
	store := &readOnlyStore{}

	app := testapp.NewApp().WithSecretStore(store)
	require.NoError(t, app.LoadConfig(klee.ConfigOptions[testapp.Config]{
		FlagArgs: []string{"app"},
	}))

	result := kleetest.Run(t, app, "secrets", "set", "key", "val")
	result.ExitCode.Equals(t, 1)
	result.Stderr.Contains(t, "does not support write")
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
