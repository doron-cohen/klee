package xdg_test

import (
	"path/filepath"
	"runtime"
	"testing"

	adrgxdg "github.com/adrg/xdg"
	"github.com/doron-cohen/klee/xdg"
	"github.com/stretchr/testify/require"
)

// fakeHome points HOME at a temp dir and re-resolves the XDG base
// directories from it. adrg/xdg caches them at init, so every test that
// changes the environment has to reload afterwards.
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

func TestConfigDirsAreAppScoped(t *testing.T) {
	fakeHome(t)
	for _, dir := range xdg.New("myapp").ConfigDirs() {
		require.Equal(t, "myapp", filepath.Base(dir), "unexpected dir: %s", dir)
	}
}

func TestConfigDirsFollowXDGConfigDirs(t *testing.T) {
	fakeHome(t)
	t.Setenv("XDG_CONFIG_DIRS", "/one:/two")
	adrgxdg.Reload()

	require.Equal(t,
		[]string{"/one/myapp", "/two/myapp"},
		xdg.New("myapp").ConfigDirs())
}

func TestConfigDirsIncludeDotConfigOnDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("~/.config is only a secondary dir on darwin")
	}
	home := fakeHome(t)

	require.Contains(t,
		xdg.New("myapp").ConfigDirs(),
		filepath.Join(home, ".config", "myapp"))
}

func TestConfigDirsExcludeConfigHome(t *testing.T) {
	fakeHome(t)
	dirs := xdg.New("myapp")
	require.NotContains(t, dirs.ConfigDirs(), dirs.ConfigHome())
}
