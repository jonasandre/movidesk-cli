# movidesk-cli

A command-line interface for the [Movidesk](https://www.movidesk.com) public REST API.

- Multi-tenant: keep separate tokens for sandbox/staging/production, switch with one command.
- Tokens stored in the OS keychain (macOS Keychain, Windows Credential Manager, Linux libsecret/kwallet); encrypted file fallback for headless environments.
- OData passthrough on every list endpoint: `--filter`, `--select`, `--expand`, `--orderby`, `--top`, `--skip`.
- Output as JSON (default), human-readable table, or CSV.
- Honors Movidesk's 10 req/min rate limit with automatic backoff and `Retry-After` support.

> Status: **Phase 1.5 — Ticket schema + collections shipped.** Auth, HTTP client, full ticket schema typing (~80 fields), and the `tickets` family (`list`, `get`, `create`, `update`, `html`, `past list`, `merged list`, `attach`, `actions list/get/add/update/delete`, `clients list`, `relations`, `timeline`, `assets list`, `histories list`, `customfields show/set/clear/catalog`) are usable. Other resource families (persons, services, surveys, …) land in subsequent phases — see [the plan](./docs/plan.md).

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
