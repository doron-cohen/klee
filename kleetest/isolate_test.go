package kleetest_test

import (
	"strings"
	"testing"

	"github.com/doron-cohen/klee/kleetest"
	"github.com/doron-cohen/klee/xdg"
	"github.com/stretchr/testify/require"
)

func TestIsolateConfigRedirectsTheWholeSearch(t *testing.T) {
	dir := kleetest.IsolateConfig(t)

	dirs := xdg.New("myapp")
	searched := append([]string{dirs.ConfigHome()}, dirs.ConfigDirs()...)

	for _, path := range searched {
		require.True(t, strings.HasPrefix(path, dir),
			"%s is outside the isolated directory %s", path, dir)
	}
}

func TestIsolateConfigIsUndoneAfterTheTest(t *testing.T) {
	before := xdg.New("myapp").ConfigHome()

	t.Run("isolated", func(t *testing.T) {
		kleetest.IsolateConfig(t)
		require.NotEqual(t, before, xdg.New("myapp").ConfigHome())
	})

	require.Equal(t, before, xdg.New("myapp").ConfigHome())
}
