package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/jonasandre/movidesk-cli/internal/output"
)

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

// renderRows is like renderJSON but accepts a pre-decoded slice. Used by --all.
func renderRows(w io.Writer, rows any, format, resource string, columns []string) error {
	opts := output.Options{
		Compact:  flags.compact,
		Color:    !flags.noColor && isTerminal(w),
		Columns:  columns,
		Resource: resource,
	}
	return output.Render(w, format, rows, opts)
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
