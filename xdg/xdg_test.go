package xdg_test

import (
	"strings"
	"testing"

	"github.com/doron-cohen/klee/xdg"
	"github.com/stretchr/testify/require"
)

func TestDirsContainAppName(t *testing.T) {
	const appName = "myapp"
	dirs := xdg.New(appName)

	tests := []struct {
		name string
		path string
	}{
		{"ConfigHome", dirs.ConfigHome()},
		{"DataHome", dirs.DataHome()},
		{"CacheHome", dirs.CacheHome()},
		{"RuntimeDir", dirs.RuntimeDir()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, strings.Contains(tt.path, appName),
				"expected path to contain %q, got: %s", appName, tt.path)
		})
	}
}

func TestConfigFile(t *testing.T) {
	dirs := xdg.New("myapp")
	path := dirs.ConfigFile("config.yaml")
	require.True(t, strings.HasSuffix(path, "myapp/config.yaml"),
		"unexpected path: %s", path)
}

func TestDifferentAppNames(t *testing.T) {
	a := xdg.New("app-a")
	b := xdg.New("app-b")
	require.NotEqual(t, a.ConfigHome(), b.ConfigHome())
	require.NotEqual(t, a.DataHome(), b.DataHome())
}
