package secrets

import (
	"fmt"

	"github.com/doron-cohen/klee/config"
)

// Secret is a lazily-resolved secret value. It is wired with metadata during
// config loading and resolves its value on each call to Value.
//
// Precedence: value set via UnmarshalText (env var or explicit) → secret store.
//
// Secret is safe to copy: it holds no mutable state.
type Secret struct {
	envKey    string
	secretKey string
	store     config.SecretStore
	value     string
	loaded    bool
}

// Literal returns a Secret pre-loaded with value. Intended for tests.
func Literal(value string) Secret {
	return Secret{value: value, loaded: true}
}

// Configure implements config.ConfigurableField.
// Called by klee during the field wiring pass to record tag metadata and store.
func (s *Secret) Configure(cfg config.FieldConfig) error {
	s.envKey = cfg.Tags.Get("env")
	s.secretKey = cfg.Tags.Get("secret")
	s.store = cfg.Store
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// Called by klee when a string value is available (env var).
func (s *Secret) UnmarshalText(b []byte) error {
	s.value = string(b)
	s.loaded = true
	return nil
}

// Value returns the secret. If set via env var, returns that directly.
// Otherwise fetches from the configured store on each call.
func (s *Secret) Value() (string, error) {
	if s.loaded {
		return s.value, nil
	}
	if s.store == nil {
		return "", fmt.Errorf("secret %q: no store configured and no value provided", s.secretKey)
	}
	if s.secretKey == "" {
		return "", fmt.Errorf("secret has no key configured")
	}
	return s.store.Get(s.secretKey)
}
