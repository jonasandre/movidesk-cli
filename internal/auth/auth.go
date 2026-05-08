// Package auth manages Movidesk API tokens per tenant. Tokens live in the OS
// keychain (macOS Keychain, Windows Credential Manager, Linux libsecret/kwallet)
// when available. When no keychain is reachable, tokens fall back to an
// AES-GCM encrypted file under the config dir.
package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	// Service is the keychain service name used for all tokens.
	Service = "movidesk-cli"

	// EnvToken short-circuits all storage when set.
	EnvToken = "MOVIDESK_TOKEN"
)

var ErrNotFound = errors.New("token not found")

// Store abstracts token storage so tests can inject a fake.
type Store interface {
	Get(tenant string) (string, error)
	Set(tenant, token string) error
	Delete(tenant string) error
}

// keyringStore is the production store backed by go-keyring with a file fallback.
type keyringStore struct {
	fallback Store
}

// New returns the default Store: keychain with encrypted-file fallback.
func New() Store {
	return &keyringStore{fallback: newFileStore()}
}

func (s *keyringStore) Get(tenant string) (string, error) {
	tok, err := keyring.Get(Service, tenant)
	if err == nil {
		return tok, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		// Try fallback (in case it was previously stored there).
		return s.fallback.Get(tenant)
	}
	if isKeyringUnavailable(err) {
		return s.fallback.Get(tenant)
	}
	return "", fmt.Errorf("keyring get: %w", err)
}

func (s *keyringStore) Set(tenant, token string) error {
	if err := keyring.Set(Service, tenant, token); err == nil {
		return nil
	} else if !isKeyringUnavailable(err) {
		return fmt.Errorf("keyring set: %w", err)
	}
	return s.fallback.Set(tenant, token)
}

func (s *keyringStore) Delete(tenant string) error {
	err := keyring.Delete(Service, tenant)
	if err == nil {
		// Also clear fallback to be safe.
		_ = s.fallback.Delete(tenant)
		return nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		return s.fallback.Delete(tenant)
	}
	if isKeyringUnavailable(err) {
		return s.fallback.Delete(tenant)
	}
	return fmt.Errorf("keyring delete: %w", err)
}

// isKeyringUnavailable returns true when the OS keychain backend is missing
// (typically headless Linux without libsecret/kwallet).
func isKeyringUnavailable(err error) bool {
	// go-keyring exposes ErrUnsupportedPlatform on some backends and a generic
	// dbus error elsewhere. Match both.
	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		return true
	}
	msg := err.Error()
	return containsAny(msg, "dbus", "no such interface", "secret service", "Service was not provided")
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) == 0 {
			continue
		}
		if indexOf(s, sub) >= 0 {
			return true
		}
	}
	return false
}

func indexOf(s, sub string) int {
	// Simple substring search; avoids importing strings for one usage.
	if len(sub) > len(s) {
		return -1
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// ResolveToken returns the token for a tenant, honoring the EnvToken override.
func ResolveToken(s Store, tenant string) (string, error) {
	if v := os.Getenv(EnvToken); v != "" {
		return v, nil
	}
	tok, err := s.Get(tenant)
	if err != nil {
		return "", err
	}
	if tok == "" {
		return "", fmt.Errorf("%w: tenant %q", ErrNotFound, tenant)
	}
	return tok, nil
}
