package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/doron-cohen/klee/config"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Host string `yaml:"host" env:"KLEE_TEST_HOST" default:"localhost"`
	Port int    `yaml:"port" env:"KLEE_TEST_PORT" default:"8080"`
	DB   struct {
		Name string `yaml:"name" env:"KLEE_TEST_DB_NAME" default:"mydb"`
	} `yaml:"db"`
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		env         map[string]string
		projectPath string // set to "-" to use no project file
		wantHost    string
		wantPort    int
		wantDBName  string
		wantErr     bool
	}{
		{
			name:       "defaults only",
			wantHost:   "localhost",
			wantPort:   8080,
			wantDBName: "mydb",
		},
		{
			name:       "yaml overrides defaults",
			yaml:       "host: myserver\nport: 9090\n",
			wantHost:   "myserver",
			wantPort:   9090,
			wantDBName: "mydb",
		},
		{
			name:       "env overrides defaults",
			env:        map[string]string{"KLEE_TEST_PORT": "7777"},
			wantHost:   "localhost",
			wantPort:   7777,
			wantDBName: "mydb",
		},
		{
			name:       "env overrides yaml",
			yaml:       "host: fromfile\n",
			env:        map[string]string{"KLEE_TEST_HOST": "fromenv"},
			wantHost:   "fromenv",
			wantPort:   8080,
			wantDBName: "mydb",
		},
		{
			name:       "nested struct from yaml",
			yaml:       "db:\n  name: mydb2\n",
			wantHost:   "localhost",
			wantPort:   8080,
			wantDBName: "mydb2",
		},
		{
			name:       "nested struct from env",
			env:        map[string]string{"KLEE_TEST_DB_NAME": "envdb"},
			wantHost:   "localhost",
			wantPort:   8080,
			wantDBName: "envdb",
		},
		{
			name:    "invalid yaml returns error",
			yaml:    "host: [not: valid yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			opts := config.Options{AppName: "kleetest"}

			if tt.yaml != "" {
				dir := t.TempDir()
				path := filepath.Join(dir, "config.yaml")
				require.NoError(t, os.WriteFile(path, []byte(tt.yaml), 0o644))
				opts.ProjectPath = path
			} else {
				opts.ProjectPath = "/nonexistent/klee-test-config.yaml"
			}

			var cfg testConfig
			err := config.Load(&cfg, opts)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantHost, cfg.Host)
			require.Equal(t, tt.wantPort, cfg.Port)
			require.Equal(t, tt.wantDBName, cfg.DB.Name)
		})
	}
}
