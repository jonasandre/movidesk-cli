package output

import (
	"encoding/json"
	"io"
)

type jsonFormatter struct{}

func (jsonFormatter) Render(w io.Writer, v any, opts Options) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if !opts.Compact {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(v)
}
