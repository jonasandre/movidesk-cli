// Package mcp embeds a Model Context Protocol server inside the CLI binary.
//
// The server runs over stdio so MCP-aware chat applications (Claude Desktop,
// Cline, etc.) can call Movidesk endpoints as MCP tools without spawning
// arbitrary CLI subprocesses themselves. All tools wrap the existing service
// packages under internal/movidesk; the package is read-only in v1.
package mcp

import "github.com/jonasandre/movidesk-cli/internal/movidesk/odata"

// ODataArgs is the shared input shape embedded by every list-style tool. The
// MCP SDK generates a JSON-Schema from the struct tags; the `jsonschema` tag
// becomes the property description that LLMs see in tools/list responses.
type ODataArgs struct {
	Filter  string   `json:"filter,omitempty" jsonschema:"Expressão OData $filter. Consulte o resource movidesk://odata-filter-syntax para a sintaxe completa, operadores e armadilhas."`
	Select  []string `json:"select,omitempty" jsonschema:"Campos a projetar ($select)."`
	Expand  []string `json:"expand,omitempty" jsonschema:"Relações a embarcar ($expand). Ex.: [\"actions\",\"customFieldValues\"]."`
	OrderBy []string `json:"orderby,omitempty" jsonschema:"Ordenações ($orderby). Ex.: [\"createdDate desc\"]."`
	Top     int      `json:"top,omitempty" jsonschema:"Tamanho da página ($top). 0 = padrão do servidor."`
	Skip    int      `json:"skip,omitempty" jsonschema:"Offset server-side ($skip)."`
	All     bool     `json:"all,omitempty" jsonschema:"Auto-paginar até esgotar resultados ou atingir 'max'."`
	Max     int      `json:"max,omitempty" jsonschema:"Com 'all'=true, limite de linhas. 0 aplica o padrão (500). Use valores explícitos para subir o teto."`
}

// Query maps the structured args to the internal odata.Query type consumed by
// every service package.
func (a ODataArgs) Query() odata.Query {
	return odata.Query{
		Filter:  a.Filter,
		Select:  a.Select,
		Expand:  a.Expand,
		OrderBy: a.OrderBy,
		Top:     a.Top,
		Skip:    a.Skip,
	}
}
