package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonasandre/movidesk-cli/internal/movidesk"
	"github.com/jonasandre/movidesk-cli/internal/movidesk/odata"
)

const (
	uriODataFilterSyntax   = "movidesk://odata-filter-syntax"
	uriCustomfieldsCatalog = "movidesk://customfields-catalog"
	uriServerInfo          = "movidesk://server-info"
)

// registerResources publishes the static and tenant-derived resources the
// chat-app can fetch on demand. They give the model the same up-front
// guidance the CLI provides via `topics filters` plus a per-tenant custom-field
// catalog for label↔id resolution.
func registerResources(s *mcpsdk.Server, c *movidesk.Client, cfg Config) {
	s.AddResource(
		&mcpsdk.Resource{
			URI:         uriODataFilterSyntax,
			Name:        "odata-filter-syntax",
			Description: "Sintaxe OData $filter aceita pela API do Movidesk: operadores, funções, literais, campos comuns e armadilhas frequentes.",
			MIMEType:    "text/markdown",
		},
		func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
			_ = ctx
			return &mcpsdk.ReadResourceResult{
				Contents: []*mcpsdk.ResourceContents{{
					URI:      uriODataFilterSyntax,
					MIMEType: "text/markdown",
					Text:     odata.FilterTopic,
				}},
			}, nil
		},
	)

	s.AddResource(
		&mcpsdk.Resource{
			URI:         uriServerInfo,
			Name:        "server-info",
			Description: "Metadados de diagnóstico do servidor MCP: tenant ativo, base URL e limites de uso.",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
			_ = ctx
			info := map[string]any{
				"tenant":     cfg.Tenant,
				"base_url":   c.BaseURL,
				"rate_limit": "10 req/min (compartilhado com qualquer outro consumo do mesmo token)",
				"transport":  "stdio",
				"scope":      "read-only",
			}
			buf, err := json.Marshal(info)
			if err != nil {
				return nil, fmt.Errorf("server-info: %w", err)
			}
			return &mcpsdk.ReadResourceResult{
				Contents: []*mcpsdk.ResourceContents{{
					URI:      uriServerInfo,
					MIMEType: "application/json",
					Text:     string(buf),
				}},
			}, nil
		},
	)

	if cfg.CustomFields != nil {
		s.AddResource(
			&mcpsdk.Resource{
				URI:         uriCustomfieldsCatalog,
				Name:        "customfields-catalog",
				Description: "Catálogo local de custom fields do tenant ativo (lido de ~/.movidesk/<tenant>/customfields.yaml). Mapeia rótulo legível → {id, rule_id, type, options}. Use para resolver labels antes de filtrar/expandir.",
				MIMEType:    "application/json",
			},
			func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
				_ = ctx
				return &mcpsdk.ReadResourceResult{
					Contents: []*mcpsdk.ResourceContents{{
						URI:      uriCustomfieldsCatalog,
						MIMEType: "application/json",
						Text:     string(cfg.CustomFields),
					}},
				}, nil
			},
		)
	}
}
