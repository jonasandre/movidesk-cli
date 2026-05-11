// Package mcp embeds a Model Context Protocol server inside the CLI binary.
//
// The server runs over stdio so MCP-aware chat applications (Claude Desktop,
// Cline, etc.) can call Movidesk endpoints as MCP tools without spawning
// arbitrary CLI subprocesses themselves. All tools wrap the existing service
// packages under internal/movidesk; the package is read-only in v1.
package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

// StringList accepts three shapes from LLM clients and normalises to []string:
//
//   - canonical JSON array: ["id","name"]
//   - JSON-encoded string of an array: "[\"id\",\"name\"]"
//   - comma-separated string: "id,name"
//
// LLMs frequently mis-encode list parameters; absorbing the common variants at
// the wire keeps tool calls from failing schema validation while still
// preserving the canonical array form for well-behaved clients. The matching
// JSON-Schema relaxation lives in relaxStringListProps below.
type StringList []string

// UnmarshalJSON implements the lenient deserialisation described on
// [StringList].
func (sl *StringList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*sl = nil
		return nil
	}
	if arr, ok := tryJSONArray(data); ok {
		*sl = arr
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("StringList: esperado array de strings ou string, recebido %s", data)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		*sl = nil
		return nil
	}
	if strings.HasPrefix(s, "[") {
		if arr, ok := tryJSONArray([]byte(s)); ok {
			*sl = arr
			return nil
		}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	*sl = out
	return nil
}

func tryJSONArray(data []byte) ([]string, bool) {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil, false
	}
	return arr, true
}

// ODataArgs is the shared input shape embedded by every list-style tool. The
// MCP SDK generates a JSON-Schema from the struct tags; the `jsonschema` tag
// becomes the property description that LLMs see in tools/list responses.
type ODataArgs struct {
	Filter  string     `json:"filter,omitempty" jsonschema:"Expressão OData $filter. Consulte o resource movidesk://odata-filter-syntax para a sintaxe completa, operadores e armadilhas."`
	Select  StringList `json:"select,omitempty" jsonschema:"Campos a projetar ($select). Prefira JSON array (ex.: [\"id\",\"protocol\"])."`
	Expand  StringList `json:"expand,omitempty" jsonschema:"Relações a embarcar ($expand). Ex.: [\"actions\",\"customFieldValues\"]."`
	OrderBy StringList `json:"orderby,omitempty" jsonschema:"Ordenações ($orderby). Ex.: [\"createdDate desc\"]."`
	Top     int        `json:"top,omitempty" jsonschema:"Tamanho da página ($top). 0 = padrão do servidor. Tickets: ≤250 (clamp implícito)."`
	Skip    int        `json:"skip,omitempty" jsonschema:"Offset server-side ($skip)."`
	All     bool       `json:"all,omitempty" jsonschema:"Auto-paginar até esgotar resultados ou atingir 'max'."`
	Max     int        `json:"max,omitempty" jsonschema:"Com 'all'=true, limite de linhas. 0 aplica o padrão (500). Use valores explícitos para subir o teto."`
}

// Query maps the structured args to the internal odata.Query type consumed by
// every service package.
func (a ODataArgs) Query() odata.Query {
	return odata.Query{
		Filter:  a.Filter,
		Select:  []string(a.Select),
		Expand:  []string(a.Expand),
		OrderBy: []string(a.OrderBy),
		Top:     a.Top,
		Skip:    a.Skip,
	}
}

// buildInputSchema infers a JSON-Schema for the given args type, then relaxes
// the type constraint on every property whose Go type is StringList so the
// schema validator accepts the array/JSON-string/comma-string forms that
// UnmarshalJSON consumes. Without this step the SDK would reject non-array
// inputs before the handler ever ran.
func buildInputSchema[T any]() *jsonschema.Schema {
	s, err := jsonschema.For[T](nil)
	if err != nil {
		panic(fmt.Errorf("infer input schema: %w", err))
	}
	relaxStringListProps(s)
	return s
}

// stringListPropNames lists the ODataArgs fields backed by StringList. Only
// these properties get the permissive oneOf schema; other array-of-string
// fields that may appear in handler-specific arg types stay strict.
var stringListPropNames = map[string]bool{
	"select":  true,
	"expand":  true,
	"orderby": true,
}

func relaxStringListProps(s *jsonschema.Schema) {
	if s == nil {
		return
	}
	for name, prop := range s.Properties {
		if prop == nil || !stringListPropNames[name] {
			continue
		}
		s.Properties[name] = &jsonschema.Schema{
			Description: prop.Description +
				" Aceita JSON array (preferido), JSON-string (\"[\\\"a\\\",\\\"b\\\"]\") ou comma-string (\"a,b\").",
			OneOf: []*jsonschema.Schema{
				{Type: "array", Items: &jsonschema.Schema{Type: "string"}},
				{Type: "string"},
				{Type: "null"},
			},
		}
	}
}
