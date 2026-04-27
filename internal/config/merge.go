package config

import (
	"fmt"
	"reflect"
)

// MergeOptions holds the ordered file paths and destination struct.
type MergeOptions struct {
	// Paths in order of lowest to highest precedence (system → user → project).
	Paths []string
	// Dest is a pointer to the config struct to populate.
	Dest any
}

// Merge loads config layers in order, applies env vars, then applies defaults.
// Later layers win over earlier ones. Env vars win over all file layers.
// Defaults fill in any remaining zero values.
func Merge(opts MergeOptions) error {
	v := reflect.ValueOf(opts.Dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	infos := parseFields(v.Elem().Type(), nil)

	for _, path := range opts.Paths {
		if err := loadFile(path, opts.Dest); err != nil {
			return err
		}
	}

	if err := applyEnv(infos, v); err != nil {
		return err
	}

	if err := applyDefaults(infos, v); err != nil {
		return err
	}

	return nil
}
