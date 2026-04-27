package config_test

import (
	"os"
	"path/filepath"
	"testing"

	internal "github.com/doron-cohen/klee/internal/config"
	"github.com/stretchr/testify/require"
)

type mergeConfig struct {
	Host  string  `yaml:"host"  env:"KLEE_MERGE_HOST"  default:"localhost"`
	Port  int     `yaml:"port"  env:"KLEE_MERGE_PORT"  default:"8080"`
	Debug bool    `yaml:"debug" env:"KLEE_MERGE_DEBUG" default:"false"`
	Rate  float64 `yaml:"rate"  env:"KLEE_MERGE_RATE"  default:"1.5"`
}

func writeYAML(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestMergePrecedence(t *testing.T) {
	tests := []struct {
		name      string
		files     []string // YAML content for each file, in order
		env       map[string]string
		wantHost  string
		wantPort  int
		wantDebug bool
	}{
		{
			name:      "defaults only",
			wantHost:  "localhost",
			wantPort:  8080,
			wantDebug: false,
		},
		{
			name:     "single file overrides defaults",
			files:    []string{"host: filehost\nport: 9090\n"},
			wantHost: "filehost",
			wantPort: 9090,
		},
		{
			name:     "later file wins over earlier",
			files:    []string{"host: first\n", "host: second\n"},
			wantHost: "second",
			wantPort: 8080,
		},
		{
			name:     "file only sets present fields, others keep defaults",
			files:    []string{"host: onlyhost\n"},
			wantHost: "onlyhost",
			wantPort: 8080,
		},
		{
			name:     "env overrides file",
			files:    []string{"host: fromfile\n"},
			env:      map[string]string{"KLEE_MERGE_HOST": "fromenv"},
			wantHost: "fromenv",
			wantPort: 8080,
		},
		{
			name:     "env overrides defaults with no file",
			env:      map[string]string{"KLEE_MERGE_PORT": "7777"},
			wantHost: "localhost",
			wantPort: 7777,
		},
		{
			name:      "env overrides both files",
			files:     []string{"host: first\n", "host: second\n"},
			env:       map[string]string{"KLEE_MERGE_HOST": "fromenv"},
			wantHost:  "fromenv",
			wantPort:  8080,
			wantDebug: false,
		},
		{
			name:     "missing file is silently skipped",
			files:    nil, // no files written, paths will be nonexistent
			wantHost: "localhost",
			wantPort: 8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			var paths []string
			if tt.files != nil {
				for _, content := range tt.files {
					paths = append(paths, writeYAML(t, content))
				}
			} else {
				paths = []string{"/nonexistent/klee-test-merge.yaml"}
			}

			var cfg mergeConfig
			err := internal.Merge(internal.MergeOptions{Paths: paths, Dest: &cfg})
			require.NoError(t, err)

			require.Equal(t, tt.wantHost, cfg.Host)
			require.Equal(t, tt.wantPort, cfg.Port)
			require.Equal(t, tt.wantDebug, cfg.Debug)
		})
	}
}

func TestMergeTypeCoercion(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		wantPort  int
		wantDebug bool
		wantRate  float64
		wantErr   bool
	}{
		{
			name:     "int from env",
			env:      map[string]string{"KLEE_MERGE_PORT": "9999"},
			wantPort: 9999, wantDebug: false, wantRate: 1.5,
		},
		{
			name:      "bool from env",
			env:       map[string]string{"KLEE_MERGE_DEBUG": "true"},
			wantPort:  8080, wantDebug: true, wantRate: 1.5,
		},
		{
			name:     "float from env",
			env:      map[string]string{"KLEE_MERGE_RATE": "3.14"},
			wantPort: 8080, wantDebug: false, wantRate: 3.14,
		},
		{
			name:    "invalid int returns error",
			env:     map[string]string{"KLEE_MERGE_PORT": "notanumber"},
			wantErr: true,
		},
		{
			name:    "invalid bool returns error",
			env:     map[string]string{"KLEE_MERGE_DEBUG": "notabool"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			var cfg mergeConfig
			err := internal.Merge(internal.MergeOptions{
				Paths: []string{"/nonexistent/klee-test-types.yaml"},
				Dest:  &cfg,
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantPort, cfg.Port)
			require.Equal(t, tt.wantDebug, cfg.Debug)
			require.InDelta(t, tt.wantRate, cfg.Rate, 0.001)
		})
	}
}

func TestMergeErrors(t *testing.T) {
	t.Run("invalid YAML returns error", func(t *testing.T) {
		path := writeYAML(t, "host: [invalid yaml")
		var cfg mergeConfig
		err := internal.Merge(internal.MergeOptions{Paths: []string{path}, Dest: &cfg})
		require.Error(t, err)
	})

	t.Run("non-pointer dest returns error", func(t *testing.T) {
		var cfg mergeConfig
		err := internal.Merge(internal.MergeOptions{Dest: cfg})
		require.Error(t, err)
	})

	t.Run("empty file is valid and no-ops", func(t *testing.T) {
		path := writeYAML(t, "")
		var cfg mergeConfig
		err := internal.Merge(internal.MergeOptions{Paths: []string{path}, Dest: &cfg})
		require.NoError(t, err)
		require.Equal(t, "localhost", cfg.Host)
		require.Equal(t, 8080, cfg.Port)
	})
}
