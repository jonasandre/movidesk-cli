# movidesk-cli

A command-line interface for the [Movidesk](https://www.movidesk.com) public REST API.

- Multi-tenant: keep separate tokens for sandbox/staging/production, switch with one command.
- Tokens stored in the OS keychain (macOS Keychain, Windows Credential Manager, Linux libsecret/kwallet); encrypted file fallback for headless environments.
- OData passthrough on every list endpoint: `--filter`, `--select`, `--expand`, `--orderby`, `--top`, `--skip`.
- Output as JSON (default), human-readable table, or CSV.
- Honors Movidesk's 10 req/min rate limit with automatic backoff and `Retry-After` support.

> Status: **Phase 0 — Foundation.** Auth and the core HTTP client are working. Resource commands (tickets, persons, services, …) ship in subsequent phases — see [the plan](./docs/plan.md) for the roadmap.

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
