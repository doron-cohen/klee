package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	adrgxdg "github.com/adrg/xdg"
	"github.com/doron-cohen/klee/config"
	"github.com/stretchr/testify/require"
)

// fakeHome points HOME at a temp dir and re-resolves the XDG base
// directories from it. adrg/xdg caches them at init, so every test that
// touches the search path has to reload after changing the environment.
func fakeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	// adrg/xdg treats an empty value as unset, so this is how a test drops
	// an inherited override without disturbing the real environment.
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CONFIG_DIRS", "")
	adrgxdg.Reload()
	t.Cleanup(adrgxdg.Reload)
	return home
}

func writeConfig(t *testing.T, path, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
}

// noProjectFile keeps the default ./<app>.yaml out of the way.
func noProjectFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "absent.yaml")
}

func TestDotConfigIsSearched(t *testing.T) {
	home := fakeHome(t)
	writeConfig(t, filepath.Join(home, ".config", "kleetest", "config.yaml"), "host: dotconfig\n")

	var cfg testConfig
	require.NoError(t, config.Load(&cfg, config.Options{
		AppName:          "kleetest",
		ProjectPath:      noProjectFile(t),
		SearchConfigDirs: true,
	}))

	require.Equal(t, "dotconfig", cfg.Host)
}

func TestConfigDirsAreNotSearchedByDefault(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("config home and ~/.config only differ on darwin")
	}
	home := fakeHome(t)
	writeConfig(t, filepath.Join(home, ".config", "kleetest", "config.yaml"), "host: dotconfig\n")

	var cfg testConfig
	require.NoError(t, config.Load(&cfg, config.Options{
		AppName:     "kleetest",
		ProjectPath: noProjectFile(t),
	}))

	require.Equal(t, "localhost", cfg.Host, "staying opted out must keep v0.2.2 behaviour")
}

// The secondary dirs are platform-specific but XDG_CONFIG_DIRS drives them
// everywhere, so this pins the precedence rule on darwin and Linux alike.
func TestConfigHomeWinsOverConfigDirs(t *testing.T) {
	fakeHome(t)
	secondary, primary := t.TempDir(), t.TempDir()
	t.Setenv("XDG_CONFIG_DIRS", secondary)
	t.Setenv("XDG_CONFIG_HOME", primary)
	adrgxdg.Reload()

	writeConfig(t, filepath.Join(secondary, "kleetest", "config.yaml"), "host: secondary\nport: 1111\n")
	writeConfig(t, filepath.Join(primary, "kleetest", "config.yaml"), "host: primary\n")

	var cfg testConfig
	require.NoError(t, config.Load(&cfg, config.Options{
		AppName:          "kleetest",
		ProjectPath:      noProjectFile(t),
		SearchConfigDirs: true,
	}))

	require.Equal(t, "primary", cfg.Host)
	require.Equal(t, 1111, cfg.Port, "keys the config home file omits still come from the secondary dir")
}

func TestConfigHomeWinsOverDotConfig(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("config home and ~/.config only differ on darwin")
	}
	home := fakeHome(t)
	writeConfig(t, filepath.Join(home, ".config", "kleetest", "config.yaml"), "host: dotconfig\nport: 1111\n")
	writeConfig(t, filepath.Join(home, "Library", "Application Support", "kleetest", "config.yaml"), "host: apphome\n")

	var cfg testConfig
	require.NoError(t, config.Load(&cfg, config.Options{
		AppName:          "kleetest",
		ProjectPath:      noProjectFile(t),
		SearchConfigDirs: true,
	}))

	require.Equal(t, "apphome", cfg.Host)
	require.Equal(t, 1111, cfg.Port, "keys the config home file omits still come from ~/.config")
}
