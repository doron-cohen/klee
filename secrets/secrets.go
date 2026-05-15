package secrets

import (
	"fmt"
	"sync"

	"github.com/doron-cohen/klee/config"
)

// Secret is a lazily-loaded secret value. It is wired with metadata during
// config loading and fetches its value on the first call to Value.
//
// Precedence: value set via UnmarshalText (env var or explicit) → secret store.
type Secret struct {
	envKey    string
	secretKey string
	store     config.SecretStore

	mu     sync.Mutex
	cached string
	loaded bool
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cached = string(b)
	s.loaded = true
	return nil
}

// Value returns the secret. Cached after first successful fetch.
func (s *Secret) Value() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded {
		return s.cached, nil
	}
	if s.store == nil {
		return "", fmt.Errorf("secret %q: no store configured and no value provided", s.secretKey)
	}
	if s.secretKey == "" {
		return "", fmt.Errorf("secret has no key configured")
	}
	v, err := s.store.Get(s.secretKey)
	if err != nil {
		return "", err
	}
	s.cached = v
	s.loaded = true
	return s.cached, nil
}
