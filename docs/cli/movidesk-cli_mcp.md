## movidesk-cli mcp

Inicia um servidor MCP via stdio expondo a API do Movidesk como ferramentas

### Synopsis

Sobe um servidor Model Context Protocol no stdin/stdout do processo, expondo
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
JSON-RPC, portanto não imprima nada manualmente nesse stream durante a sessão.

```
movidesk-cli mcp [flags]
```

### Options

```
  -h, --help   help for mcp
```

### Options inherited from parent commands

```
      --compact         JSON compacto (sem indentação)
      --no-color        desativa cores na saída
      --no-retry        desativa retentativa automática em 429/5xx
  -o, --output string   formato de saída: json|table|csv (padrão: do tenant ou 'json')
      --tenant string   nome do tenant (sobrepõe o tenant atual; env: MOVIDESK_TENANT)
      --user string     id do usuário padrão (Cod. Ref.) usado em createdBy nas escritas; sobrepõe a configuração do tenant; env: MOVIDESK_USER
  -v, --verbose         log detalhado em stderr
```

### SEE ALSO

* [movidesk-cli](movidesk-cli.md)	 - CLI para a API REST do Movidesk

