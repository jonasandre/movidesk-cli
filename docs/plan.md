# Movidesk CLI — Plano de Projeto

## Contexto

A Movidesk expõe ~14 APIs REST públicas (Tickets, Pessoas, Serviços, Anexos, Atividades, Contratos, Pesquisa de Satisfação, Telefonia, KB, Tickets Past/Merged, Campos Adicionais). Cada uma usa o mesmo padrão (base `https://api.movidesk.com/public/v1`, autenticação por `?token=...`, OData para queries: `$filter`/`$select`/`$expand`/`$orderby`/`$top`/`$skip` com lambdas `any/all`, rate limit de 10 req/min, 429 com `retry-after`).

Hoje o consumo dessa API depende de scripts custom ou Postman, sem padronização nem suporte multi-tenant. Este projeto entrega um CLI único (`movidesk-cli`) que cobre todas as APIs, lida com auth multi-tenant de forma segura (keychain OS), formata saída em JSON/tabela/CSV, e respeita rate limits automaticamente. Distribuído como binário Go via GitHub Releases + Homebrew tap, MIT.

Resultado pretendido: substituir scripts ad-hoc por uma CLI consistente, automatizar workflows de operações e integrações com pipelines (CI/data export/migrações), e dar uma base extensível pra novas APIs Movidesk sem retrabalho.

## Decisões locked (fase de discovery)

| Decisão | Valor |
|---|---|
| Linguagem | Go (1.22+) |
| Framework CLI | spf13/cobra + spf13/viper |
| HTTP client | net/http nativo + middleware customizado (retry/rate limit) |
| Keychain | zalando/go-keyring |
| YAML config | gopkg.in/yaml.v3 |
| Output table | jedib0t/go-pretty |
| Output CSV | encoding/csv (stdlib) |
| Testing | stdlib testing + testify (assert) + httptest |
| Lint | golangci-lint |
| Release | GoReleaser → GitHub Releases + Homebrew tap |
| Binário | `movidesk-cli` |
| Licença | MIT |
| Escopo MVP | Cobertura COMPLETA de todas as APIs listadas |

## Arquitetura — camadas

```
cmd/movidesk-cli/main.go              — entry point
internal/cli/                          — comandos Cobra (auth, tickets, persons, ...)
internal/config/                       — Viper + multi-tenant + keychain
internal/auth/                         — login/logout/switch/status, keychain wrapper
internal/movidesk/                     — SDK interno (HTTP client + recursos)
  client.go                            — cliente HTTP base, injeção de token
  ratelimit.go                         — token bucket 10 req/min + retry-after honor
  retry.go                             — backoff exponencial pra 429/5xx
  odata/                               — query builder ($filter/$select/$expand/$orderby/$top/$skip)
  tickets/                             — service por recurso (one file por recurso)
  persons/
  services/
  attachments/                         — multipart pra /ticketFileUpload
  activities/
  contracts/
  surveys/
  telephony/
  knowledgebase/
  customfields/
internal/output/                       — formatters (json, table, csv)
  formatter.go                         — interface comum
  json.go                              — pretty + compact
  table.go                             — colunas configuráveis por recurso
  csv.go
internal/version/                      — version, commit, build date (ldflags)
```

**Princípio:** SDK interno (`internal/movidesk`) é puro Go, sem deps de Cobra/Viper. CLI é fina camada que mapeia flags → SDK → output. Permite reuso futuro como lib em pipelines Go.

## Multi-tenant + auth — design

**Storage:**
- `~/.movidesk/config.yaml` (chmod 0600): lista de tenants (nome, label, default flag, output prefs), tenant atual.
- Tokens **nunca em arquivo** (em ambientes com keyring): cada tenant tem token salvo em keychain com chave `movidesk-cli:<tenant>`.
- Fallback Linux headless (sem libsecret): aviso + opt-in pra arquivo `~/.movidesk/credentials` chmod 0600 cifrado com AES-GCM derivando chave de passphrase (PBKDF2). Preferir Recommendation: instalar gnome-keyring/kwallet.

**Override por env (CI):** `MOVIDESK_TOKEN` + `MOVIDESK_TENANT` curto-circuitam config/keychain.

**Comandos auth:**
```
movidesk-cli auth login --tenant <name> [--label "Acme Prod"] [--make-default]
  → prompt secreto pro token (term.ReadPassword), valida via GET /persons?$top=1, salva no keychain
movidesk-cli auth switch <name>
movidesk-cli auth list                  # mostra tenants, * no atual, NUNCA mostra token
movidesk-cli auth status [--tenant x]   # valida token, mostra rate limit info
movidesk-cli auth logout [--tenant x] [--all]
movidesk-cli auth token --tenant x      # imprime token em stdout (uso explícito, ex: piping)
```

`config.yaml` schema:
```yaml
current_tenant: acme
defaults:
  output: json     # json|table|csv
  page_size: 100
tenants:
  acme:
    label: Acme Prod
    base_url: https://api.movidesk.com/public/v1   # override só se sandbox
    output: table  # override por tenant
  beta:
    label: Beta Dev
```

## Mapeamento de comandos CLI → endpoints Movidesk

Padrão geral: `movidesk-cli <recurso> <ação> [flags]`. Flags comuns: `--filter` (OData passthrough), `--select`, `--expand`, `--orderby`, `--top`, `--skip`, `--all` (auto-paginação), `--output json|table|csv`, `--tenant`.

| Comando CLI | Endpoint Movidesk | Métodos |
|---|---|---|
| `tickets list` | `GET /tickets` | OData query |
| `tickets get <id\|--protocol>` | `GET /tickets?id=` | + `--include-deleted` |
| `tickets create -f file.json` / `--from-template` | `POST /tickets` | |
| `tickets update <id> -f patch.json` / `--set field=val` | `PATCH /tickets?id=` | |
| `tickets html <id> [--action-id N]` | `GET /tickets/htmldescription` | |
| `tickets past list` | `GET /tickets/past` | tickets >90 dias |
| `tickets merged list` | `GET /tickets/merged` | |
| `tickets attach <id> --action-id N --file path` | `POST /ticketFileUpload` | multipart |
| `persons list / get / create / update / delete` | `/persons` | GET/POST/PATCH/DELETE |
| `services list / get / create / update / delete` | `/services` | |
| `activities list / get / create / update / delete` | `/activities` | |
| `contracts list / get / create / update` | `/contracts` | |
| `contracts consumption list` | `GET /contracts/.../hourConsumption` | |
| `surveys questions list / get / create / update` | `/surveys/questions` | |
| `surveys responses list` | `/surveys/responses` | |
| `telephony queue ...` | `/telephony` (com fila) | |
| `telephony nonqueue ...` | `/telephony` (sem fila) | |
| `kb articles list / get` | `/articles` | KB consulta |
| `customfields list / get / create / update / delete` | `/customFieldValues` (manipulação) | |
| `query <resource> --filter ... --raw` | qualquer endpoint | escape hatch p/ OData ad-hoc |

**Auto-paginação:** `--all` chama em loop incrementando `$skip` até resposta vazia, respeitando rate limit.

**Templates:** `tickets create --from-template <name>` carrega `~/.movidesk/templates/<name>.json` com placeholders pra produtividade.

## Rate limiting + retry

- Token bucket interno: 10 tokens/min, blocking acquire antes de cada request.
- Em 429: ler header `retry-after`, sleep + retry (max 3 tentativas), cap 5min total.
- Em 5xx: backoff exponencial 1s → 2s → 4s, max 3 tentativas.
- Em 401/403: fail fast, sugerir `auth status`.
- Flag `--no-retry` pra debug.

## Output formatters

- `json` (default): `json.MarshalIndent` 2 spaces. Flag `--compact` desabilita.
- `table`: colunas pré-definidas por recurso em `internal/output/columns.go`. `--columns id,subject,status` override.
- `csv`: header + rows. Suporta nested via dot-path (`owner.businessName`).
- Sem cor por default em pipe (detect `isatty`). Flag `--color=auto|always|never`.

## Roadmap por fases

**Fase 0 — Foundation (1-2 semanas)**
1. `go mod init github.com/jonasandre/movidesk-cli`
2. Esqueleto Cobra+Viper, comando root `movidesk-cli`, `--version`.
3. `internal/config` lendo/escrevendo YAML.
4. `internal/auth` com keychain + fallback.
5. Comandos `auth login/switch/list/status/logout/token`.
6. `internal/movidesk/client.go` HTTP base + injeção de token + rate limiter + retry.
7. `internal/output` com 3 formatters + flag global `--output`.
8. Testes unitários (config, auth mock, ratelimit, formatters) + httptest pra client.
9. CI GitHub Actions: lint + test em PR. Codecov opcional.
10. GoReleaser config + tag v0.1.0 trigger build (sem release público ainda).

**Fase 1 — Tickets completo (1-2 semanas)**
- `tickets list/get/create/update/html/past/merged/attach`
- Auto-paginação (`--all`)
- Template engine pra create
- Tabela default columns: `id, subject, status, owner.businessName, lastUpdate`
- Doc + exemplos no README

**Fase 2 — Pessoas + Serviços (1 semana)**
- `persons` CRUD completo + delete
- `services` CRUD
- Suporte a campos adicionais embutidos no body (`customFieldValues`)

**Fase 3 — Restantes (2-3 semanas)**
- `activities`, `contracts` (+ consumption), `surveys` (questions/responses), `telephony` (com/sem fila), `kb articles`, `customfields` manipulação
- Comando `query <resource>` escape hatch

**Fase 4 — Release v1.0 (1 semana)**
- README completo, doc por comando (`docs/` gerada via `cobra-cli completion docs`)
- Brew tap repo `homebrew-movidesk`
- GoReleaser pipeline release-on-tag → assina checksums, publica binários, atualiza tap
- LICENSE MIT, CONTRIBUTING.md, CHANGELOG.md
- Anúncio + screencast curto

**Fase 5+ — Pós-MVP (futuro)**
- Webhooks listener local (`movidesk-cli webhook listen --port 8080`)
- Watch mode (`tickets watch --filter` polling a cada N min)
- Bulk import CSV → `persons import --csv file.csv`
- Plugin system pra recursos custom

## Arquivos críticos a criar

| Arquivo | Propósito |
|---|---|
| `cmd/movidesk-cli/main.go` | Entry, chama `cli.Execute()` |
| `internal/cli/root.go` | Cobra root cmd, persistent flags (`--tenant`, `--output`, `-v`) |
| `internal/cli/auth.go` | Subcomandos auth |
| `internal/cli/tickets.go` (e um por recurso) | Comandos por recurso |
| `internal/config/config.go` | Carrega/salva YAML, resolve tenant atual |
| `internal/auth/keyring.go` | Wrapper go-keyring + fallback |
| `internal/movidesk/client.go` | HTTP, injeção de token, headers |
| `internal/movidesk/ratelimit.go` | Token bucket 10/min |
| `internal/movidesk/retry.go` | Backoff + retry-after |
| `internal/movidesk/odata/builder.go` | Query string OData type-safe |
| `internal/movidesk/tickets/tickets.go` | Tipos + métodos por recurso |
| `internal/output/formatter.go` | Interface comum |
| `internal/output/{json,table,csv}.go` | Implementações |
| `.goreleaser.yaml` | Build matriz darwin/linux/windows × amd64/arm64 |
| `.github/workflows/{ci,release}.yml` | Lint+test em PR; release em tag |
| `Makefile` | `make build/test/lint/run` |
| `README.md` | Quickstart, exemplos, comandos |
| `LICENSE` | MIT |

## Bibliotecas externas (locked)

```
github.com/spf13/cobra
github.com/spf13/viper
github.com/zalando/go-keyring
gopkg.in/yaml.v3
github.com/jedib0t/go-pretty/v6
github.com/stretchr/testify
golang.org/x/term       # leitura segura de senha
```

Sem cliente HTTP externo (resty/req): stdlib `net/http` é suficiente. Sem axios-equivalente.

## Verificação end-to-end

**Setup local:**
```bash
git clone <repo> && cd movidesk-cli
make build                    # gera ./bin/movidesk-cli
./bin/movidesk-cli --version
```

**Smoke test:**
```bash
# 1. Auth flow
./bin/movidesk-cli auth login --tenant test
# (cola token de uma conta sandbox Movidesk)
./bin/movidesk-cli auth list
./bin/movidesk-cli auth status

# 2. Tickets
./bin/movidesk-cli tickets list --top 5
./bin/movidesk-cli tickets list --top 5 --output table
./bin/movidesk-cli tickets list --top 5 --output csv > tickets.csv
./bin/movidesk-cli tickets get 1
./bin/movidesk-cli tickets list --filter "createdDate gt 2026-01-01T00:00:00Z" --select id,subject

# 3. Multi-tenant
./bin/movidesk-cli auth login --tenant prod
./bin/movidesk-cli auth switch test
./bin/movidesk-cli tickets list --tenant prod --top 3

# 4. Rate limit (forçar)
for i in {1..15}; do ./bin/movidesk-cli persons list --top 1; done
# → deve ver pause após 10 reqs, retry transparente

# 5. Outras APIs
./bin/movidesk-cli persons list --top 5
./bin/movidesk-cli services list
./bin/movidesk-cli tickets attach 1 --action-id 1 --file ./test.pdf
```

**Testes automatizados:**
```bash
make test            # unit tests, target >70% coverage em internal/movidesk
make lint            # golangci-lint run
make test-integration TENANT=test  # contra sandbox real, opt-in
```

**Release dry-run:**
```bash
goreleaser release --snapshot --clean   # valida config sem publicar
```

## Riscos + mitigações

| Risco | Mitigação |
|---|---|
| API Movidesk muda sem aviso | Testes de integração contra sandbox em schedule semanal |
| Rate limit 10/min muito restritivo pra bulk | Cache local opt-in (`--cache`) + warning antes de operações grandes |
| Token vazado em logs | Nunca logar URL completa; redact token em error messages; never echo input |
| Linux sem keyring | Fallback documentado, warning na primeira execução |
| Movidesk lança CLI oficial | Nome `movidesk-cli` (não `movidesk`) reduz colisão; preparar para coexistir |
| Quebras OData entre versões | `odata` builder isolado, fácil refactor central |

## Próximo passo após aprovação

Iniciar Fase 0 criando estrutura de diretórios, `go mod init`, e esqueleto Cobra. Estimativa total para v1.0 (cobertura completa): **6-9 semanas** em ritmo de meio período.
