package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/jonasandre/movidesk-cli/internal/config"
)

// FieldType enumerates the custom field types the catalog tracks. The set
// matches the Movidesk types described in the public ticket schema doc.
type FieldType string

const (
	FieldText          FieldType = "text"
	FieldMultilineText FieldType = "multiline-text"
	FieldHTML          FieldType = "html"
	FieldRegex         FieldType = "regex"
	FieldNumber        FieldType = "number"
	FieldDate          FieldType = "date"
	FieldTime          FieldType = "time"
	FieldDateTime      FieldType = "datetime"
	FieldEmail         FieldType = "email"
	FieldPhone         FieldType = "phone"
	FieldURL           FieldType = "url"
	FieldListOfValues  FieldType = "list-of-values"
	FieldListOfPersons FieldType = "list-of-persons"
	FieldListOfClients FieldType = "list-of-clients"
	FieldListOfAgents  FieldType = "list-of-agents"
	FieldSingleSelect  FieldType = "single-select"
	FieldMultiSelect   FieldType = "multi-select"
)

// CatalogEntry describes a custom field as known locally per tenant.
type CatalogEntry struct {
	ID      int       `yaml:"id"`
	RuleID  int       `yaml:"rule_id"`
	Type    FieldType `yaml:"type"`
	Options []string  `yaml:"options,omitempty"`
}

// Catalog maps human label → entry. Stored under
// ~/.movidesk/<tenant>/customfields.yaml so each tenant has its own catalog.
type Catalog struct {
	Fields map[string]CatalogEntry `yaml:"fields"`
}

func catalogPath(tenant string) (string, error) {
	if tenant == "" {
		return "", errors.New("tenant é obrigatório para resolver caminho do catálogo")
	}
	d, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, tenant, "customfields.yaml"), nil
}

func loadCatalog(tenant string) (*Catalog, error) {
	p, err := catalogPath(tenant)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return &Catalog{Fields: map[string]CatalogEntry{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Catalog
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("interpretar catálogo: %w", err)
	}
	if c.Fields == nil {
		c.Fields = map[string]CatalogEntry{}
	}
	return &c, nil
}

func saveCatalog(tenant string, c *Catalog) error {
	p, err := catalogPath(tenant)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	raw, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// labelForID returns a label registered for the given numeric field id, or "".
func (c *Catalog) labelForID(id int) string {
	for label, entry := range c.Fields {
		if entry.ID == id {
			return label
		}
	}
	return ""
}

// sortedLabels returns labels ordered alphabetically for stable display.
func (c *Catalog) sortedLabels() []string {
	out := make([]string, 0, len(c.Fields))
	for k := range c.Fields {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
