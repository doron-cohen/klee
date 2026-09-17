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

// Source is one candidate config file and whether Load read it.
type Source struct {
	// Path is the candidate file.
	Path string
	// Loaded is true if the file existed and was merged in.
	Loaded bool
}

// Load populates dest from config files and environment variables, and
// reports every file it considered, in ascending order of precedence.
// dest must be a pointer to a struct.
//
// Precedence (lowest to highest): system file → XDG config dirs → user file
// → project file → env vars → defaults. The XDG config dirs are the
// secondary directories from $XDG_CONFIG_DIRS or the platform defaults; on
// darwin they include ~/.config, which the config home does not.
//
// Sources come back even when loading fails, so a caller can show what had
// been read before the error.
func Load(dest any, opts Options) ([]Source, error) {
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

	// ConfigDirs is most-preferred first; this list is least-preferred first.
	paths := []string{filepath.Join("/etc", opts.AppName, opts.Filename)}
	secondary := dirs.ConfigDirs()
	for i := len(secondary) - 1; i >= 0; i-- {
		paths = append(paths, filepath.Join(secondary[i], opts.Filename))
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
