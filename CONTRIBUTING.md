# Contributing

Thanks for your interest! This project is an open-source CLI for the
Movidesk public REST API.

## Development setup

```bash
git clone https://github.com/jonasandre/movidesk-cli
cd movidesk-cli
make build
./bin/movidesk-cli --help
```

Requirements:

- Go 1.23+
- (Optional) `golangci-lint` for `make lint`
- (Optional) `goreleaser` for `make release-snapshot`

## Workflow

1. Open an issue describing the change before writing significant code.
2. Branch from `main`. Use a descriptive name (`feat/tickets-list`, `fix/retry-headers`).
3. Keep changes focused. One logical change per PR.
4. Add or update tests. We aim for >70% coverage in `internal/movidesk`.
5. Run `make lint test` locally; CI will block on either.
6. Use [Conventional Commits](https://www.conventionalcommits.org/) for messages.

## Code style

- `gofmt -s` is enforced. `make fmt` if needed.
- Errors flow up; never swallow without context.
- New Movidesk endpoints live under `internal/movidesk/<resource>/`. Each gets
  a CLI subcommand under `internal/cli/<resource>.go` and a default-columns
  hint for the table formatter.
- Never log a full URL containing a token. Use the `redact()` helper.

### PATCH semantics on `/tickets`

Movidesk's `PATCH /tickets` is partial for top-level scalar fields but
**replaces** array-valued fields like `customFieldValues`: any entry omitted
from the body is deleted server-side. **Never** issue a raw PATCH that touches
those arrays. Always go through the read-merge-patch helpers in
`internal/movidesk/tickets/customfields.go` (or analogous helpers for other
arrays). Tests in `customfields_test.go` cover the merge path; extend them
when adding new array-valued operations.

## Testing against a real Movidesk tenant

Integration tests are opt-in. Export `MOVIDESK_TOKEN` and
`MOVIDESK_BASE_URL` (sandbox), then run:

```bash
go test -tags=integration ./...
```

Do not commit recorded fixtures that contain real tickets, customer names, or
tokens.

## Releasing (maintainers)

1. Tag a new version: `git tag v0.X.Y && git push --tags`.
2. The `release` workflow runs `goreleaser release --clean` and produces
   archives, checksums, and an updated Homebrew tap formula.
3. Promote the GitHub release from draft to published once the release notes
   are reviewed.
