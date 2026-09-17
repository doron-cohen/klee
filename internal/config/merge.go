package config

import (
	"fmt"
	"reflect"
)

// MergeOptions holds the ordered file paths and destination struct.
type MergeOptions struct {
	// Paths in order of lowest to highest precedence (system → user → project).
	Paths []string
	// DotEnvFiles are .env files parsed for KEY=VALUE pairs.
	// Real environment variables take precedence over dotenv values.
	// Later files win for duplicate keys.
	DotEnvFiles []string
	// Dest is a pointer to the config struct to populate.
	Dest any
}

// Merge loads config layers in order, applies env vars, then applies defaults.
// Later layers win over earlier ones. Env vars win over all file layers.
// Defaults fill in any remaining zero values.
//
// The returned slice says, for each entry of Paths and in the same order,
// whether that file was actually read. A missing file is not an error, so
// this is the only place the read/not-read distinction exists.
func Merge(opts MergeOptions) ([]bool, error) {
	v := reflect.ValueOf(opts.Dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("dest must be a pointer to a struct")
	}

	infos := parseFields(v.Elem().Type(), nil)

	read := make([]bool, len(opts.Paths))
	for i, path := range opts.Paths {
		found, err := loadFile(path, opts.Dest)
		if err != nil {
			return read, err
		}
		read[i] = found
	}

	dotEnv, err := parseDotEnvFiles(opts.DotEnvFiles)
	if err != nil {
		return read, err
	}

	if err := applyEnv(infos, v, dotEnv); err != nil {
		return read, err
	}

	if err := applyDefaults(infos, v, dotEnv); err != nil {
		return read, err
	}

	return read, nil
}
