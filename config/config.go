package config

import (
	"fmt"
	"path/filepath"
	"reflect"

	internal "github.com/doron-cohen/klee/internal/config"
	"github.com/doron-cohen/klee/xdg"
)

// SecretStore retrieves secrets by key.
type SecretStore interface {
	Get(key string) (string, error)
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

// Load populates dest from config files and environment variables.
// dest must be a pointer to a struct.
//
// Precedence (lowest to highest): system file → user file → project file → env vars → defaults.
func Load(dest any, opts Options) error {
	if opts.Filename == "" {
		opts.Filename = "config.yaml"
	}

	v := reflect.ValueOf(dest)
	if err := configureFields(v, opts.SecretStore); err != nil {
		return fmt.Errorf("configuring fields: %w", err)
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
		Paths:       paths,
		DotEnvFiles: opts.DotEnvFiles,
		Dest:        dest,
	})
}
