package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// loadFile unmarshals a YAML file into dest.
// Returns nil if the file does not exist.
func loadFile(path string, dest any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading config file %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("parsing config file %s: %w", path, err)
	}
	return nil
}

// applyEnv sets struct fields from environment variables based on fieldInfo.
func applyEnv(infos []fieldInfo, dest reflect.Value) error {
	if dest.Kind() == reflect.Ptr {
		dest = dest.Elem()
	}
	for _, info := range infos {
		if info.EnvKey == "" {
			continue
		}
		val, ok := os.LookupEnv(info.EnvKey)
		if !ok {
			continue
		}
		field := dest.FieldByIndex(info.Index)
		if err := setField(field, val); err != nil {
			return fmt.Errorf("env %s: %w", info.EnvKey, err)
		}
	}
	return nil
}

// applyDefaults sets zero-value fields to their declared defaults.
func applyDefaults(infos []fieldInfo, dest reflect.Value) error {
	if dest.Kind() == reflect.Ptr {
		dest = dest.Elem()
	}
	for _, info := range infos {
		if !info.HasDefault {
			continue
		}
		field := dest.FieldByIndex(info.Index)
		if !field.IsZero() {
			continue
		}
		if err := setField(field, info.Default); err != nil {
			return fmt.Errorf("default for field: %w", err)
		}
	}
	return nil
}

func setField(field reflect.Value, raw string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		field.SetFloat(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}
