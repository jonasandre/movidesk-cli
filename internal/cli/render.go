package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jonasandre/movidesk-cli/internal/output"
)

var _ = fmt.Sprintf // keep fmt import even if downstream callers stop using it

// renderJSON parses raw API JSON into a generic value and renders it via the
// configured formatter. resource sets default columns for table/csv when the
// user didn't override them.
func renderJSON(w io.Writer, raw []byte, format, resource string, columns []string) error {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	opts := output.Options{
		Compact:  flags.compact,
		Color:    !flags.noColor && isTerminal(w),
		Columns:  columns,
		Resource: resource,
	}
	return output.Render(w, format, v, opts)
}

// renderRows accepts any value (typed slice, single struct, []json.RawMessage,
// etc.) and renders it. For non-JSON formats we round-trip through encoding/json
// so typed Go values become map[string]any compatible with the column dig/path
// logic. The JSON formatter sees the original value and uses Marshaler hooks.
func renderRows(w io.Writer, rows any, format, resource string, columns []string) error {
	opts := output.Options{
		Compact:  flags.compact,
		Color:    !flags.noColor && isTerminal(w),
		Columns:  columns,
		Resource: resource,
	}
	if format == output.FormatJSON || format == "" {
		return output.Render(w, format, rows, opts)
	}
	normalized, err := normalize(rows)
	if err != nil {
		return err
	}
	return output.Render(w, format, normalized, opts)
}

// normalize converts any Go value into a generic any backed by map[string]any
// or []any. This lets the table/csv formatters introspect fields uniformly.
func normalize(v any) (any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("normalize: %w", err)
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("normalize: %w", err)
	}
	return out, nil
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
