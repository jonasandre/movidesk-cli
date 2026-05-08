package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultBaseURL  = "https://api.movidesk.com/public/v1"
	DefaultOutput   = "json"
	DefaultPageSize = 100

	envTenant = "MOVIDESK_TENANT"
	envHome   = "MOVIDESK_HOME"

	dirName  = ".movidesk"
	fileName = "config.yaml"
	fileMode = 0o600
	dirMode  = 0o700
)

type Defaults struct {
	Output   string `yaml:"output,omitempty"`
	PageSize int    `yaml:"page_size,omitempty"`
}

type Tenant struct {
	Name        string `yaml:"-"`
	Label       string `yaml:"label,omitempty"`
	BaseURL     string `yaml:"base_url,omitempty"`
	Output      string `yaml:"output,omitempty"`
	DefaultUser string `yaml:"default_user,omitempty"`
}

type Config struct {
	CurrentTenant string             `yaml:"current_tenant,omitempty"`
	Defaults      Defaults           `yaml:"defaults,omitempty"`
	Tenants       map[string]*Tenant `yaml:"tenants,omitempty"`
}

var ErrTenantNotFound = errors.New("tenant not found")

func Dir() (string, error) {
	if h := os.Getenv(envHome); h != "" {
		return h, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, dirName), nil
}

func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, fileName), nil
}

func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{Tenants: map[string]*Tenant{}}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if c.Tenants == nil {
		c.Tenants = map[string]*Tenant{}
	}
	for name, t := range c.Tenants {
		t.Name = name
	}
	return &c, nil
}

func (c *Config) Save() error {
	d, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, dirMode); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	p := filepath.Join(d, fileName)
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, fileMode); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return os.Rename(tmp, p)
}

func (c *Config) Get(name string) (*Tenant, error) {
	t, ok := c.Tenants[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrTenantNotFound, name)
	}
	return t, nil
}

func (c *Config) Set(t *Tenant) {
	if c.Tenants == nil {
		c.Tenants = map[string]*Tenant{}
	}
	c.Tenants[t.Name] = t
}

func (c *Config) Delete(name string) {
	delete(c.Tenants, name)
	if c.CurrentTenant == name {
		c.CurrentTenant = ""
	}
}

// Resolve returns the active tenant, considering: explicit override, env var,
// CurrentTenant. Returns ErrTenantNotFound if none can be resolved.
func (c *Config) Resolve(override string) (*Tenant, error) {
	name := override
	if name == "" {
		name = os.Getenv(envTenant)
	}
	if name == "" {
		name = c.CurrentTenant
	}
	if name == "" {
		return nil, fmt.Errorf("%w: no tenant configured (run `movidesk-cli auth login --tenant <name>`)", ErrTenantNotFound)
	}
	return c.Get(name)
}

// EffectiveOutput returns output format, in priority: cli flag, tenant, defaults, "json".
func (c *Config) EffectiveOutput(t *Tenant, flag string) string {
	if flag != "" {
		return flag
	}
	if t != nil && t.Output != "" {
		return t.Output
	}
	if c.Defaults.Output != "" {
		return c.Defaults.Output
	}
	return DefaultOutput
}

// EffectiveBaseURL returns the API base URL, in priority: tenant override, default.
func (t *Tenant) EffectiveBaseURL() string {
	if t != nil && t.BaseURL != "" {
		return t.BaseURL
	}
	return DefaultBaseURL
}
