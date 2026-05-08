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

**Fase 0 — Foundation ✅ ENTREGUE**
1. `go mod init github.com/jonasandre/movidesk-cli` ✅
2. Esqueleto Cobra+Viper, comando root `movidesk-cli`, `--version`. ✅
3. `internal/config` lendo/escrevendo YAML. ✅
4. `internal/auth` com keychain + fallback AES-GCM/PBKDF2. ✅
5. Comandos `auth login/switch/list/status/logout/token`. ✅
6. `internal/movidesk/client.go` HTTP base + injeção de token + rate limiter + retry. ✅
7. `internal/output` com 3 formatters (json/table/csv) + flag global `--output`. ✅
8. Testes unitários + httptest pra client. ✅
9. CI GitHub Actions: lint + test + build em PR. ✅
10. GoReleaser config (darwin/linux/windows × amd64/arm64) + Homebrew tap. ✅

**Fase 1 — Tickets ✅ ENTREGUE**
- `tickets list/get/create/update/html/past/merged/attach` ✅
- Auto-paginação (`--all` + `--max`) ✅
- Template engine (`--from-template`, `--from-template-file`, `--set k=v`) ✅
- Tabela default columns por recurso (`internal/output/columns.go`) ✅
- 12 E2E tests via httptest ✅
- README docs com exemplos por verbo ✅

**Fase 1.5 — Schema completion + ticket collections (3 dias) — PRÓXIMA**

Motivação: o struct `Ticket` em `internal/movidesk/tickets/tickets.go` cobre ~22 dos ~80 campos do schema Movidesk. Saída via `--output json` preserva tudo (round-trip-safe), mas o SDK Go é incompleto e há gaps de ergonomia + um bug no `--all + table/csv`.

**1.5.A — Bug fix: paginate + table/csv (1h)**
- `internal/output/path.go:70` `asRows()` não trata `[]json.RawMessage`. Hoje `tickets list --all --output table` imprime "(no rows)".
- Adicionar case `[]json.RawMessage` que decodifica cada item em `map[string]any`.
- Teste de regressão em `internal/output/output_test.go` cobrindo o caso.
- Teste E2E em `internal/cli/tickets_e2e_test.go` rodando `tickets list --all --output table` e validando linhas.

**1.5.B — Schema completo no SDK (1-1.5 dias)**

Expandir `internal/movidesk/tickets/tickets.go` com 100% dos campos doc'd. Novo arquivo `tickets_types.go` pra organizar nested:

| Tipo Go | JSON | Notas |
|---|---|---|
| `Ticket` (existente, expandir) | — | adiciona: `serviceFull`, `serviceFirstLevelId`, `serviceSecondLevel`, `serviceThirdLevel`, `contactForm`, `cc`, `reopenedIn`, `lifetimeWorkingTime`, `stoppedTime`, `stoppedTimeWorkingTime`, `resolvedInFirstCall`, `chatWidget`, `chatGroup`, `chatTalkTime`, `chatWaitingTime`, `sequence`, `slaAgreement`, `slaAgreementRule`, `slaSolutionTime`, `slaResponseTime`, `slaSolutionChangedByUser`, `slaSolutionChangedBy`, `slaSolutionDate`, `slaSolutionDateIsPaused`, `slaResponseDate`, `slaRealResponseDate`, `jiraIssueKey`, `redmineIssueId`, `clients`, `actions`, `parentTickets`, `childrenTickets`, `ownerHistories`, `statusHistories`, `customFieldValues`, `assets` |
| `Client` | `ticket.clients[n]` | id, businessName, email, phone, personType, profileType, isDeleted, organization (Person) |
| `Action` | `ticket.actions[n]` | id, type, origin, description, htmlDescription, status, justification, createdDate, createdBy (Person), isDeleted, timeAppointments[], expenses[], attachments[], tags[] |
| `Attachment` | `action.attachments[n]` | fileName, path, createdDate, createdBy, isDeleted |
| `TimeAppointment` | `action.timeAppointments[n]` | id, activity, date, periodStart, periodEnd, workTime, workTypeName, accountable (Person), isDeleted |
| `Expense` | `action.expenses[n]` | id, type, value, serviceReport, createdBy, createdDate, isDeleted |
| `ParentChild` | `parentTickets[n]` / `childrenTickets[n]` | id, isDeleted |
| `OwnerHistory` | `ownerHistories[n]` | ownerTeam, owner (Person), changedBy, changedDate, permanencyTime, permanencyTimeFullTime, permanencyTimeWorkingTime |
| `StatusHistory` | `statusHistories[n]` | status, justification, changedBy (Person), changedDate, permanencyTime |
| `CustomFieldValue` | `ticket.customFieldValues[n]` | customFieldId, customFieldRuleId, line, column, value, items[] (objetos com id/personId/clientId/team/customFieldItem) |
| `Asset` | `ticket.assets[n]` | id, name, label |
| `Person` (existente, expandir) | — | manter campos atuais; adicionar lookup helpers se útil |

**Round-trip preservation:** manter `Extra json.RawMessage` em `Ticket`, `Action`, `CustomFieldValue` pra não quebrar rotas onde Movidesk adicione campos sem aviso. Decoder injeta `Extra` no UnmarshalJSON custom (não na struct tag), pra serialização `MarshalJSON` round-trip via `Extra` quando presente.

Testes: `tickets_types_test.go` com fixture JSON realista (sample completo do doc) cobrindo unmarshal → marshal → unmarshal idempotente.

**1.5.C — Subcomandos de coleção (1.5 dias)**

Implementar em `internal/cli/tickets_collections.go` + helpers no SDK.

> ⚠️ **Trap do PATCH `/tickets`:** o doc é explícito (em `customFieldValues`) — *"Os campos que estiverem na base de dados e não forem enviados no corpo da requisição serão excluídos."* Para `customFieldValues` (e potencialmente outras arrays), todo write precisa ler o ticket, mergear com a mudança e enviar a lista completa. O SDK encapsula isso como **read-merge-patch** — comandos de write nunca expõem essa armadilha.

**Coleções de leitura:**

| Comando | Mecanismo | Output default |
|---|---|---|
| `tickets actions list <id>` | `GET /tickets?id=<id>&$expand=actions` → extrai `actions[]` | table cols `id, type, origin, createdDate, createdBy.businessName, isDeleted` |
| `tickets actions get <id> --action-id N` | mesmo + filtro local | JSON da `Action` completa |
| `tickets actions html <id> --action-id N` | reusa `tickets html <id> --action-id N` (já existe) | description HTML |
| `tickets clients list <id>` | `$expand=clients` → extrai | table cols `id, businessName, email, personType, profileType, isDeleted` |
| `tickets relations <id>` | `$expand=parentTickets,childrenTickets` | duas seções: "Parents:" e "Children:", table |
| `tickets timeline <id>` | merge ordenado por data: `actions[].createdDate` + `statusHistories[].changedDate` + `ownerHistories[].changedDate` | table cols `when, kind, actor, summary` |
| `tickets assets list <id>` | `$expand=assets` → extrai | table cols `id, name, label, categoryFirstLevel` |
| `tickets histories list <id>` | `$expand=ownerHistories,statusHistories` | table com seções "Owner:" e "Status:" |

**Coleções de escrita (actions):**

| Comando | Body | Notas |
|---|---|---|
| `tickets actions add <id>` | PATCH com `{actions:[{type, description, ...}]}` (sem id) | flags: `--type 1\|2`, `--description "text"`, `--description-file path.html`, `--public` (alias `--type 2`), `--internal` (alias `--type 1`), `--tag` (repeatable). Action recém-criada: ler ticket → append → PATCH. |
| `tickets actions update <id> --action-id N` | PATCH com `{actions:[{id:N, ...}]}` | mantém demais ações intactas; só envia a alterada |
| `tickets actions delete <id> --action-id N` | PATCH com `{actions:[{id:N, isDeleted:true}]}` | Movidesk usa soft-delete por flag |
| `tickets actions attach <id> --action-id N --file path` | reusa `tickets attach` (já existe) | sem código novo; só apelido |

`internal/output/columns.go` ganha:
```go
"tickets.actions":     {"id", "type", "origin", "createdDate", "createdBy.businessName", "isDeleted"},
"tickets.clients":     {"id", "businessName", "email", "personType", "profileType", "isDeleted"},
"tickets.relations":   {"id", "isDeleted"},
"tickets.timeline":    {"when", "kind", "actor", "summary"},
"tickets.assets":      {"id", "name", "label", "categoryFirstLevel"},
"tickets.histories":   {"changedDate", "kind", "actor", "from", "to", "permanencyTime"},
"tickets.customfields":{"customFieldId", "customFieldRuleId", "line", "value", "items"},
```

**1.5.C.bis — Custom fields (write é o caso mais delicado)**

Schema em `ticket.customFieldValues[n]` (do doc):

| Campo | Tipo | Quando usar |
|---|---|---|
| `customFieldId` (int, **req**) | id numérico do campo | sempre — obtido na UI Movidesk (Painel → Campos adicionais) |
| `customFieldRuleId` (int, **req**) | id da regra de exibição | sempre — obtido na UI Movidesk |
| `line` (int, **req**) | número da linha | `1` quando regra não permite múltiplas linhas; `>1` para grids |
| `value` (string) | valor texto | tipos: texto-uma-linha, texto-multilinha, HTML, regex, numérico, data, hora, datetime, email, telefone, URL |
| `items[]` (array) | lista de itens | tipos: lista-de-valores, lista-de-pessoas, lista-de-clientes, lista-de-agentes, seleção-múltipla, seleção-única |

Cada `items[n]`:

| Sub-campo | Quando |
|---|---|
| `customFieldItem` (string) | lista-de-valores, seleção-única, seleção-múltipla |
| `personId` (string) | lista-de-pessoas |
| `clientId` (string) | lista-de-clientes |
| `team` (string) | lista-de-agentes (equipes) |

**Formatação especial:**
- Datas: `YYYY-MM-DDThh:MM:ss.000Z` em UTC.
- Hora-only: prefixar data fixa `1991-01-01`.
- Numérico: formato brasileiro com vírgula decimal, ex: `"1.530,75"`.

**Obstáculo de discovery:** a Movidesk **não tem API pública** que liste definições de campos adicionais (id, label, tipo, regra, options). Só dá pra ver na UI web. Para mitigar:

→ **Catálogo local** em `~/.movidesk/<tenant>/customfields.yaml`, gerenciado por comando, mapeando label → ids:

```yaml
fields:
  Severidade:
    id: 125529
    rule_id: 1
    type: list-of-values        # text|number|date|time|datetime|email|phone|url|html|regex
                                # | list-of-values | list-of-persons | list-of-clients
                                # | list-of-agents | single-select | multi-select
    options: [Baixa, Média, Alta, Crítica]
  Data de Implantação:
    id: 125530
    rule_id: 2
    type: date
```

Após cadastrar, o usuário usa `--field-label` em vez de `--field id`.

**Comandos de custom fields:**

| Comando | Mecanismo | Notas |
|---|---|---|
| `tickets customfields show <id>` | `GET /tickets?id=<id>&$expand=customFieldValues` | render como tabela. Se catálogo presente, resolve `customFieldId → label` na coluna |
| `tickets customfields set <id>` (write) | **read-merge-patch**: GET ticket com `$expand=customFieldValues`, mescla nova entrada (mesma `customFieldId+ruleId+line`), PATCH com lista completa | flags: `--field N` ou `--field-label "X"` (resolve via catálogo); `--rule N` (default catálogo); `--line N` (default 1); `--value "x"` ou `--item "x"` (repeatable) ou `--item-person ID` ou `--item-client ID` ou `--item-team NAME` (repeatable). Se catálogo conhece tipo, valida que flags batem |
| `tickets customfields clear <id> --field N [--rule N --line N]` | **read-merge-patch**: GET, omite a entrada da lista, PATCH | sem `--line` remove todas as linhas daquele field |
| `tickets customfields catalog list` | `~/.movidesk/<tenant>/customfields.yaml` | local-only |
| `tickets customfields catalog add --label X --field N --rule N --type T [--options "a,b"]` | edita YAML | local-only |
| `tickets customfields catalog remove --label X` | edita YAML | local-only |
| `tickets customfields catalog import --file path.yaml` | merge externo | útil pra distribuir catálogo padronizado em equipe |

> ℹ️ Movidesk tem outro endpoint, `/ticketCustomFieldValue/{InsertValues\|UpdateValues\|DeleteValues}`, que **só** gerencia opções (pool de valores) de campos do tipo lista. Não confundir com setar valor em ticket. Vai pro `customfields options` na **Fase 3** (`api de manutenção e manipulação de campos adicionais`):
> ```
> customfields options add --field N --values "Opção A,Opção B"
> customfields options rename --field N --from "X" --to "Y"
> customfields options remove --field N --values "Opção C"
> ```

**Helpers no SDK** (`internal/movidesk/tickets/customfields.go`):
```go
func (s *Service) ReadMergePatchCustomField(ctx, ticketID, change CustomFieldValue) ([]byte, error)
func (s *Service) ReadMergePatchClearCustomField(ctx, ticketID, fieldID, ruleID, line int) ([]byte, error)
```
Ambos garantem semântica correta sem expor o trap. Comando `--no-merge` opt-in pra emergências.

`Options.Resource` adicionar: `tickets.actions`, `tickets.clients`, `tickets.timeline`, `tickets.assets`, `tickets.histories`, `tickets.customfields`, `tickets.relations`.

**1.5.D — Default columns enriquecidos (30min)**

Ampliar `defaultColumns` em `internal/output/columns.go`:
```go
"tickets": {"id", "protocol", "subject", "status", "urgency", "owner.businessName", "ownerTeam", "slaSolutionDate", "lastUpdate"},
"tickets.actions":      {"id", "type", "origin", "createdDate", "createdBy.businessName"},
"tickets.clients":      {"id", "businessName", "email", "personType"},
"tickets.timeline":     {"when", "kind", "actor", "summary"},
"tickets.customfields": {"customFieldId", "line", "column", "value"},
"tickets.past":         {"id", "subject", "status", "createdDate", "lastUpdate"},
```

**1.5.E — Doc + smoke (1h)**

- README: nova seção "Tickets — coleções" com exemplos `tickets actions list/add/update/delete`, `tickets clients list`, `tickets relations`, `tickets timeline`, `tickets histories list`, `tickets assets list`.
- README: seção "Custom fields" cobrindo:
  - Trap do PATCH (read-merge-patch automático).
  - Como popular `~/.movidesk/<tenant>/customfields.yaml` via `catalog add`.
  - Tabela de tipos (lista-de-valores ↔ `--item`, lista-de-pessoas ↔ `--item-person`, etc.).
  - Convenções de formato (UTC, Brazilian numeric).
- CONTRIBUTING: nota sobre nunca PATCH cru em `customFieldValues` (sempre via helper read-merge-patch).
- E2E test cada subcomando contra httptest mock incluindo cenário read-merge-patch.

**Verificação 1.5 — comandos novos:**
```bash
make test                                                             # tudo verde, race
./bin/movidesk-cli tickets list --all --output table                  # bug fix valida (sem "(no rows)")
./bin/movidesk-cli tickets list --output table                        # cols enriquecidas: protocol, urgency, ownerTeam, slaSolutionDate
./bin/movidesk-cli tickets actions list 1
./bin/movidesk-cli tickets actions add 1 --type 1 --description "Internal note"
./bin/movidesk-cli tickets actions update 1 --action-id 5 --description "Edited"
./bin/movidesk-cli tickets actions delete 1 --action-id 5
./bin/movidesk-cli tickets clients list 1
./bin/movidesk-cli tickets relations 1
./bin/movidesk-cli tickets timeline 1 --output table
./bin/movidesk-cli tickets histories list 1
./bin/movidesk-cli tickets assets list 1
./bin/movidesk-cli tickets customfields catalog add --label "Severidade" --field 125529 --rule 1 --type list-of-values --options "Baixa,Média,Alta,Crítica"
./bin/movidesk-cli tickets customfields show 1
./bin/movidesk-cli tickets customfields set 1 --field-label "Severidade" --item "Alta"
./bin/movidesk-cli tickets customfields clear 1 --field-label "Severidade"
```

**Fase 2 — Pessoas + Serviços ✅ ENTREGUE**
- `persons list/get/create/update/delete` ✅
- `persons customfields show/set/clear` (read-merge-patch, compartilha catálogo de tickets) ✅
- `services list/get/create/update/delete` ✅
- Round-trip safe via `Extra json.RawMessage` em `persons.Person` e `services.Service` ✅
- E2E tests + SDK tests todos verde com `-race -count=1` ✅
- Default columns enriquecidos (persons: 7 cols incl. corporateName/accessProfile; services: 7 cols incl. parent/visibility) ✅
- README seções "Persons", "Person custom fields", "Services" ✅
- Confirmação obrigatória pra `delete` (`--force` pra skip; refusa em pipe sem `--force`) ✅

**Fase 3 — Restantes ✅ ENTREGUE**
- `activities` (CRUD + add-teams, cursor pagination) ✅
- `contracts` (CRUD) + `contracts consumption list` (period filter requires name) ✅
- `surveys questions list/get` (read-only) ✅
- `surveys responses list` (cursor pagination) ✅
- `kb articles get <id>` (single-article only — no public list endpoint) ✅
- `telephony queue/nonqueue/made-call-link` cobrindo todos os asterisk_* ✅
- `customfields options add/rename/remove` (`/ticketCustomFieldValue/{Insert,Update,Delete}Values`) ✅
- `query <path>` escape hatch (GET/DELETE com auto-paginação) ✅
- Default columns por recurso (activities, contracts, contracts.consumption, surveys.questions, surveys.responses, articles) ✅
- README com seções por família + cobertura completa ✅

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

Fases 0 e 1 entregues. **Próximo:** executar Fase 1.5 (schema completion + collections de tickets) — ordem A → B → C → D → E. Depois seguir Fase 2 (Persons/Services).

Estimativa restante para v1.0: **5-7 semanas** em ritmo de meio período.

## Apêndice — gap analysis dos tickets (ago/2026)

Estado atual de `internal/movidesk/tickets/tickets.go` (entregue na Fase 1):

- ✅ Round-trip safe via `--output json`: o CLI faz `json.Unmarshal` em `any` e re-emite, então usuário não perde dados na linha de comando.
- ⚠️ Struct `Ticket` cobre ~22 de ~80 campos. Quem usa o pacote `tickets` como lib Go perde tipagem em: cluster SLA (10 campos), coleções (`clients`, `actions`, `parentTickets`, `childrenTickets`, `ownerHistories`, `statusHistories`, `customFieldValues`, `assets`), serviços encadeados (`serviceFull`, `serviceSecondLevel`, `serviceThirdLevel`, `serviceFirstLevelId`), cluster chat (5 campos), working time (3 campos), `jiraIssueKey`/`redmineIssueId`, `cc`, `reopenedIn`, `sequence`.
- 🐛 `internal/output/path.go:70` `asRows()` não trata `[]json.RawMessage` (retorno de `Paginate()`). `tickets list --all --output table` imprime "(no rows)" silenciosamente.
- ❌ Sem subcomandos de coleção (`tickets actions list <id>`, `tickets clients list <id>`, `tickets customfields show/set <id>`, `tickets relations <id>`, `tickets timeline <id>`).
- ❌ `defaultColumns["tickets"]` rasos (5 colunas); não inclui urgency, ownerTeam, slaSolutionDate, baseStatus, protocol.

Fase 1.5 fecha esses pontos.
