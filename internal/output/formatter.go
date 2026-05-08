// Package output renders API responses as JSON, table, or CSV.
package output

import (
	"errors"
	"fmt"
	"io"
)

const (
	FormatJSON  = "json"
	FormatTable = "table"
	FormatCSV   = "csv"
)

// Options configure rendering.
type Options struct {
	Compact bool     // JSON only: no indentation
	Color   bool     // table only: ANSI colors
	Columns []string // table/csv: explicit column list (dot-paths)
}

// Formatter renders v to w. v is typically a []any or a single object decoded
// from the Movidesk API response.
type Formatter interface {
	Render(w io.Writer, v any, opts Options) error
}

// Get returns the formatter for a name, or an error if unknown.
func Get(name string) (Formatter, error) {
	switch name {
	case FormatJSON, "":
		return jsonFormatter{}, nil
	case FormatTable:
		return tableFormatter{}, nil
	case FormatCSV:
		return csvFormatter{}, nil
	default:
		return nil, fmt.Errorf("unknown output format %q (expected json|table|csv)", name)
	}
}

// Render is a top-level convenience.
func Render(w io.Writer, name string, v any, opts Options) error {
	if w == nil {
		return errors.New("nil writer")
	}
	f, err := Get(name)
	if err != nil {
		return err
	}
	return f.Render(w, v, opts)
}
