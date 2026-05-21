# movidesk-cli

Interface de linha de comando para a API REST pública do
[Movidesk](https://www.movidesk.com).

- Multi-tenant: mantenha tokens separados para sandbox/staging/produção e
  alterne com um único comando.
- Tokens armazenados no chaveiro do sistema operacional (macOS Keychain,
  Windows Credential Manager, Linux libsecret/kwallet); fallback em
  arquivo criptografado para ambientes headless.
- Passthrough OData em todo endpoint list: `--filter`, `--select`,
  `--expand`, `--orderby`, `--top`, `--skip`, mais auto-paginação via
  `--all`.
- Saída em JSON (padrão), tabela legível ou CSV.
- Respeita o limite de 10 requisições/minuto do Movidesk com backoff
  automático e suporte ao header `Retry-After`.
- Usuário padrão por tenant que o CLI injeta automaticamente em
  `createdBy` nas escritas que exigem atribuição.

## Instalação

### macOS

```bash
# Homebrew — distribuído como Cask
brew install --cask jonasandre/movidesk/movidesk-cli

# Ou baixe o tar.gz correspondente em:
#   https://github.com/jonasandre/movidesk-cli/releases
```

> **Nota macOS — primeira execução.** O binário publicado ainda não é
> notarizado pela Apple, então o Gatekeeper o coloca em quarentena na
> primeira execução e o processo é encerrado silenciosamente
> (exit `137`). Remova o atributo de quarentena uma vez após a
> instalação:
>
> ```bash
> xattr -dr com.apple.quarantine "$(brew --caskroom)/movidesk-cli"
> # ou, sem Homebrew:
> xattr -dr com.apple.quarantine /caminho/para/movidesk-cli
> ```
>
> Repita após cada atualização até que a assinatura/notarização seja
> implantada (acompanhado em [`docs/RELEASING.md`](./docs/RELEASING.md)).

### Linux

```bash
# Homebrew — distribuído como Cask
brew install --cask jonasandre/movidesk/movidesk-cli

# Ou extraia o tar.gz correspondente em:
#   https://github.com/jonasandre/movidesk-cli/releases
tar -xzf movidesk-cli_<versao>_linux_amd64.tar.gz
sudo install -m 0755 movidesk-cli /usr/local/bin/movidesk-cli
```

Os releases publicados incluem binários para `linux_amd64` e
`linux_arm64`.

### Windows

```powershell
# Um comando só: baixa o último release, extrai, adiciona ao Path do usuário
# e instala a skill do Claude se ~/.claude já existir
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/jonasandre/movidesk-cli/main/scripts/install-windows.ps1 | iex"
```

Se preferir revisar o script antes de executar, ele está em
[`scripts/install-windows.ps1`](./scripts/install-windows.ps1). Também dá
para instalar uma versão específica:

```powershell
Invoke-WebRequest https://raw.githubusercontent.com/jonasandre/movidesk-cli/main/scripts/install-windows.ps1 -OutFile .\install-windows.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File .\install-windows.ps1 -Version 1.2.1
```

Se o instalador detectar `~/.claude`, ele também baixa
`movidesk-mcp/SKILL.md` da mesma versão instalada e a registra em
`~/.claude/skills/movidesk-mcp/`.

Fallback manual:

```powershell
# Baixe o zip correspondente em:
#   https://github.com/jonasandre/movidesk-cli/releases
$installDir = "$env:LOCALAPPDATA\Programs\movidesk-cli"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Expand-Archive .\movidesk-cli_<versao>_windows_amd64.zip -DestinationPath $installDir -Force
```

Os releases publicados incluem binários para `windows_amd64` e
`windows_arm64`.

### Compilando da fonte

```bash
go install github.com/jonasandre/movidesk-cli/cmd/movidesk-cli@latest
```

Habilitar autocompletar no shell (Cobra entrega bash, zsh, fish, PowerShell):

```bash
movidesk-cli completion zsh  > "${fpath[1]}/_movidesk-cli"
movidesk-cli completion bash > /usr/local/etc/bash_completion.d/movidesk-cli
movidesk-cli completion fish > ~/.config/fish/completions/movidesk-cli.fish
```

```powershell
movidesk-cli completion powershell >> $PROFILE
```

O Cask do Homebrew registra automaticamente as completions de `bash`,
`zsh` e `fish`; os snippets acima só são necessários para instalações
fora do Homebrew.

## Início rápido

```bash
# 1. Adicione um tenant. O token é lido em prompt oculto e validado
#    contra GET /persons?$top=1 antes de ser armazenado. Opcionalmente
#    configure um usuário padrão (Cod. Ref.) para atribuição em createdBy.
movidesk-cli auth login --tenant prod --label "Acme Produção" --make-default

# 2. Listar tenants configurados (o token nunca é exibido).
movidesk-cli auth list

# 3. Validar a conexão.
movidesk-cli auth status

# 4. Usar.
movidesk-cli tickets list --top 5 --output table
```

A referência completa de cada comando fica em [`docs/cli/`](./docs/cli/).
Alguns pontos de entrada comuns:

| Família | Destaques |
|---|---|
| `auth` | `login`, `list`, `switch`, `status`, `set-user`, `logout`, `token` |
| `tickets` | `list`, `get`, `create`, `update`, `bulk-update`, `bulk-close`, `bulk-cancel`, `attach`, `actions`, `clients`, `relations`, `timeline`, `customfields`, `past`, `merged`, `html` |
| `persons` | `list`, `get`, `create`, `update`, `delete`, `customfields` |
| `services` | `list`, `get`, `create`, `update`, `delete` |
| `activities` | `list`, `get`, `create`, `update`, `delete`, `add-teams` |
| `contracts` | `list`, `get`, `create`, `update`, `delete`, `consumption list` |
| `surveys` | `questions list/get`, `responses list` |
| `kb` | `articles get` |
| `telephony` | `queue`, `nonqueue`, `made-call-link` |
| `customfields` | `options add/rename/remove` |
| `query` | escape hatch OData bruto |
| `topics` | páginas de ajuda longas (ex.: `topics filters` — sintaxe OData) |

`movidesk-cli <comando> --help` sempre exibe as flags locais. As páginas
de referência em `docs/cli/` são geradas automaticamente e cobrem
todo subcomando. Para a sintaxe completa aceita por `--filter` /
`--select` / `--orderby` (operadores, funções, datas, campos comuns por
recurso e armadilhas como `ownerTeam` ser string e não navegação), rode
`movidesk-cli topics filters`.

## Servidor MCP (uso em chat apps)

A CLI também sobe como **servidor [Model Context Protocol](https://modelcontextprotocol.io/)**
via `movidesk-cli mcp`, expondo a API do Movidesk como ferramentas
estruturadas para chat apps que falam MCP (Claude Desktop, Cline, Continue,
etc.) e não conseguem chamar CLIs arbitrárias.

```jsonc
// ~/Library/Application Support/Claude/claude_desktop_config.json
{
  "mcpServers": {
    "movidesk": {
      "command": "movidesk-cli",
      "args": ["mcp", "--tenant", "acme"]
    }
  }
}
```

- Transport: stdio (stdin/stdout). `--verbose` vai para stderr.
- Um processo MCP = um tenant; resolvido no boot via `--tenant` /
  `MOVIDESK_TENANT` / tenant atual em `~/.movidesk/config.yaml`.
- Escopo v1: só leitura. 22 tools (`tickets_list`, `tickets_get`,
  `tickets_timeline`, `persons_list`, `services_list`, `contracts_list`,
  `kb_article_get`, `activities_list`, `surveys_responses_list`,
  `query`, …).
- Resources: `movidesk://odata-filter-syntax` (sintaxe completa de filtros),
  `movidesk://server-info` (diagnóstico) e
  `movidesk://customfields-catalog` (catálogo local de custom fields do
  tenant — exibido quando `~/.movidesk/<tenant>/customfields.yaml` existir).
- Defaults de segurança: `all=true` sem `max` limita em 500 linhas para não
  estourar o limite de 10 req/min; respostas acima de 256 KiB são truncadas
  com aviso.

Se o tenant guarda o token em arquivo (em vez do chaveiro do SO), inclua
`MOVIDESK_PASSPHRASE` na chave `env` da configuração do chat app.

### Skill do Claude para o servidor MCP

Para que o Claude acerte filtros, paginação e expansões deste servidor
MCP sem chutar, instale a skill `movidesk-mcp`:

```bash
mkdir -p ~/.claude/skills/movidesk-mcp && \
  curl -fsSL https://raw.githubusercontent.com/jonasandre/movidesk-cli/main/.claude/skills/movidesk-mcp/SKILL.md \
    -o ~/.claude/skills/movidesk-mcp/SKILL.md
```

A skill fica disponível na próxima sessão do Claude Code / Claude
Desktop. Para remover, apague o diretório.

## Convenções

**Token e tenant override** — `--tenant <nome>` ou `MOVIDESK_TENANT`
para mirar um tenant específico; `MOVIDESK_TOKEN` ignora o chaveiro
(útil em CI).

**Usuário padrão** — `--user <id>` ou `MOVIDESK_USER` sobrepõe o
usuário configurado no tenant. Injetado em `createdBy` apenas em
`tickets create` e `tickets actions add` — nunca em updates, nunca
quando o corpo já fornece um.

**Entrada de corpo** — comandos de escrita (`create`, `update`, ...)
aceitam três modos de entrada, em ordem de precedência:

1. `--file <caminho>` — corpo JSON completo.
2. `--from-template <nome>` / `--from-template-file <caminho>` —
   templates reutilizáveis em `~/.movidesk/templates/<nome>.json`.
3. `--set chave=valor` (repetível) — sobrescritas inline; os valores
   são interpretados primeiro como JSON, com fallback para string
   (`--set type=2` → número, `--set tags='["a","b"]'` → array).

**Formatos de saída** — `--output json|table|csv`, mais `--compact`
para JSON em linha única, `--columns id,subject,owner.businessName`
para sobrescrever as colunas table/CSV (dot-paths suportados para
campos aninhados).

**Rate limiting** — toda requisição passa pelo limiter interno de 10
req/min e tenta novamente em 429/5xx respeitando `Retry-After`.
`--no-retry` desativa retentativas pra debug.

**Operações em lote (`tickets bulk-update` / `tickets bulk-close` / `tickets bulk-cancel`)** —
seleciona chamados por `--ids`, `--ids-file` ou `--filter` OData; em
TTY abre seletor com checkbox + busca. Aplica o mesmo PATCH a todos.
`bulk-close` ajusta `status` (padrão `Resolvido`) e cria uma `action`
com `--message` (use `--public` para visível ao cliente). `bulk-cancel`
tem a mesma mecânica, mas padroniza `status=Cancelado` — use para
descartar chamados sem solução (duplicados, abandonados, fora de escopo).
Flags úteis:
`--dry-run` (não chama API), `--force` (pula confirmação),
`--continue-on-error` (não aborta na primeira falha), `--report
caminho.jsonl` (uma linha por chamado: `{id, ok, error, at}`).
`--source live|past|both` escolhe a fonte da listagem — `live`
(`/tickets`, padrão, cobre só os últimos 90 dias), `past`
(`/tickets/past`, arquivados/parados há mais de 90 dias) ou `both`
(mescla as duas, dedup por id). Lotes grandes respeitam o limiter —
100 chamados levam ~10 min.

## Arquivos de configuração

- `~/.movidesk/config.yaml` — lista de tenants, tenant atual,
  padrões. Modo `0600`.
- `~/.movidesk/credentials.enc` — fallback criptografado do token,
  usado apenas quando o chaveiro do SO não está disponível. AES-GCM,
  modo `0600`.
- `~/.movidesk/templates/<nome>.json` — corpos reutilizáveis para
  comandos de escrita.
- `~/.movidesk/<tenant>/customfields.yaml` — catálogo por tenant que
  mapeia rótulos legíveis para ids numéricos de campo personalizado
  (Movidesk não expõe API pública para descobri-los).

Defina `MOVIDESK_HOME` para realocar o diretório — útil em testes e
ambientes de CI efêmeros.

> **Nunca commite esses arquivos.** Eles concedem o mesmo acesso do
> usuário dono.

## Pegadinhas que vale saber de cara

- **PATCH em `/tickets` substitui campos com valor de array** como
  `customFieldValues`. Os helpers tipados de escrita do CLI passam
  por read-merge-patch, então você só descreve a alteração. Se montar
  o corpo manualmente, envie o array completo.
- **O Movidesk não expõe API para listar definições de campos
  personalizados.** Popule o catálogo local a partir dos ids visíveis
  na UI do Movidesk.
- **`tickets list` retorna apenas chamados atualizados nos últimos 90
  dias.** Os mais antigos ficam em `/tickets/past`
  (`tickets past list`).
- **Activities e surveys/responses usam paginação por cursor**, não
  OData (`--limit` / `--starting-after`); use `--all` pra percorrer
  todas as páginas.

## Desenvolvimento

```bash
make build              # compila o binário em ./bin
make test               # com race detector, sem cache
make lint               # golangci-lint
make cover              # resumo de cobertura
make docs               # regenera docs/cli/*.md a partir do Cobra
make run ARGS="auth list"
make release-check      # valida .goreleaser.yaml
make release-snapshot   # dry-run local do goreleaser
```

Para maintainers preparando releases, veja [`docs/RELEASING.md`](./docs/RELEASING.md).

Layout do módulo:

```
cmd/movidesk-cli/      ponto de entrada
cmd/gen-docs/          gerador de docs/cli/
internal/cli/          comandos Cobra
internal/config/       config YAML multi-tenant
internal/auth/         chaveiro + armazenamento de token em arquivo
internal/movidesk/     SDK: HTTP client, rate limiter, retry, OData builder
internal/output/       formatters JSON / table / CSV
internal/version/      metadados de build (definidos via -ldflags)
```

## Contribuindo

Issues e PRs são bem-vindos. Por favor rode `make lint test` antes de
enviar. Veja [CONTRIBUTING.md](./CONTRIBUTING.md) para convenções de
desenvolvimento.

## Licença

MIT — veja [LICENSE](./LICENSE).
