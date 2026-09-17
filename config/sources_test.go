package config_test

import (
	"path/filepath"
	"testing"

	adrgxdg "github.com/adrg/xdg"
	"github.com/doron-cohen/klee/config"
	"github.com/stretchr/testify/require"
)

func paths(sources []config.Source) []string {
	out := make([]string, 0, len(sources))
	for _, s := range sources {
		out = append(out, s.Path)
	}
	return out
}

func loaded(sources []config.Source) []string {
	out := []string{}
	for _, s := range sources {
		if s.Loaded {
			out = append(out, s.Path)
		}
	}
	return out
}

func TestSourcesListEverySearchPath(t *testing.T) {
	home := fakeHome(t)
	project := noProjectFile(t)

	var cfg testConfig
	sources, err := config.LoadWithSources(&cfg, config.Options{
		AppName:     "kleetest",
		ProjectPath: project,
	})
	require.NoError(t, err)

	got := paths(sources)
	require.Equal(t, "/etc/kleetest/config.yaml", got[0])
	require.Equal(t, project, got[len(got)-1])
	require.Contains(t, got, filepath.Join(home, ".config", "kleetest", "config.yaml"))
	require.Contains(t, got, filepath.Join(adrgxdg.ConfigHome, "kleetest", "config.yaml"))
	require.Empty(t, loaded(sources), "no file exists, so nothing was read")
}

func TestSourcesMarkTheFileThatWasRead(t *testing.T) {
	home := fakeHome(t)
	dotConfig := filepath.Join(home, ".config", "kleetest", "config.yaml")
	writeConfig(t, dotConfig, "host: dotconfig\n")

	var cfg testConfig
	sources, err := config.LoadWithSources(&cfg, config.Options{
		AppName:     "kleetest",
		ProjectPath: noProjectFile(t),
	})
	require.NoError(t, err)

	require.Equal(t, []string{dotConfig}, loaded(sources))
}

// DisableConfigDirs reduces the search to exactly the three paths v0.2.2
// looked at, which is the whole of what an app gives up by setting it.
func TestDisableConfigDirsRestoresTheV022SearchPath(t *testing.T) {
	fakeHome(t)
	project := noProjectFile(t)

	var cfg testConfig
	sources, err := config.LoadWithSources(&cfg, config.Options{
		AppName:           "kleetest",
		ProjectPath:       project,
		DisableConfigDirs: true,
	})
	require.NoError(t, err)

	require.Equal(t, []string{
		"/etc/kleetest/config.yaml",
		filepath.Join(adrgxdg.ConfigHome, "kleetest", "config.yaml"),
		project,
	}, paths(sources))
}
