package output

import (
	"fmt"
	"strings"
)

// dig walks v using a dot-separated path (e.g. "owner.businessName"). Empty
// path returns v itself. Missing keys yield "" rather than an error to keep
// rendering forgiving.
func dig(v any, path string) any {
	if path == "" {
		return v
	}
	cur := v
	for _, part := range strings.Split(path, ".") {
		switch m := cur.(type) {
		case map[string]any:
			cur = m[part]
		case map[any]any:
			cur = m[part]
		default:
			return nil
		}
		if cur == nil {
			return nil
		}
	}
	return cur
}

// stringify renders v as a single-line cell value.
func stringify(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		// JSON numbers come back as float64; render integers cleanly.
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case []any:
		parts := make([]string, 0, len(t))
		for _, x := range t {
			parts = append(parts, stringify(x))
		}
		return strings.Join(parts, ",")
	case map[string]any:
		// Render compact for tables; long objects truncate.
		s := fmt.Sprintf("%v", t)
		if len(s) > 60 {
			return s[:60] + "…"
		}
		return s
	default:
		return fmt.Sprintf("%v", t)
	}
}

// asRows normalizes v to a slice of maps. Single objects become a 1-row slice.
func asRows(v any) []map[string]any {
	switch t := v.(type) {
	case []map[string]any:
		return t
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, x := range t {
			if m, ok := x.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case map[string]any:
		return []map[string]any{t}
	default:
		return nil
	}
}
