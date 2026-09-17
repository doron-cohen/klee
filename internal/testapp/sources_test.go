package testapp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	adrgxdg "github.com/adrg/xdg"
	"github.com/doron-cohen/klee"
	"github.com/doron-cohen/klee/internal/testapp"
	"github.com/doron-cohen/klee/kleetest"
	"github.com/stretchr/testify/require"
)

// TestMain keeps the whole package off the developer's real config files:
// the default search path now reaches ~/.config, and a stray config.yaml
// there must not be able to change a test result.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "klee-testapp-home")
	if err != nil {
		panic(err)
	}
	for k, v := range map[string]string{"HOME": home, "XDG_CONFIG_HOME": "", "XDG_CONFIG_DIRS": ""} {
		if err := os.Setenv(k, v); err != nil {
			panic(err)
		}
	}
	adrgxdg.Reload()

	code := m.Run()
	_ = os.RemoveAll(home)
	os.Exit(code)
}

// fakeHome points HOME at a temp dir and re-resolves the XDG base
// directories from it. adrg/xdg caches them at init, so every test that
// changes the environment has to reload afterwards.
func fakeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	// Registered before the Setenv calls so it runs after their restores:
	// a reload against a deleted temp dir would leak into the next test.
	t.Cleanup(adrgxdg.Reload)
	t.Setenv("HOME", home)
	// adrg/xdg treats an empty value as unset, so this is how a test drops
	// an inherited override without disturbing the real environment.
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CONFIG_DIRS", "")
	adrgxdg.Reload()
	return home
}

// runConfigHomeOnly builds an app that opts out of the XDG config dirs.
// The default case is covered by run, which passes no options at all.
func runConfigHomeOnly(t *testing.T, args ...string) *kleetest.Result {
	t.Helper()
	app := testapp.NewApp()
	require.NoError(t, app.LoadConfig(klee.ConfigOptions[testapp.Config]{
		FlagArgs:          append([]string{"app"}, args...),
		DisableConfigDirs: true,
	}))
	return kleetest.Run(t, app, args...)
}

func TestConfigPrintNamesSourcesWithNoFile(t *testing.T) {
	home := fakeHome(t)

	result := run(t, "config", "print")
	result.ExitCode.Equals(t, 0)

	result.Stderr.Contains(t, "config sources")
	result.Stderr.Contains(t, "not found  /etc/testapp/config.yaml")
	result.Stderr.Contains(t, "not found  "+filepath.Join(home, ".config", "testapp", "config.yaml"))
	result.Stdout.Contains(t, "host: localhost")
}

func TestConfigPrintNamesTheFileItRead(t *testing.T) {
	home := fakeHome(t)
	dotConfig := filepath.Join(home, ".config", "testapp", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(dotConfig), 0o755))
	require.NoError(t, os.WriteFile(dotConfig, []byte("host: dotconfig\n"), 0o644))

	result := run(t, "config", "print")
	result.ExitCode.Equals(t, 0)

	result.Stderr.Contains(t, "loaded     "+dotConfig)
	result.Stdout.Contains(t, "host: dotconfig")
}

func TestConfigPrintStdoutStaysMachineReadable(t *testing.T) {
	fakeHome(t)

	result := run(t, "--json", "config", "print")
	result.ExitCode.Equals(t, 0)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Stdout.String()), &parsed),
		"sources must not leak into stdout")
	require.Equal(t, "localhost", parsed["host"])
	result.Stderr.Contains(t, "config sources")
}

func TestConfigPrintQuietOmitsSources(t *testing.T) {
	fakeHome(t)

	result := run(t, "--quiet", "config", "print")
	result.ExitCode.Equals(t, 0)
	result.Stdout.Contains(t, "host: localhost")
	require.NotContains(t, result.Stderr.String(), "config sources")
}

// An app that opts out still gets its narrower search path named, rather
// than silently omitting the directories it chose not to look at.
func TestConfigPrintSourcesWithConfigDirsDisabled(t *testing.T) {
	home := fakeHome(t)

	result := runConfigHomeOnly(t, "config", "print")
	result.ExitCode.Equals(t, 0)

	result.Stderr.Contains(t, "config sources")
	result.Stderr.Contains(t, filepath.Join(adrgxdg.ConfigHome, "testapp", "config.yaml"))
	if runtime.GOOS == "darwin" {
		require.NotContains(t, result.Stderr.String(),
			filepath.Join(home, ".config", "testapp", "config.yaml"))
	}
}
