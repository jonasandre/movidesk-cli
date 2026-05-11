package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jonasandre/movidesk-cli/internal/mcp"
)

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Inicia um servidor MCP via stdio expondo a API do Movidesk como ferramentas",
		Long: `Sobe um servidor Model Context Protocol no stdin/stdout do processo, expondo
operações somente-leitura da API do Movidesk como ferramentas (tools) e
recursos (resources) consumíveis por clientes MCP — Claude Desktop, Cline,
Continue, etc.

Configuração típica em Claude Desktop (~/Library/Application Support/Claude/
claude_desktop_config.json):

  {
    "mcpServers": {
      "movidesk": {
        "command": "movidesk-cli",
        "args": ["mcp", "--tenant", "acme"]
      }
    }
  }

O tenant é resolvido uma única vez no boot — um processo MCP atende um tenant.
Logs verbose (--verbose) vão para stderr; stdout é reservado ao framing
JSON-RPC, portanto não imprima nada manualmente nesse stream durante a sessão.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r, err := resolveClient(cmd)
			if err != nil {
				return err
			}

			cfg := mcp.Config{Tenant: r.tenant.Name}
			cfg.CustomFields = catalogJSON(r.tenant.Name)

			return mcp.Run(cmd.Context(), r.client, cfg, os.Stdin, os.Stdout)
		},
	}
}

// catalogJSON loads the local per-tenant custom-field catalog and returns it
// serialized as JSON. Missing/unreadable catalog is treated as "no resource":
// the MCP server simply omits movidesk://customfields-catalog and the chat app
// falls back to discovery via OData filters. We never block server startup on
// catalog problems.
func catalogJSON(tenant string) []byte {
	cat, err := loadCatalog(tenant)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: catálogo de custom fields indisponível: %v\n", err)
		return nil
	}
	if cat == nil || len(cat.Fields) == 0 {
		return nil
	}
	buf, err := json.Marshal(cat.Fields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: serializar catálogo: %v\n", err)
		return nil
	}
	return buf
}
