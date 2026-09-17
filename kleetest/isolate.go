package kleetest

import (
	"path/filepath"
	"testing"

	"github.com/doron-cohen/klee/xdg"
)

// IsolateConfig points the whole XDG config search at a fresh temp directory
// for the duration of the test, and returns it. Write a file under
// <dir>/<appName>/config.yaml to have the app read it.
//
// Setting XDG_CONFIG_HOME by hand is not enough: it leaves the secondary
// config directories pointing at the real ones — on darwin that includes
// ~/.config — and the base directories are resolved once at process start,
// so a bare t.Setenv has no effect at all.
func IsolateConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	// Registered before the Setenv calls so it runs after their restores.
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_CONFIG_DIRS", filepath.Join(dir, "dirs"))
	xdg.Reload()

	return dir
}
