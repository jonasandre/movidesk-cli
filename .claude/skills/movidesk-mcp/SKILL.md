---
name: movidesk-mcp
description: How to query the Movidesk help-desk API through the embedded Movidesk MCP server — picking the right tool, building OData `$filter` expressions that the API actually accepts, paginating safely under the 10 req/min rate limit, and trimming responses with `$select`/`$expand`. Use this skill whenever the user mentions Movidesk, tickets, contracts, time-agreements, support agents, surveys, or invokes any tool whose name starts with `tickets_`, `persons_`, `services_`, `contracts_`, `kb_`, `activities_`, `surveys_`, or the generic `query` tool — even if the user does not explicitly say "MCP". Also use it when a previous Movidesk MCP call returned an HTTP 400 / 429 / truncation message, so the next call is built correctly the first time.
---

# Movidesk MCP guide

The Movidesk MCP server is read-only and wraps the Movidesk REST/OData API. All list-style tools accept the same shared OData arguments (`filter`, `select`, `expand`, `orderby`, `top`, `skip`, `all`, `max`). The shape that the LLM sends is JSON, but the values follow OData v3 conventions, **not** standard SQL or JS — most failures come from forgetting that.

The hard constraints to keep in mind:

- **Rate limit: 10 requests/minute per tenant token**, shared with anything else hitting Movidesk. Burning the budget mid-task strands the user. Prefer one well-shaped page over many speculative calls.
- **Response cap: 256 KB per tool call.** Beyond that the result is truncated with `[truncado: resposta excedeu 262144 bytes; refine $select ou $top]`. When you see that marker, narrow `select` or shrink `top` — do not retry with the same shape.
- **Default `max` is 500 rows** when `all: true` and `max` is unset. Set `max` explicitly when the user clearly needs more, and keep it bounded — `all: true` with no plan is how rate-limit blowouts happen.
- **Tickets list only covers the last 90 days.** Older tickets live in `tickets_past_list`. If `lastUpdate` is older than ~90 days and `tickets_list` returns nothing, switch tools instead of fiddling with the filter.

## Tools at a glance

Pick the most specific tool — `query` is the escape hatch, not the default.

| Domain | Tool | When to use |
|---|---|---|
| Tickets | `tickets_list` | Recent tickets (≤90d). Supports `$filter`/`$select`/`$expand`/`$orderby`. |
| Tickets | `tickets_past_list` | Archived tickets (`lastUpdate` >90d). Same OData shape, no `include_deleted`. |
| Tickets | `tickets_get` | Single ticket by `id` **xor** `protocol`. Never pass both. |
| Tickets | `tickets_html_description` | Raw HTML body of ticket (`action_id=0`) or a specific action. |
| Tickets | `tickets_actions_list` | All actions (messages, internal notes) of one ticket. |
| Tickets | `tickets_timeline` | Chronological merge of actions + status history + owner history. |
| Tickets | `tickets_customfields_show` | Current customFieldValues of a single ticket. |
| Persons | `persons_list` | People/companies/agents/departments. Filter on `personType`/`profileType`/`isActive`. |
| Persons | `persons_get` | Single person by `id` (Cod. Ref.). |
| Persons | `persons_customfields_show` | Custom fields of a single person. |
| Catalog | `services_list` / `services_get` | Service catalog (`id`, `name`, `parentServiceId`, …). |
| Contracts | `contracts_list` / `contracts_get` | Time-agreement contracts (`isActive`, `beginDate`, `endDate`). |
| Contracts | `contracts_consumption_list` | Consumption entries from `/timeAgreementConsumption`. |
| KB | `kb_article_get` | Knowledge-base article by `id`. |
| Activities | `activities_list` / `activities_get` | Configured activity types. `activities_list` takes `name_filter` (substring), not OData. |
| Surveys | `surveys_questions_list` / `_get` | Satisfaction-survey questions. `type` filter is an int, **not** an OData expression. |
| Surveys | `surveys_responses_list` | Responses, cursor-paginated server-side; capped by `max`. |
| Escape hatch | `query` | Any read-only OData endpoint not covered above. Path must start with `/`. |

## Resources to read on demand

Three MCP resources sit alongside the tools. Read them when you need ground truth instead of guessing:

- `movidesk://odata-filter-syntax` — the canonical filter reference (operators, literals, common fields, traps). Read it before composing any non-trivial `$filter`.
- `movidesk://customfields-catalog` — per-tenant map of human label → `{id, rule_id, type, options}`. Available only when the tenant has a local `customfields.yaml`. Read it before filtering or expanding on `customFieldValues` so you reference the right numeric id.
- `movidesk://server-info` — tenant name, base URL, transport, rate limit, scope. Read it once at the start of a session if you are unsure which tenant you are talking to.

## Building `$filter` expressions

The string in `filter` is passed verbatim to the Movidesk API. Nothing is parsed client-side, so a typo becomes an HTTP 400 with the raw upstream message.

### Operators

```
eq  ne  lt  le  gt  ge          comparison
and or not (…)                  logical, parentheses control precedence
```

### String functions

```
startswith(field, 'text')
endswith(field, 'text')
contains(field, 'text')
substringof('text', field)      OData v3 form, still accepted
tolower(field)  toupper(field)
```

### Literals

| Type | Form |
|---|---|
| String | Single quotes: `'Qlik'`. Embedded apostrophe doubles: `'O''Brien'`. |
| Number | Bare: `42`, `3.14`. |
| Date | ISO 8601 **UTC with `Z`**: `2026-04-01T00:00:00.000Z`. `+00:00` and naive dates are rejected. |
| Boolean | `true` / `false`. |
| Null | `null`. |

### Common fields per resource

These are the fields that almost always work — not exhaustive, but enough for most filters. For everything else, expand the resource via `$expand` or fall back to `query` and inspect the raw response.

- **tickets**: `id`, `protocol`, `subject`, `type` (1=internal, 2=external), `status`, `baseStatus` (`'New'`, `'InAttendance'`, `'Stopped'`, `'Resolved'`, `'Closed'`, `'Canceled'`), `justification`, `ownerTeam` *(string, not a navigation)*, `category`, `urgency`, `urgencyId`, `createdDate`, `lastUpdate`, `lastActionDate`, `slaSolutionDate`, `createdBy/id`, `owner/id`, `owner/businessName`. Nested collections: `actions[]` (messages + notes), `actions[].timeAppointments[]` (logged hours, see below), `actions[].expenses[]`, `parentTickets[]` / `childrenTickets[]`, `customFieldValues[]`.
- **persons**: `id`, `businessName`, `corporateName`, `codeReferenceAdditional`, `personType` (1=Person, 2=Company, 4=Department), `profileType` (1=Agent, 2=Customer, 3=Both), `isActive`, `userName`.
- **services**: `id`, `name`, `parentServiceId`, `isVisible`, `allowFinalUser`.
- **contracts**: `id`, `name`, `isActive`, `beginDate`, `endDate`.

### Common traps

These come up over and over — internalise them:

1. **`ownerTeam` is a string, not a navigation property.** ✅ `ownerTeam eq 'Qlik'`  ❌ `ownerTeam/name eq 'Qlik'` → HTTP 400 *"segment ... isn't Navigation/Structural/Complex/Collections"*.
2. **Dates must be UTC with `Z`.** Movidesk frequently rejects `+00:00` or fuso-less timestamps. Always normalise to `…Z`.
3. **`$select` and `$expand` shape only the response — never the predicate.** To filter on a nested field, reference it directly in `$filter` (e.g. `createdBy/id eq 'abc'`); you don't need to expand it.
4. **Numeric enums vs string enums.** `baseStatus` is a string (`'InAttendance'`). `personType` / `profileType` / `type` are integers. Mixing them yields 400.
5. **Custom fields go through `customFieldValues`.** Resolve the human label to its `id` via `movidesk://customfields-catalog`, then `$filter` looks like `customFieldValues/any(c: c/customFieldId eq 12345 and c/value eq 'X')`.
6. **`include_deleted` is a separate boolean argument**, not an OData clause. Don't try to express it in `filter`.

### Example filters

```jsonc
// Tickets opened in April 2026 owned by team "Qlik"
{
  "filter": "createdDate ge 2026-04-01T00:00:00.000Z and createdDate lt 2026-05-01T00:00:00.000Z and ownerTeam eq 'Qlik'",
  "select": ["id", "protocol", "subject", "createdDate"]
}

// Active companies
{
  "filter": "personType eq 2 and isActive eq true",
  "select": ["id", "businessName"]
}

// Open tickets, most recent first
{
  "filter": "baseStatus eq 'InAttendance'",
  "orderby": ["createdDate desc"],
  "top": 20
}

// SLA breached, last 7 days
{
  "filter": "slaSolutionDate lt 2026-05-04T00:00:00.000Z and baseStatus ne 'Closed' and createdDate ge 2026-05-04T00:00:00.000Z",
  "select": ["id", "protocol", "subject", "slaSolutionDate", "baseStatus"]
}
```

## Paginating safely

The MCP server exposes pagination through four args:

- `top` — page size sent as `$top`. `0` lets the server pick (typically 50). Push it up to ~100 when the rows are slim; keep it ≤25 when expanding heavy relations.
- `skip` — offset (`$skip`). Use to fetch a specific window without auto-pagination.
- `all: true` — auto-paginate until exhaustion or `max`. `top` becomes the page size. This is **one call from the user's perspective but many from the rate-limiter's**.
- `max` — row cap when `all: true`. Defaults to 500. Always set it explicitly when you suspect the dataset is large.

### Decision rules

- **Counting / sampling**: do **not** use `all: true`. Send a single page with `top: 1` (or the count you need) and read the length.
- **Single page, known offset**: set `top` + `skip` directly. Useful when the user asks for "page 3 of 20".
- **Full sweep, bounded**: `all: true` + an explicit `max`. Pick `max` based on the user's stated need — never default to "as many as possible".
- **Heavy expand**: shrink `top` to 20–50. Each row is bigger, so you hit the 256 KB cap sooner and burn the request budget faster.

### Surviving the 256 KB cap

When the response is truncated, the tool result ends with `[truncado: resposta excedeu 262144 bytes; refine $select ou $top]`. Do not retry the same shape — Movidesk has already charged you the request. Instead:

1. Add `$select` listing only the fields actually needed.
2. Halve `top`.
3. If still too big, drop `$expand` and fetch the relation in a follow-up call for the rows that matter.

## `$select` and `$expand`

Projection cuts both transfer time and the chance of hitting the 256 KB cap. Default to a tight `select` for any list of more than ~10 fields.

```jsonc
{
  "filter": "baseStatus eq 'InAttendance'",
  "select": ["id", "protocol", "subject", "owner/businessName"],
  "expand": ["owner"],
  "top": 50
}
```

Notes:

- `expand` is a JSON array of strings (e.g. `["actions", "customFieldValues"]`). Nested expand uses slashes inside one entry: `"actions/createdBy"`.
- `select` paths can reach into expanded relations: `"owner/businessName"` works as long as `owner` is expanded **or** the API exposes it as a navigation.
- For `customFieldValues`, prefer the dedicated `tickets_customfields_show` / `persons_customfields_show` when you only need that subset — they bypass `$expand` and return a clean array.

## Time appointments (apontamentos de tempo)

Time appointments are the hours an agent logs against a ticket. **They are not a top-level resource in the MCP tool surface** — they live nested inside `ticket.actions[n].timeAppointments[n]`. The Movidesk OData backend does expose `/timeAppointments` directly, so when you need a cross-ticket query (e.g. "all hours logged by team X this week") fall back to the `query` escape hatch.

### Shape of a `TimeAppointment`

| Field | Notes |
|---|---|
| `id` | Numeric, unique per appointment. |
| `activity` | Free-form label of what the agent was doing (e.g. `"Triage"`). |
| `date` | The work date (UTC `Z`). |
| `periodStart` / `periodEnd` | Optional time window inside `date`. |
| `workTime` | Duration string, typically `"HH:MM"` (e.g. `"01:30"`). |
| `workTypeName` / `workTypeId` | Work-type taxonomy (billable / non-billable / training / etc — depends on tenant). |
| `accountable` | Person who performed the work (the agent the hours belong to for reporting). |
| `createdBy` | Person who logged the entry. Usually equal to `accountable`, but not always — a supervisor may log on behalf of an agent. |
| `createdDate` | When the entry was created. |
| `isDeleted` | Soft-delete flag. |

For "who actually did the work" use `accountable`. For "who keyed the entry" use `createdBy`. Don't conflate them.

### Retrieving them

**Per ticket (most common):** `tickets_get` already returns `actions[].timeAppointments[]` when you ask for it. The cheapest call is:

```jsonc
{
  "id": 12345,
  // tickets_get returns the full payload — actions and their timeAppointments come embedded by default
}
```

If you only want the appointment summary and not the rest of the action body, route through `tickets_list` with a tight `$select` and `$expand`:

```jsonc
{
  "filter": "id eq 12345",
  "select": ["id", "actions/id", "actions/createdDate", "actions/timeAppointments"],
  "expand": ["actions/timeAppointments", "actions/timeAppointments/accountable"]
}
```

Nested expand uses the slash form inside one `expand` entry (`"actions/timeAppointments"`). Expanding `accountable` brings the agent's `businessName` so you don't need a follow-up `persons_get`.

**Cross-ticket (escape hatch):** use `query` against `/timeAppointments` directly. Useful for "all hours logged by an agent in May", regardless of which ticket they're on:

```jsonc
{
  "path": "/timeAppointments",
  "filter": "accountable/id eq 'agent-id' and date ge 2026-05-01T00:00:00.000Z and date lt 2026-06-01T00:00:00.000Z",
  "select": ["id", "ticketId", "activity", "date", "workTime", "workTypeName"],
  "expand": ["accountable"],
  "top": 100,
  "all": true,
  "max": 1000
}
```

### Common pitfalls

- **`workTime` is a string `"HH:MM"`**, not a decimal. Parse it client-side: `(h * 60 + m)` for sortable minutes; reformat back to `HH:MM` for display.
- **Date vs createdDate**: filter on `date` for "when the work happened", on `createdDate` for "when it was logged". Late retroactive entries skew the latter.
- **Soft-deleted appointments**: `isDeleted: true` rows still come back. Drop them client-side unless the user explicitly asked for the deleted ones.
- **TimeAppointment vs timeAgreementConsumption**: don't confuse them. `timeAppointments` are *the agent's logbook entries* on a ticket action. `timeAgreementConsumption` (via `contracts_consumption_list`) is the **contract's ledger** — it draws from the agent's logged hours but is its own resource, often with different field names (`workTime` on appointment vs `hours` on consumption, depending on tenant). If the user asks about a *contract*, use consumption; if they ask about *agent productivity* or *ticket effort*, use appointments.

### Recipes

**"How many hours did agent X log on ticket Y?"** → `tickets_get` by id, walk `actions[].timeAppointments[]`, filter where `accountable.id == X && !isDeleted`, sum parsed `workTime`.

**"All hours agent X logged last week, grouped by ticket"** → `query` `/timeAppointments` with `accountable/id eq 'X' and date ge <monday-Z> and date lt <next-monday-Z>`, tight `$select` including `ticketId`. Aggregate client-side by `ticketId`.

**"Billable vs non-billable hours for team T in May"** → `query` `/timeAppointments` with date window + `$expand` on `accountable`, then filter client-side by `accountable.team` (workTypeName is the billable signal; the taxonomy is tenant-specific so confirm with the tenant first).

## Common workflows

### "Find tickets opened today for company X"

1. Resolve company id with `persons_list` filtered by `corporateName` or `businessName`.
2. `tickets_list` with `filter: "clients/any(c: c/id eq '<id>') and createdDate ge <today-UTC-Z>"`, tight `select`.

### "How is contract <name> being consumed this month?"

1. `contracts_list` filtered by `name` to get the contract id.
2. `contracts_consumption_list` filtered on `timeAgreementId eq <id>` and `workTime ge <month-start-Z>`. Pull a tight `select` (`id`, `workTime`, `ticketId`, `quantity`).

### "Reconstruct the conversation on ticket <protocol>"

1. `tickets_get` with `protocol` → grab the `id`.
2. `tickets_timeline` for the chronological view, **or** `tickets_actions_list` for raw actions.
3. If a specific message body matters, `tickets_html_description` with the matching `action_id`.

### "How many tickets did agent <name> close last week?"

1. `persons_list` with `profileType eq 1 and contains(businessName, '<name>')` to get the agent id.
2. `tickets_list` with `filter: "owner/id eq '<id>' and baseStatus eq 'Closed' and lastUpdate ge <week-start-Z> and lastUpdate lt <week-end-Z>"`, `select: ["id"]`, `top: 1` first to gauge volume, then re-run with `all: true` and a sane `max` if the user needs the full list.

### "Anything from this endpoint Movidesk added last week"

Use `query`. Always start with a tiny page (`top: 5`) so the structure is visible, then iterate.

## Error handling cheatsheet

| Symptom | Likely cause | Fix |
|---|---|---|
| HTTP 400 `"segment '…' isn't Navigation/Structural/…"` | Treated a string field as nav (e.g. `ownerTeam/name`). | Use the bare field. |
| HTTP 400 on date filter | `+00:00` or naive timestamp. | Switch to `…Z`. |
| HTTP 400 `"Could not find a property named '…'"` | Field name wrong or pluralisation off. | Cross-check against the resource table above or `movidesk://odata-filter-syntax`. |
| HTTP 429 | Burned 10 req/min budget. | Stop, wait, plan a tighter strategy. Mention to the user. |
| `[truncado: …]` marker | 256 KB cap hit. | Narrow `select`, shrink `top`, drop heavy `expand`. |
| Empty list for old ticket | `tickets_list` only covers ~90 days. | Switch to `tickets_past_list`. |
| `informe id OU protocol (exatamente um)` | Sent both (or neither) to `tickets_get`. | Pick one. |
