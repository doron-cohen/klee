package xdg

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

// Dirs holds XDG-compliant paths scoped to an app name.
type Dirs struct {
	appName string
}

// New returns Dirs scoped to appName.
func New(appName string) Dirs {
	return Dirs{appName: appName}
}

// Reload re-resolves the base directories from the environment. They are
// resolved once at process start, so anything that changes XDG_CONFIG_HOME
// or XDG_CONFIG_DIRS afterwards — a test, mostly — has no effect until this
// is called.
func Reload() {
	xdg.Reload()
}

// ConfigHome returns $XDG_CONFIG_HOME/<appName>.
func (d Dirs) ConfigHome() string {
	return filepath.Join(xdg.ConfigHome, d.appName)
}

// ConfigDirs returns the secondary config directories scoped to appName,
// in descending order of preference. These are $XDG_CONFIG_DIRS, or the
// platform defaults when it is unset. ConfigHome is not among them; it
// always outranks every entry here.
//
// The set is platform-specific and on darwin it includes ~/.config, which
// is where users reasonably expect to put config even though the XDG
// config home resolves to ~/Library/Application Support there.
func (d Dirs) ConfigDirs() []string {
	dirs := make([]string, 0, len(xdg.ConfigDirs))
	for _, dir := range xdg.ConfigDirs {
		dirs = append(dirs, filepath.Join(dir, d.appName))
	}
	return dirs
}

// DataHome returns $XDG_DATA_HOME/<appName>.
func (d Dirs) DataHome() string {
	return filepath.Join(xdg.DataHome, d.appName)
}

// CacheHome returns $XDG_CACHE_HOME/<appName>.
func (d Dirs) CacheHome() string {
	return filepath.Join(xdg.CacheHome, d.appName)
}

// RuntimeDir returns $XDG_RUNTIME_DIR/<appName>.
func (d Dirs) RuntimeDir() string {
	return filepath.Join(xdg.RuntimeDir, d.appName)
}

// ConfigFile returns the full path to a config file in ConfigHome.
func (d Dirs) ConfigFile(filename string) string {
	return filepath.Join(d.ConfigHome(), filename)
}
