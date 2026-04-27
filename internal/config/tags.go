package config

import (
	"reflect"
	"strings"
)

type fieldInfo struct {
	Index      []int
	YAMLKey    string
	EnvKey     string
	Default    string
	HasDefault bool
	Secret     bool
	FlagName   string
}

// parseFields walks a struct type and returns field metadata.
// Handles embedded and nested structs recursively.
func parseFields(t reflect.Type, indexPrefix []int) []fieldInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var fields []fieldInfo
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		index := append(append([]int{}, indexPrefix...), i)

		if f.Anonymous || f.Type.Kind() == reflect.Struct {
			fields = append(fields, parseFields(f.Type, index)...)
			continue
		}

		info := fieldInfo{
			Index:  index,
			Secret: f.Tag.Get("secret") == "true",
		}

		if v := f.Tag.Get("yaml"); v != "" {
			info.YAMLKey = strings.Split(v, ",")[0]
		}
		if v := f.Tag.Get("env"); v != "" {
			info.EnvKey = v
		}
		if v := f.Tag.Get("default"); v != "" {
			info.Default = v
			info.HasDefault = true
		}
		if v := f.Tag.Get("flag"); v != "" {
			info.FlagName = v
		}

		fields = append(fields, info)
	}
	return fields
}
