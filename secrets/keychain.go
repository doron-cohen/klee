package secrets

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

// Keychain retrieves secrets from the OS keychain.
// Uses go-keyring which routes to the appropriate backend per platform
// (macOS Keychain, Linux Secret Service, Windows Credential Manager).
type Keychain struct {
	service string
}

// NewKeychain returns a Keychain that scopes all lookups under service.
func NewKeychain(service string) *Keychain {
	return &Keychain{service: service}
}

// Get retrieves the secret for key under the configured service.
func (k *Keychain) Get(key string) (string, error) {
	val, err := keyring.Get(k.service, key)
	if err == keyring.ErrNotFound {
		return "", fmt.Errorf("secret %q not found in keychain (service: %q)", key, k.service)
	}
	if err != nil {
		return "", fmt.Errorf("keychain get %q: %w", key, err)
	}
	return val, nil
}
