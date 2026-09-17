package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"

	internal "github.com/doron-cohen/klee/internal/config"
	"github.com/doron-cohen/klee/xdg"
)

// ErrSecretNotFound is returned by SecretStore.Get when a key is not found.
var ErrSecretNotFound = errors.New("secret not found")

// SecretStore retrieves secrets by key.
type SecretStore interface {
	Get(key string) (string, error)
}

// WritableSecretStore extends SecretStore with write and delete operations.
type WritableSecretStore interface {
	SecretStore
	Set(key, value string) error
}

// FieldConfig is passed to ConfigurableField.Configure during the wiring pass.
type FieldConfig struct {
	Tags  reflect.StructTag
	Store SecretStore
}

// ConfigurableField is implemented by types that need to inspect their own
// struct tags and receive the secret store during config loading.
type ConfigurableField interface {
	Configure(FieldConfig) error
}

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
	// DisableConfigDirs restricts the user-level search to the XDG config
	// home, skipping the secondary XDG config directories
	// (xdg.Dirs.ConfigDirs) that are searched beneath it by default.
	//
	// Those directories include machine-wide ones, so an app that must not
	// take configuration written by anyone with admin rights wants this.
	// It is not needed merely to keep ~/.config out of the search: that is
	// the user's to control, via XDG_CONFIG_DIRS.
	DisableConfigDirs bool
	// DotEnvFiles are .env files to load KEY=VALUE pairs from.
	// Real environment variables take precedence over values in these files.
	DotEnvFiles []string
	// SecretStore is used to lazily fetch secrets for Secret fields.
	// Optional — if nil, Secret fields must be populated via env var.
	SecretStore SecretStore
}

var configurableFieldType = reflect.TypeOf((*ConfigurableField)(nil)).Elem()

// configureFields walks dest and calls Configure on any ConfigurableField fields.
func configureFields(v reflect.Value, store SecretStore) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		if reflect.PointerTo(f.Type).Implements(configurableFieldType) {
			cf := fv.Addr().Interface().(ConfigurableField)
			if err := cf.Configure(FieldConfig{Tags: f.Tag, Store: store}); err != nil {
				return fmt.Errorf("field %s: %w", f.Name, err)
			}
			continue
		}

		if fv.Kind() == reflect.Struct {
			if err := configureFields(fv, store); err != nil {
				return err
			}
		}
	}
	return nil
}

// Load populates dest from config files and environment variables.
// dest must be a pointer to a struct.
//
// Precedence (lowest to highest): system file → XDG config dirs → user file
// → project file → env vars → defaults.
func Load(dest any, opts Options) error {
	_, err := LoadWithSources(dest, opts)
	return err
}

// Source is one candidate config file and whether Load read it.
type Source struct {
	// Path is the candidate file.
	Path string
	// Loaded is true if the file existed and was merged in.
	Loaded bool
}

// LoadWithSources is Load, and also reports every file it considered, in
// ascending order of precedence. Apps use it to tell a user where config
// was looked for, which is otherwise impossible to answer from outside
// without reimplementing the search.
//
// Sources are returned even when loading fails, so an app can show what it
// had tried before the error.
func LoadWithSources(dest any, opts Options) ([]Source, error) {
	paths := searchPaths(opts)

	if err := configureFields(reflect.ValueOf(dest), opts.SecretStore); err != nil {
		return sources(paths, nil), fmt.Errorf("configuring fields: %w", err)
	}

	read, err := internal.Merge(internal.MergeOptions{
		Paths:       paths,
		DotEnvFiles: opts.DotEnvFiles,
		Dest:        dest,
	})
	return sources(paths, read), err
}

func sources(paths []string, read []bool) []Source {
	out := make([]Source, len(paths))
	for i, path := range paths {
		out[i] = Source{Path: path, Loaded: i < len(read) && read[i]}
	}
	return out
}

// searchPaths returns the config files to consider for opts, in ascending
// order of precedence.
func searchPaths(opts Options) []string {
	if opts.Filename == "" {
		opts.Filename = "config.yaml"
	}

	dirs := xdg.New(opts.AppName)

	projectPath := opts.ProjectPath
	if projectPath == "" {
		projectPath = fmt.Sprintf("./%s.yaml", opts.AppName)
	}

	paths := []string{filepath.Join("/etc", opts.AppName, opts.Filename)}
	if !opts.DisableConfigDirs {
		// ConfigDirs is most-preferred first; this list is least-preferred first.
		secondary := dirs.ConfigDirs()
		for i := len(secondary) - 1; i >= 0; i-- {
			paths = append(paths, filepath.Join(secondary[i], opts.Filename))
		}
	}
	paths = append(paths, dirs.ConfigFile(opts.Filename), projectPath)

	return dedupe(paths)
}

// dedupe drops repeated paths, keeping each one at its highest-precedence
// position. XDG_CONFIG_DIRS can name the config home, or /etc, and a file
// must not be merged in twice.
func dedupe(paths []string) []string {
	seen := make(map[string]bool, len(paths))
	out := make([]string, 0, len(paths))
	for i := len(paths) - 1; i >= 0; i-- {
		if seen[paths[i]] {
			continue
		}
		seen[paths[i]] = true
		out = append(out, paths[i])
	}
	slices.Reverse(out)
	return out
}
