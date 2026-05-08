# movidesk-cli

A command-line interface for the [Movidesk](https://www.movidesk.com) public REST API.

- Multi-tenant: keep separate tokens for sandbox/staging/production, switch with one command.
- Tokens stored in the OS keychain (macOS Keychain, Windows Credential Manager, Linux libsecret/kwallet); encrypted file fallback for headless environments.
- OData passthrough on every list endpoint: `--filter`, `--select`, `--expand`, `--orderby`, `--top`, `--skip`.
- Output as JSON (default), human-readable table, or CSV.
- Honors Movidesk's 10 req/min rate limit with automatic backoff and `Retry-After` support.

> Status: **Phase 3 — All resources covered.** The CLI now wraps every API in the
> Movidesk integration menu: `auth`, `tickets` (full schema + collections + custom fields),
> `persons`, `services`, `activities`, `contracts` (+ consumption), `surveys`
> (questions + responses), `kb articles`, `telephony` (queue + nonqueue + made-call-link),
> `customfields options` (option pool), plus a `query` escape hatch for raw OData.
> See [the plan](./docs/plan.md) for the roadmap to v1.0.

## Install

### From source

```bash
go install github.com/jonasandre/movidesk-cli/cmd/movidesk-cli@latest
```

### Pre-built binaries

Download the appropriate archive for your platform from the [Releases](https://github.com/jonasandre/movidesk-cli/releases) page.

### Homebrew (macOS / Linux)

```bash
brew install jonasandre/movidesk/movidesk-cli
```

## Quickstart

```bash
# 1. Add a tenant. The token is read from a hidden prompt and validated
#    against GET /persons?$top=1 before being stored.
movidesk-cli auth login --tenant prod --label "Acme Production" --make-default

# 2. List configured tenants (token never displayed).
movidesk-cli auth list

# 3. Verify everything is wired up.
movidesk-cli auth status
```

## Multiple tenants

```bash
movidesk-cli auth login --tenant sandbox --label "Acme Sandbox"
movidesk-cli auth switch sandbox
movidesk-cli auth list
movidesk-cli auth status --tenant prod        # one-off check without switching
movidesk-cli auth logout --tenant sandbox     # remove just one
movidesk-cli auth logout --all                # nuke everything
```

You can also override the active tenant per-command via `--tenant` or the
`MOVIDESK_TENANT` environment variable. CI pipelines may skip the keychain
entirely by exporting `MOVIDESK_TOKEN`; if both are set, the env var wins.

## Tickets

```bash
# List recent tickets (≤ 90 days). Token is sent automatically.
movidesk-cli tickets list --top 5 --output table

# OData filters and projection.
movidesk-cli tickets list \
  --filter "createdDate gt 2026-01-01T00:00:00Z" \
  --select id,subject,status \
  --orderby "id desc" \
  --top 50

# Auto-pagination: walk every page (respects rate limit; --max caps the total).
movidesk-cli tickets list --filter "status eq 'Novo'" --all --max 1000 --output csv > novos.csv

# Single ticket by id or protocol.
movidesk-cli tickets get 1
movidesk-cli tickets get --protocol MOVI202109000001

# HTML body of a specific action.
movidesk-cli tickets html 1 --action-id 3 --output table

# Tickets older than 90 days live on a separate endpoint.
movidesk-cli tickets past list --filter "createdDate lt 2025-01-01T00:00:00Z" --top 50
movidesk-cli tickets merged list --top 20
```

### Creating tickets

Three input modes — pick whichever fits:

```bash
# 1) JSON body file.
cat > new.json <<'JSON'
{ "type": 2, "subject": "API outage", "category": "Suporte", "createdBy": { "id": "u-123" } }
JSON
movidesk-cli tickets create --file new.json --return-all

# 2) Saved template under ~/.movidesk/templates/<name>.json.
mkdir -p ~/.movidesk/templates
cat > ~/.movidesk/templates/support.json <<'JSON'
{ "type": 2, "category": "Suporte", "urgency": "Média" }
JSON
movidesk-cli tickets create --from-template support \
  --set subject="API outage" --set createdBy='{"id":"u-123"}'

# 3) Plain --set overrides for trivial cases (values are JSON-parsed first).
movidesk-cli tickets create --set type=2 --set subject="Quick test"
```

`--set key=value` accepts JSON values; `--set type=2` is parsed as a number,
`--set tags='["alpha","beta"]'` as an array, anything else as a string.

### Updating tickets

```bash
movidesk-cli tickets update 42 --set subject="New subject"
movidesk-cli tickets update 42 --file patch.json
```

### Attaching files

```bash
movidesk-cli tickets attach 12 --action-id 34 --file ./report.pdf
movidesk-cli tickets attach 12 --action-id 34 --file ./image.png --name "screenshot.png"
```

The action id is required; Movidesk records the upload against an existing
ticket action.

## Tickets — coleções

Each subcommand fetches the ticket with the right `$expand` and surfaces the
nested collection with sensible default columns.

```bash
# Actions of a ticket — read.
movidesk-cli tickets actions list 1
movidesk-cli tickets actions get 1 --action-id 5

# Actions — write. Add a new action; --internal/--public are aliases for --type.
movidesk-cli tickets actions add 1 --internal --description "Triage notes"
movidesk-cli tickets actions add 1 --public  --description-file reply.html

# Edit / soft-delete an action.
movidesk-cli tickets actions update 1 --action-id 5 --description "Edited reply"
movidesk-cli tickets actions delete 1 --action-id 5

# Clients listed on a ticket.
movidesk-cli tickets clients list 1 --output table

# Parent / child tickets.
movidesk-cli tickets relations 1

# Chronological merge of actions, status changes, owner changes.
movidesk-cli tickets timeline 1 --output table

# Linked assets and full ownership/status histories.
movidesk-cli tickets assets list 1
movidesk-cli tickets histories list 1
```

## Custom fields

Movidesk's `PATCH /tickets` deletes any `customFieldValues` entry not
present in the body. The CLI handles that for you with **read-merge-patch**
under the hood — describe the change you want, never the whole list.

There is **no public API** to discover custom field definitions, so we keep
a per-tenant local catalog at `~/.movidesk/<tenant>/customfields.yaml` that
maps human labels to numeric `customFieldId` and `customFieldRuleId` values.
Populate it from the IDs you see in the Movidesk UI.

```bash
# Register a field once per tenant.
movidesk-cli tickets customfields catalog add \
  --label "Severidade" --field 125529 --rule 7 \
  --type list-of-values --options "Baixa,Média,Alta,Crítica"

movidesk-cli tickets customfields catalog list
movidesk-cli tickets customfields catalog remove --label "Severidade"
```

```bash
# Read the values currently on a ticket.
movidesk-cli tickets customfields show 1

# Set a text/number/date value.
movidesk-cli tickets customfields set 1 --field-label "Data de implantação" \
  --value "2026-04-01T13:00:00.000Z"

# Set a list-of-values item.
movidesk-cli tickets customfields set 1 --field-label "Severidade" --item "Alta"

# List of persons / clients / agent-teams.
movidesk-cli tickets customfields set 1 --field 200 --rule 1 --item-person   "u-123"
movidesk-cli tickets customfields set 1 --field 201 --rule 1 --item-client   "c-456"
movidesk-cli tickets customfields set 1 --field 202 --rule 1 --item-team     "Suporte"

# Clear one specific line, or every line of a field.
movidesk-cli tickets customfields clear 1 --field-label "Severidade" --line 1
movidesk-cli tickets customfields clear 1 --field-label "Severidade"
```

**Field type → flag mapping**

| Movidesk field type | CLI flag |
|---|---|
| `text`, `multiline-text`, `html`, `regex`, `email`, `phone`, `url`, `number`, `date`, `time`, `datetime` | `--value "x"` |
| `list-of-values`, `single-select`, `multi-select` | `--item "x"` (repeatable) |
| `list-of-persons` | `--item-person <personId>` |
| `list-of-clients` | `--item-client <clientId>` |
| `list-of-agents` | `--item-team <name>` |

**Format conventions enforced by Movidesk**

- Datetime: `YYYY-MM-DDThh:MM:ss.000Z` (UTC).
- Time only: prepend the fixed date `1991-01-01`.
- Numeric: Brazilian format with comma decimals, e.g. `"1.530,75"`.

> ℹ️ The `/ticketCustomFieldValue/{InsertValues|UpdateValues|DeleteValues}`
> endpoints manage the **option pool** of list-type fields (the dropdown
> values themselves), not values on tickets. Those land in a future
> `customfields options` subcommand — see the plan.

## Persons

The `/persons` endpoint serves agents, clients, companies, and departments —
disambiguated by `personType` (1=Pessoa, 2=Empresa, 4=Departamento) and
`profileType` (1=Agente, 2=Cliente, 3=Both).

```bash
# OData filters apply just like on tickets.
movidesk-cli persons list --filter "isActive eq true and personType eq 1" --top 50
movidesk-cli persons list --select id,businessName,emails --output csv > persons.csv
movidesk-cli persons list --all --output table

# Get a single person by Cod. Ref.
movidesk-cli persons get acme-1234

# Create from JSON, template, or --set overrides (same modes as tickets).
movidesk-cli persons create --file person.json --return-all
movidesk-cli persons create \
  --set personType=1 --set profileType=2 --set isActive=true \
  --set 'businessName=Joe Doe' --set 'cpfCnpj=01234567890'

# Update partial fields. Note: array fields like `addresses`/`emails`/`teams`
# are replaced by what you send.
movidesk-cli persons update acme-1234 --set 'role=Manager'

# Delete (irreversible, prompts unless --force).
movidesk-cli persons delete acme-1234 --force
```

### Person custom fields

Same read-merge-patch model as tickets, sharing the per-tenant catalog at
`~/.movidesk/<tenant>/customfields.yaml`. Register a field once via
`tickets customfields catalog add` and use the label here too.

```bash
movidesk-cli persons customfields show acme-1234
movidesk-cli persons customfields set  acme-1234 --field-label "Squad" --item "Platform"
movidesk-cli persons customfields set  acme-1234 --field 200 --rule 1 --item-team "Suporte"
movidesk-cli persons customfields clear acme-1234 --field-label "Squad"
```

## Services

```bash
movidesk-cli services list --filter "isActive eq false" --orderby "id desc" --top 100
movidesk-cli services get 5712
movidesk-cli services list --all --output csv > services.csv

# Create a service.
movidesk-cli services create \
  --set 'name=Suporte Avançado' \
  --set isActive=true \
  --set serviceForTicketType=2 \
  --set isVisible=3 --set allowSelection=3 \
  --set allowFinishTicket=true \
  --set allowAllCategories=false \
  --set 'categories=["Problema","Sugestão"]'

# Update — array fields like `categories` are fully replaced by what you send.
movidesk-cli services update 5712 --set 'categories=["Problema"]'

# Delete (irreversible).
movidesk-cli services delete 5712 --force
```

> ℹ️ When updating a parent service that has children, Movidesk keeps the
> existing tree structure intact unless you explicitly set `parentServiceId`
> on each child. The CLI does not auto-cascade.

## Activities

`/activity` uses cursor-based pagination (`limit`/`startingAfter`) instead of OData.

```bash
movidesk-cli activities list --name "Triage" --limit 50
movidesk-cli activities list --all --max 500
movidesk-cli activities get 12
movidesk-cli activities create --set name=Triage --set isActive=true --set isAllowsAllTeams=false
movidesk-cli activities update 12 --set name="Triage v2"
movidesk-cli activities add-teams 12 --team Suporte --team Comercial
movidesk-cli activities delete 12 --force
```

## Contracts

```bash
# Time agreements (cadastro de contratos de horas).
movidesk-cli contracts list --filter "isActive eq true"
movidesk-cli contracts get 1
movidesk-cli contracts create --file contract.json --return-all
movidesk-cli contracts update 1 --set baseAmount=1500.00
movidesk-cli contracts delete 1 --force

# Consumption rows. When filtering by period, name is required by Movidesk.
movidesk-cli contracts consumption list \
  --filter "name eq 'Default' and period gt 2026-01-01T00:00:00Z" \
  --output csv > consumption.csv
```

## Surveys

```bash
movidesk-cli surveys questions list                       # all questions
movidesk-cli surveys questions list --type 3              # NPS only
movidesk-cli surveys questions get QWMv

movidesk-cli surveys responses list --limit 100
movidesk-cli surveys responses list --all --max 5000 --output csv > responses.csv
```

## Knowledge base

The public KB API exposes single-article reads only. You must already know
the article id (no list endpoint).

```bash
movidesk-cli kb articles get 19040
```

## Telephony

These commands fire call events at Movidesk so a phone system integration
can attach calls to tickets. Two flavors:

```bash
# Queue-controlled (POST). event ∈ receivedCall|transferedCall|completedCall|lostCall|canceledCall.
movidesk-cli telephony queue --event receivedCall \
  --set id=call-abc --set queueId=1 --set clientNumber="+55 47 9999-9999" \
  --set callDate=2026-04-01T13:00

# No-queue (GET). event ∈ startTransferedCall|completedCall|startCanceledCall.
movidesk-cli telephony nonqueue --event startTransferedCall \
  --param id=call-abc --param branchLine=1001

# Attach a recording link to an outbound call.
movidesk-cli telephony made-call-link --set id=call-abc --set link=https://recordings/x
```

## Custom field option pool

These commands manage the **dropdown options** of list-type custom fields
tenant-wide — they are NOT for setting a value on a specific ticket or
person (use `tickets customfields set` / `persons customfields set`
for that).

```bash
movidesk-cli customfields options add    --field 125529 --value "Baixa" --value "Alta"
movidesk-cli customfields options rename --field 125529 --pair "Baixa=Pequena" --pair "Alta=Crítica"
movidesk-cli customfields options remove --field 125529 --value "Pequena"
```

## Query escape hatch

Need to call an endpoint not covered by a typed subcommand? Use `query`.
GET-only (and DELETE for cleanup); writes go through the typed subcommands.

```bash
movidesk-cli query /tickets --filter "id eq 1"
movidesk-cli query /persons --select id,businessName --top 5 --all --max 1000
movidesk-cli query /someNewEndpoint --param foo=bar
```

## Output formats

```bash
# JSON (default), pretty.
movidesk-cli ... -o json

# JSON, single line — friendly to jq.
movidesk-cli ... -o json --compact | jq '.[0].id'

# Aligned table for terminal use. Columns are picked automatically; override
# them with --columns id,subject,owner.businessName.
movidesk-cli ... -o table

# CSV for spreadsheets and pipelines.
movidesk-cli ... -o csv > out.csv
```

`--columns` accepts dot-paths (`owner.businessName`) to dig into nested objects.

## Rate limiting

Movidesk enforces 10 requests per minute. The CLI ships an in-process limiter
sized to that ceiling and automatically retries `429` responses, honoring the
`Retry-After` header. `5xx` errors back off exponentially up to three
attempts. Use `--no-retry` to disable retry while debugging.

## Configuration files

- `~/.movidesk/config.yaml` — tenant list, current tenant, default output. Mode `0600`.
- `~/.movidesk/credentials.enc` — only created on Linux without a working keychain. AES-GCM encrypted, mode `0600`.

Override the location with `MOVIDESK_HOME` (handy for tests and ephemeral CI environments).

> **Never commit these files.** They grant the same access as the user that owns them.

## Development

```bash
make build          # build binary into ./bin
make test           # race-enabled, no cache
make lint           # golangci-lint
make cover          # coverage summary
make run ARGS="auth list"
make release-snapshot   # local goreleaser dry-run
```

The Go module layout:

```
cmd/movidesk-cli/      entry point
internal/cli/          Cobra commands
internal/config/       multi-tenant YAML config
internal/auth/         keychain + encrypted-file token store
internal/movidesk/     SDK: HTTP client, rate limiter, retry, OData builder
internal/output/       JSON / table / CSV formatters
internal/version/      build metadata (set via -ldflags)
```

## Contributing

Issues and PRs welcome. Please run `make lint test` before submitting.

## License

MIT — see [LICENSE](./LICENSE).
