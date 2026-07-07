package commands

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

// ConfigCommand returns the built-in config subcommand for inspecting
// and validating the loaded configuration.
func ConfigCommand(getConfig func(context.Context) any) *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "manage configuration",
		Commands: []*cli.Command{
			{
				Name:  "validate",
				Usage: "validate the configuration",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg := getConfig(ctx)
					if cfg == nil {
						return fmt.Errorf("no configuration loaded")
					}
					if v, ok := cfg.(interface{ Validate() error }); ok {
						if err := v.Validate(); err != nil {
							return fmt.Errorf("config validation failed: %w", err)
						}
					}
					fmt.Println("config valid")
					return nil
				},
			},
			{
				Name:  "print",
				Usage: "print the resolved configuration (secrets redacted)",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg := getConfig(ctx)
					if cfg == nil {
						return fmt.Errorf("no configuration loaded")
					}
					tree := structToMap(reflect.ValueOf(cfg))
					if cmd.Bool("json") {
						out, err := json.MarshalIndent(tree, "", "  ")
						if err != nil {
							return fmt.Errorf("marshaling config: %w", err)
						}
						fmt.Println(string(out))
					} else {
						out, err := yaml.Marshal(tree)
						if err != nil {
							return fmt.Errorf("marshaling config: %w", err)
						}
						fmt.Print(string(out))
					}
					return nil
				},
			},
		},
	}
}

// structToMap walks a struct via reflection and returns a map[string]any
// suitable for YAML/JSON marshaling. Fields tagged with secret:"true" are
// replaced with "****". Anonymous structs without a yaml tag are inlined.
func structToMap(v reflect.Value) map[string]any {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]any)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		if !fv.CanInterface() {
			continue
		}

		yamlTag := f.Tag.Get("yaml")
		key, _, _ := strings.Cut(yamlTag, ",")
		if key == "-" {
			continue
		}

		if f.Tag.Get("secret") == "true" {
			if key == "" {
				key = strings.ToLower(f.Name)
			}
			result[key] = "****"
			continue
		}

		if fv.Kind() == reflect.Struct && !implementsTextUnmarshaler(fv) {
			nested := structToMap(fv)
			if f.Anonymous && key == "" {
				for k, val := range nested {
					result[k] = val
				}
			} else {
				if key == "" {
					key = strings.ToLower(f.Name)
				}
				result[key] = nested
			}
			continue
		}

		if key == "" {
			key = strings.ToLower(f.Name)
		}

		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				result[key] = nil
			} else {
				result[key] = fv.Elem().Interface()
			}
			continue
		}

		result[key] = fv.Interface()
	}

	return result
}

func implementsTextUnmarshaler(v reflect.Value) bool {
	if v.Type().Implements(textUnmarshalerType) {
		return true
	}
	if v.CanAddr() && v.Addr().Type().Implements(textUnmarshalerType) {
		return true
	}
	return reflect.PointerTo(v.Type()).Implements(textUnmarshalerType)
}
