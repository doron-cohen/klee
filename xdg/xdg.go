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

// ConfigHome returns $XDG_CONFIG_HOME/<appName>.
func (d Dirs) ConfigHome() string {
	return filepath.Join(xdg.ConfigHome, d.appName)
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
