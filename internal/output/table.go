package output

import (
	"fmt"
	"io"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
)

type tableFormatter struct{}

func (tableFormatter) Render(w io.Writer, v any, opts Options) error {
	rows := asRows(v)
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "(sem linhas)")
		return err
	}

	cols := opts.Columns
	if len(cols) == 0 && opts.Resource != "" {
		cols = Defaults(opts.Resource)
	}
	if len(cols) == 0 {
		cols = pickColumns(rows[0])
	}

	tw := table.NewWriter()
	tw.SetOutputMirror(w)
	if !opts.Color {
		tw.SetStyle(table.StyleLight)
	} else {
		tw.SetStyle(table.StyleColoredBright)
	}

	header := make(table.Row, len(cols))
	for i, c := range cols {
		header[i] = c
	}
	tw.AppendHeader(header)

	for _, row := range rows {
		r := make(table.Row, len(cols))
		for i, c := range cols {
			r[i] = stringify(dig(row, c))
		}
		tw.AppendRow(r)
	}
	tw.Render()
	return nil
}

// pickColumns returns up to 6 top-level keys of an object, sorted, preferring
// common identifier-ish names first.
func pickColumns(row map[string]any) []string {
	preferred := []string{"id", "protocol", "subject", "name", "businessName", "status", "type", "isActive", "createdDate", "lastUpdate"}
	picked := []string{}
	seen := map[string]bool{}
	for _, k := range preferred {
		if _, ok := row[k]; ok {
			picked = append(picked, k)
			seen[k] = true
		}
		if len(picked) >= 6 {
			return picked
		}
	}
	keys := make([]string, 0, len(row))
	for k := range row {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		if len(picked) >= 6 {
			break
		}
		picked = append(picked, k)
	}
	return picked
}
