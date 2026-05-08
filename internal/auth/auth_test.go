package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memStore struct {
	data map[string]string
}

func newMem() *memStore { return &memStore{data: map[string]string{}} }

func (m *memStore) Get(t string) (string, error) {
	v, ok := m.data[t]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (m *memStore) Set(t, v string) error  { m.data[t] = v; return nil }
func (m *memStore) Delete(t string) error  { delete(m.data, t); return nil }

func TestResolveToken_EnvOverrides(t *testing.T) {
	t.Setenv(EnvToken, "env-token")
	s := newMem()
	tok, err := ResolveToken(s, "anytenant")
	require.NoError(t, err)
	assert.Equal(t, "env-token", tok)
}

func TestResolveToken_FromStore(t *testing.T) {
	t.Setenv(EnvToken, "")
	s := newMem()
	require.NoError(t, s.Set("acme", "abc-123"))
	tok, err := ResolveToken(s, "acme")
	require.NoError(t, err)
	assert.Equal(t, "abc-123", tok)
}

func TestResolveToken_NotFound(t *testing.T) {
	t.Setenv(EnvToken, "")
	s := newMem()
	_, err := ResolveToken(s, "ghost")
	require.True(t, errors.Is(err, ErrNotFound))
}

func TestEncodePeek(t *testing.T) {
	short := EncodePeek("short")
	assert.Equal(t, "********", short)

	full := EncodePeek("abcdefghij")
	assert.True(t, len(full) > 4 && full[:4] == "abcd")
}

func TestFileStore_RoundTrip(t *testing.T) {
	t.Setenv("MOVIDESK_HOME", t.TempDir())
	t.Setenv(EnvPassphrase, "test-passphrase")

	s := newFileStore()
	require.NoError(t, s.Set("acme", "tok-1"))
	require.NoError(t, s.Set("beta", "tok-2"))

	got, err := s.Get("acme")
	require.NoError(t, err)
	assert.Equal(t, "tok-1", got)

	got, err = s.Get("beta")
	require.NoError(t, err)
	assert.Equal(t, "tok-2", got)

	require.NoError(t, s.Delete("acme"))
	_, err = s.Get("acme")
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestFileStore_DeleteLastRemovesFile(t *testing.T) {
	t.Setenv("MOVIDESK_HOME", t.TempDir())
	t.Setenv(EnvPassphrase, "p")

	s := newFileStore()
	require.NoError(t, s.Set("only", "tok"))
	require.NoError(t, s.Delete("only"))
	_, err := s.Get("only")
	assert.True(t, errors.Is(err, ErrNotFound))
}
