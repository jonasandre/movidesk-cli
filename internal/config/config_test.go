package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv(envHome, dir)
	return dir
}

func TestLoad_NoFile_ReturnsEmpty(t *testing.T) {
	newTempHome(t)
	c, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, c.Tenants)
	assert.Empty(t, c.Tenants)
	assert.Empty(t, c.CurrentTenant)
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := newTempHome(t)

	c := &Config{
		CurrentTenant: "acme",
		Defaults:      Defaults{Output: "table", PageSize: 50},
		Tenants: map[string]*Tenant{
			"acme": {Name: "acme", Label: "Acme Prod"},
			"beta": {Name: "beta", Label: "Beta", Output: "csv"},
		},
	}
	require.NoError(t, c.Save())

	info, err := os.Stat(filepath.Join(dir, fileName))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(fileMode), info.Mode().Perm())

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "acme", loaded.CurrentTenant)
	assert.Equal(t, "table", loaded.Defaults.Output)
	assert.Equal(t, 50, loaded.Defaults.PageSize)
	assert.Len(t, loaded.Tenants, 2)
	assert.Equal(t, "acme", loaded.Tenants["acme"].Name)
	assert.Equal(t, "csv", loaded.Tenants["beta"].Output)
}

func TestResolve_OverrideWins(t *testing.T) {
	c := &Config{
		CurrentTenant: "acme",
		Tenants: map[string]*Tenant{
			"acme": {Name: "acme"},
			"beta": {Name: "beta"},
		},
	}
	t.Setenv(envTenant, "")
	got, err := c.Resolve("beta")
	require.NoError(t, err)
	assert.Equal(t, "beta", got.Name)
}

func TestResolve_EnvWins(t *testing.T) {
	t.Setenv(envTenant, "beta")
	c := &Config{
		CurrentTenant: "acme",
		Tenants: map[string]*Tenant{
			"acme": {Name: "acme"},
			"beta": {Name: "beta"},
		},
	}
	got, err := c.Resolve("")
	require.NoError(t, err)
	assert.Equal(t, "beta", got.Name)
}

func TestResolve_FallsBackToCurrent(t *testing.T) {
	t.Setenv(envTenant, "")
	c := &Config{
		CurrentTenant: "acme",
		Tenants:       map[string]*Tenant{"acme": {Name: "acme"}},
	}
	got, err := c.Resolve("")
	require.NoError(t, err)
	assert.Equal(t, "acme", got.Name)
}

func TestResolve_NoneConfigured(t *testing.T) {
	t.Setenv(envTenant, "")
	c := &Config{Tenants: map[string]*Tenant{}}
	_, err := c.Resolve("")
	assert.ErrorIs(t, err, ErrTenantNotFound)
}

func TestResolve_UnknownTenant(t *testing.T) {
	t.Setenv(envTenant, "")
	c := &Config{Tenants: map[string]*Tenant{}}
	_, err := c.Resolve("ghost")
	assert.ErrorIs(t, err, ErrTenantNotFound)
}

func TestEffectiveOutput_Priority(t *testing.T) {
	c := &Config{Defaults: Defaults{Output: "table"}}
	tn := &Tenant{Output: "csv"}

	assert.Equal(t, "json", c.EffectiveOutput(tn, "json"))
	assert.Equal(t, "csv", c.EffectiveOutput(tn, ""))
	assert.Equal(t, "table", c.EffectiveOutput(&Tenant{}, ""))
	assert.Equal(t, "json", (&Config{}).EffectiveOutput(nil, ""))
}

func TestEffectiveBaseURL(t *testing.T) {
	assert.Equal(t, DefaultBaseURL, (&Tenant{}).EffectiveBaseURL())
	assert.Equal(t, "https://sandbox/x", (&Tenant{BaseURL: "https://sandbox/x"}).EffectiveBaseURL())
}

func TestDelete_ClearsCurrent(t *testing.T) {
	c := &Config{
		CurrentTenant: "acme",
		Tenants:       map[string]*Tenant{"acme": {Name: "acme"}},
	}
	c.Delete("acme")
	assert.Empty(t, c.CurrentTenant)
	assert.Empty(t, c.Tenants)
}
