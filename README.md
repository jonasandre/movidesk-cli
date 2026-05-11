# movidesk-cli

A command-line interface for the [Movidesk](https://www.movidesk.com)
public REST API.

- Multi-tenant: keep separate tokens for sandbox/staging/production and
  switch with one command.
- Tokens stored in the OS keychain (macOS Keychain, Windows Credential
  Manager, Linux libsecret/kwallet); encrypted file fallback for headless
  environments.
- OData passthrough on every list endpoint: `--filter`, `--select`,
  `--expand`, `--orderby`, `--top`, `--skip`, plus auto-pagination via
  `--all`.
- Output as JSON (default), human-readable table, or CSV.
- Honors Movidesk's 10 req/min rate limit with automatic backoff and
  `Retry-After` support.
- Per-tenant default user that the CLI auto-injects as `createdBy` on the
  writes that need attribution.

## Install

```bash
# Homebrew (macOS / Linux) — distributed as a Cask
brew install --cask jonasandre/movidesk/movidesk-cli

# Pre-built binary — pick your archive at:
#   https://github.com/jonasandre/movidesk-cli/releases

# From source
go install github.com/jonasandre/movidesk-cli/cmd/movidesk-cli@latest
```

> **macOS first-run note.** The released binary is not yet
> Apple-notarized, so Gatekeeper quarantines it on first launch and the
> process is killed silently (exit `137`). Strip the quarantine flag
> once after install:
>
> ```bash
> xattr -dr com.apple.quarantine "$(brew --caskroom)/movidesk-cli"
> # or, without Homebrew:
> xattr -dr com.apple.quarantine /path/to/movidesk-cli
> ```
>
> Re-run after each upgrade until proper signing/notarization lands
> (tracked in [`docs/RELEASING.md`](./docs/RELEASING.md)).

Enable shell tab-completion (Cobra ships bash, zsh, fish, PowerShell):

```bash
movidesk-cli completion zsh  > "${fpath[1]}/_movidesk-cli"
movidesk-cli completion bash > /usr/local/etc/bash_completion.d/movidesk-cli
movidesk-cli completion fish > ~/.config/fish/completions/movidesk-cli.fish
```

The Homebrew Cask registers `bash`, `zsh`, and `fish` completions
automatically; the snippets above are only needed for installs outside
Homebrew.

## Quickstart

```bash
# 1. Add a tenant. The token is read from a hidden prompt and validated
#    against GET /persons?$top=1 before being stored. Optionally configure
#    a default user (Cod. Ref.) for createdBy attribution.
movidesk-cli auth login --tenant prod --label "Acme Production" --make-default

# 2. List configured tenants (token never displayed).
movidesk-cli auth list

# 3. Verify the connection.
movidesk-cli auth status

# 4. Use it.
movidesk-cli tickets list --top 5 --output table
```

Full reference for every command lives under [`docs/cli/`](./docs/cli/).
A few common entry points:

| Family | Highlights |
|---|---|
| `auth` | `login`, `list`, `switch`, `status`, `set-user`, `logout`, `token` |
| `tickets` | `list`, `get`, `create`, `update`, `attach`, `actions`, `clients`, `relations`, `timeline`, `customfields`, `past`, `merged`, `html` |
| `persons` | `list`, `get`, `create`, `update`, `delete`, `customfields` |
| `services` | `list`, `get`, `create`, `update`, `delete` |
| `activities` | `list`, `get`, `create`, `update`, `delete`, `add-teams` |
| `contracts` | `list`, `get`, `create`, `update`, `delete`, `consumption list` |
| `surveys` | `questions list/get`, `responses list` |
| `kb` | `articles get` |
| `telephony` | `queue`, `nonqueue`, `made-call-link` |
| `customfields` | `options add/rename/remove` |
| `query` | raw OData escape hatch |

`movidesk-cli <command> --help` always shows the local flags. The reference
pages under `docs/cli/` are auto-generated and cover every subcommand.

## Conventions

**Token & tenant overrides** — `--tenant <name>` or `MOVIDESK_TENANT` to
target a specific tenant; `MOVIDESK_TOKEN` short-circuits the keychain
(useful in CI).

**Default user** — `--user <id>` or `MOVIDESK_USER` overrides the per-tenant
configured user. Auto-injected as `createdBy` on `tickets create` and
`tickets actions add` only — never on updates, never when the body already
provides one.

**Body input** — write commands (`create`, `update`, ...) accept three
input modes, in order of precedence:

1. `--file <path>` — full JSON body.
2. `--from-template <name>` / `--from-template-file <path>` — reusable
   templates under `~/.movidesk/templates/<name>.json`.
3. `--set key=value` (repeatable) — inline overrides; values are JSON-parsed
   first, falling back to strings (`--set type=2` → number,
   `--set tags='["a","b"]'` → array).

**Output formats** — `--output json|table|csv`, plus `--compact` for
single-line JSON, `--columns id,subject,owner.businessName` to override
table/CSV columns (dot-paths supported for nested fields).

**Rate limiting** — every request goes through an in-process 10 req/min
limiter and retries 429/5xx with `Retry-After` honored. `--no-retry`
disables retries when debugging.

## Configuration files

- `~/.movidesk/config.yaml` — tenant list, current tenant, defaults.
  Mode `0600`.
- `~/.movidesk/credentials.enc` — encrypted token fallback, only used when
  the OS keychain is unavailable. AES-GCM, mode `0600`.
- `~/.movidesk/templates/<name>.json` — reusable bodies for write commands.
- `~/.movidesk/<tenant>/customfields.yaml` — per-tenant catalog mapping
  human labels to numeric custom-field IDs (Movidesk has no public API to
  discover them).

Set `MOVIDESK_HOME` to relocate the directory — handy for tests and
ephemeral CI environments.

> **Never commit these files.** They grant the same access as the user
> that owns them.

## Caveats worth knowing up front

- **PATCH on `/tickets` replaces array-valued fields** like
  `customFieldValues`. The CLI's typed write helpers go through
  read-merge-patch so you only describe the change. If you build your own
  body, send the full array.
- **Movidesk does not expose an API to list custom-field definitions.**
  Populate the local catalog from the IDs visible in the Movidesk web UI.
- **`tickets list` only returns tickets updated within the last 90 days.**
  Older ones live on `/tickets/past` (`tickets past list`).
- **Activities, surveys/responses use cursor pagination**, not OData
  (`--limit` / `--starting-after`); use `--all` to walk every page.

## Development

```bash
make build              # build binary into ./bin
make test               # race-enabled, no cache
make lint               # golangci-lint
make cover              # coverage summary
make docs               # regenerate docs/cli/*.md from Cobra
make run ARGS="auth list"
make release-check      # validate .goreleaser.yaml
make release-snapshot   # local goreleaser dry-run
```

For maintainers cutting releases, see [`docs/RELEASING.md`](./docs/RELEASING.md).

Module layout:

```
cmd/movidesk-cli/      entry point
cmd/gen-docs/          docs/cli/ generator
internal/cli/          Cobra commands
internal/config/       multi-tenant YAML config
internal/auth/         keychain + encrypted-file token store
internal/movidesk/     SDK: HTTP client, rate limiter, retry, OData builder
internal/output/       JSON / table / CSV formatters
internal/version/      build metadata (set via -ldflags)
```

## Contributing

Issues and PRs welcome. Please run `make lint test` before submitting.
See [CONTRIBUTING.md](./CONTRIBUTING.md) for development conventions.

## License

MIT — see [LICENSE](./LICENSE).
