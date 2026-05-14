package config_test

import (
	"os"
	"path/filepath"
	"testing"

	internal "github.com/doron-cohen/klee/internal/config"
	"github.com/stretchr/testify/require"
)

type dotenvConfig struct {
	Host  string `yaml:"host"  env:"KLEE_DOTENV_HOST"  default:"localhost"`
	Port  int    `yaml:"port"  env:"KLEE_DOTENV_PORT"  default:"8080"`
	Debug bool   `yaml:"debug" env:"KLEE_DOTENV_DEBUG" default:"false"`
}

func mergeWithDotEnv(t *testing.T, dotenvPaths []string) (dotenvConfig, error) {
	t.Helper()
	var cfg dotenvConfig
	err := internal.Merge(internal.MergeOptions{
		Paths:       []string{"/nonexistent/klee-dotenv-test.yaml"},
		DotEnvFiles: dotenvPaths,
		Dest:        &cfg,
	})
	return cfg, err
}

func TestDotEnvParsing(t *testing.T) {
	t.Run("valid file sets fields", func(t *testing.T) {
		cfg, err := mergeWithDotEnv(t, []string{"testdata/valid.env"})
		require.NoError(t, err)
		require.Equal(t, "dotenvhost", cfg.Host)
		require.Equal(t, 7777, cfg.Port)
		require.Equal(t, true, cfg.Debug)
	})

	t.Run("comments and blank lines are skipped", func(t *testing.T) {
		cfg, err := mergeWithDotEnv(t, []string{"testdata/comments_and_blanks.env"})
		require.NoError(t, err)
		require.Equal(t, "dotenvhost", cfg.Host)
		require.Equal(t, 7777, cfg.Port)
	})

	t.Run("empty value sets field to empty string", func(t *testing.T) {
		cfg, err := mergeWithDotEnv(t, []string{"testdata/empty_value.env"})
		require.NoError(t, err)
		require.Equal(t, "", cfg.Host)
		require.Equal(t, 7777, cfg.Port)
	})

	t.Run("missing file is silently skipped", func(t *testing.T) {
		cfg, err := mergeWithDotEnv(t, []string{"/nonexistent/.env"})
		require.NoError(t, err)
		require.Equal(t, "localhost", cfg.Host)
		require.Equal(t, 8080, cfg.Port)
	})

	t.Run("line without equals sign returns error", func(t *testing.T) {
		_, err := mergeWithDotEnv(t, []string{"testdata/invalid_no_equals.env"})
		require.Error(t, err)
		require.ErrorContains(t, err, "missing '='")
	})

	t.Run("line with empty key returns error", func(t *testing.T) {
		_, err := mergeWithDotEnv(t, []string{"testdata/invalid_empty_key.env"})
		require.Error(t, err)
		require.ErrorContains(t, err, "empty key")
	})
}

func TestDotEnvPrecedence(t *testing.T) {
	t.Run("dotenv overrides yaml", func(t *testing.T) {
		yamlPath := writeYAML(t, "host: fromyaml\n")
		var cfg dotenvConfig
		err := internal.Merge(internal.MergeOptions{
			Paths:       []string{yamlPath},
			DotEnvFiles: []string{"testdata/valid.env"},
			Dest:        &cfg,
		})
		require.NoError(t, err)
		require.Equal(t, "dotenvhost", cfg.Host)
	})

	t.Run("real env var wins over dotenv", func(t *testing.T) {
		t.Setenv("KLEE_DOTENV_HOST", "realenv")
		cfg, err := mergeWithDotEnv(t, []string{"testdata/valid.env"})
		require.NoError(t, err)
		require.Equal(t, "realenv", cfg.Host)
		// port still comes from dotenv since only HOST is set in real env
		require.Equal(t, 7777, cfg.Port)
	})

	t.Run("later dotenv file wins over earlier for same key", func(t *testing.T) {
		second := filepath.Join(t.TempDir(), "override.env")
		require.NoError(t, os.WriteFile(second, []byte("KLEE_DOTENV_HOST=overridehost\n"), 0o644))

		cfg, err := mergeWithDotEnv(t, []string{"testdata/valid.env", second})
		require.NoError(t, err)
		require.Equal(t, "overridehost", cfg.Host)
		// port still from first file
		require.Equal(t, 7777, cfg.Port)
	})

	t.Run("defaults still apply for fields not in dotenv", func(t *testing.T) {
		second := filepath.Join(t.TempDir(), "partial.env")
		require.NoError(t, os.WriteFile(second, []byte("KLEE_DOTENV_HOST=partialhost\n"), 0o644))

		cfg, err := mergeWithDotEnv(t, []string{second})
		require.NoError(t, err)
		require.Equal(t, "partialhost", cfg.Host)
		require.Equal(t, 8080, cfg.Port) // default
		require.Equal(t, false, cfg.Debug) // default
	})
}
