package config

import (
	"fmt"
	"path/filepath"

	internal "github.com/doron-cohen/klee/internal/config"
	"github.com/doron-cohen/klee/xdg"
)

// Options controls how config is loaded.
type Options struct {
	// AppName is used to resolve XDG paths.
	AppName string
	// ProjectPath overrides the project-level config file path.
	// Defaults to ./<appName>.yaml in the current directory.
	ProjectPath string
	// Filename is the config filename used under XDG dirs.
	// Defaults to "config.yaml".
	Filename string
}

// Load populates dest from config files and environment variables.
// dest must be a pointer to a struct.
//
// Precedence (lowest to highest): system file → user file → project file → env vars → defaults.
func Load(dest any, opts Options) error {
	if opts.Filename == "" {
		opts.Filename = "config.yaml"
	}

	dirs := xdg.New(opts.AppName)

	projectPath := opts.ProjectPath
	if projectPath == "" {
		projectPath = fmt.Sprintf("./%s.yaml", opts.AppName)
	}

	paths := []string{
		filepath.Join("/etc", opts.AppName, opts.Filename),
		dirs.ConfigFile(opts.Filename),
		projectPath,
	}

	return internal.Merge(internal.MergeOptions{
		Paths: paths,
		Dest:  dest,
	})
}
