package output

import (
	"encoding/csv"
	"io"
	"sort"
)

type csvFormatter struct{}

func (csvFormatter) Render(w io.Writer, v any, opts Options) error {
	rows := asRows(v)
	cw := csv.NewWriter(w)
	defer cw.Flush()

	if len(rows) == 0 {
		return nil
	}

	cols := opts.Columns
	if len(cols) == 0 && opts.Resource != "" {
		cols = Defaults(opts.Resource)
	}
	if len(cols) == 0 {
		cols = collectAllKeys(rows)
	}
	if err := cw.Write(cols); err != nil {
		return err
	}
	for _, row := range rows {
		out := make([]string, len(cols))
		for i, c := range cols {
			out[i] = stringify(dig(row, c))
		}
		if err := cw.Write(out); err != nil {
			return err
		}
	}
	return cw.Error()
}

func collectAllKeys(rows []map[string]any) []string {
	set := map[string]struct{}{}
	for _, r := range rows {
		for k := range r {
			set[k] = struct{}{}
		}
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
