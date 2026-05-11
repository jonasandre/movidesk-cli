package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonasandre/movidesk-cli/internal/config"
)

// loadBody resolves the request body from --file, --from-template,
// --from-template-file, or --set k=v overrides. Exactly one of file/template
// is required, and --set values are merged on top.
func loadBody(file, template, templateFile string, sets []string) (map[string]any, error) {
	var body map[string]any
	switch {
	case file != "" && (template != "" || templateFile != ""):
		return nil, fmt.Errorf("--file e --from-template são mutuamente exclusivos")
	case template != "" && templateFile != "":
		return nil, fmt.Errorf("--from-template e --from-template-file são mutuamente exclusivos")
	case file != "":
		raw, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("ler arquivo de corpo: %w", err)
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, fmt.Errorf("interpretar arquivo de corpo: %w", err)
		}
	case template != "":
		dir, err := config.Dir()
		if err != nil {
			return nil, err
		}
		raw, err := os.ReadFile(filepath.Join(dir, "templates", template+".json"))
		if err != nil {
			return nil, fmt.Errorf("ler template %q: %w", template, err)
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, fmt.Errorf("interpretar template %q: %w", template, err)
		}
	case templateFile != "":
		raw, err := os.ReadFile(templateFile)
		if err != nil {
			return nil, fmt.Errorf("ler arquivo de template: %w", err)
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, fmt.Errorf("interpretar arquivo de template: %w", err)
		}
	default:
		body = map[string]any{}
	}

	for _, kv := range sets {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("valor --set deve ser chave=valor, recebido %q", kv)
		}
		body[k] = parseSetValue(v)
	}
	return body, nil
}

// parseSetValue tries JSON first (so the user can pass numbers, bools, arrays
// or objects via --set). Falls back to a string if parsing fails.
func parseSetValue(v string) any {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	var x any
	if err := json.Unmarshal([]byte(v), &x); err == nil {
		return x
	}
	return v
}
