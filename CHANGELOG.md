# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] — 2026-05-08

### Security

- HTTP retry is now restricted to safe/idempotent methods (`GET`, `HEAD`,
  `OPTIONS`). Previously, transient failures on `POST`, `PATCH`, and
  `DELETE` could be retried automatically, risking duplicate writes
  (double-created tickets/persons, repeated state changes). Write
  operations no longer retry; use the explicit helpers or set
  `--no-retry` to disable retry entirely.
- Rate limiter no longer rolls back a slot on context cancellation. The
  previous code popped a stamp that had not actually been reserved (the
  reservation is taken before the wait), which could let bursts exceed
  the configured 10 req/min window after a cancelled request.

## [1.0.0] — 2026-05-08

First public release. Covers every API listed in the Movidesk integration menu.

### Added

#### Foundation
- Multi-tenant configuration at `~/.movidesk/config.yaml` (chmod 0600).
- Tokens stored in the OS keychain (macOS Keychain, Windows Credential
  Manager, Linux libsecret/kwallet) via `zalando/go-keyring`, with an
  AES-GCM/PBKDF2 encrypted-file fallback for headless environments.
- HTTP client with token query-param injection, sliding-window rate limiter
  (10 req/min per Movidesk's published cap), `Retry-After`-aware retries,
  and exponential backoff on 5xx.
- OData query builder for `$filter`/`$select`/`$expand`/`$orderby`/`$top`/`$skip`/`$count`.
- JSON / table / CSV output formatters with per-resource default columns
  and dot-path support for nested fields.
- Persistent flags: `--tenant`, `--output`, `--user`, `--no-color`,
  `--verbose`, `--no-retry`, `--compact`. Env overrides: `MOVIDESK_TENANT`,
  `MOVIDESK_TOKEN`, `MOVIDESK_USER`, `MOVIDESK_HOME`, `MOVIDESK_PASSPHRASE`.
- CI workflow: vet + test + lint + build on every PR.
- GoReleaser config building darwin/linux/windows × amd64/arm64 with
  Homebrew tap publishing on tag.

#### Auth
- `auth login --tenant` with hidden token prompt and live API validation.
- `auth list` (token never displayed), `auth switch`, `auth status`,
  `auth logout [--all]`, `auth token` (explicit print for piping).
- `auth set-user` plus default-user prompt on `auth login` for
  per-tenant `createdBy` injection.

#### Tickets
- `tickets list/get/create/update/html/past list/merged list/attach`
  with auto-pagination (`--all` + `--max`) and JSON-template create
  (`--file`, `--from-template`, `--from-template-file`, `--set k=v`).
- Full schema typing (~80 fields) with `Extra json.RawMessage` for
  forward compatibility on every nested type.
- Collection subcommands: `actions list/get/add/update/delete`,
  `clients list`, `relations`, `timeline`, `assets list`,
  `histories list`.
- Custom field subsystem: `customfields show/set/clear` with
  read-merge-patch semantics, plus a per-tenant catalog at
  `~/.movidesk/<tenant>/customfields.yaml` (`catalog list/add/remove`).
- Default `createdBy` injection on `create` and `actions add` when the
  tenant has a default user, with `--user <id>` override.

#### Persons
- `persons list/get/create/update/delete` with delete confirmation.
- `persons customfields show/set/clear` reusing the tickets catalog.

#### Services
- `services list/get/create/update/delete` with delete confirmation.

#### Activities
- `activities list/get/create/update/delete/add-teams`, with cursor
  pagination (`limit`/`startingAfter`/`name`) — Movidesk's API for
  activities is not OData.

#### Contracts (`/timeAgreement`)
- `contracts list/get/create/update/delete`.
- `contracts consumption list` (read-only `/timeAgreementConsumption`).

#### Surveys
- `surveys questions list/get` (read-only).
- `surveys responses list` with cursor pagination.

#### Knowledge base
- `kb articles get <id>` (single-article read; no public list endpoint).

#### Telephony
- `telephony queue --event <name>` for `/asterisk_{received,transfered,completed,lost,canceled}Call`.
- `telephony nonqueue --event <name>` for the no-queue variants.
- `telephony made-call-link` for `/setMadeCallLink`.

#### Custom field option pool
- `customfields options add/rename/remove` wrapping
  `/ticketCustomFieldValue/{Insert,Update,Delete}Values` so
  list-type field dropdowns can be managed tenant-wide.

#### Escape hatch
- `query <path>` for any GET/DELETE OData call against arbitrary
  endpoints, with optional `--all` auto-pagination.

### Notes

- PATCH on `/tickets` replaces array-valued fields like `customFieldValues`;
  the typed write helpers always go through read-merge-patch so callers
  describe only the change they want.
- The CLI is round-trip-safe via `--output json`: the underlying handler
  re-emits the response untouched, so even fields the SDK doesn't yet
  type are preserved end-to-end.

[Unreleased]: https://github.com/jonasandre/movidesk-cli/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/jonasandre/movidesk-cli/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/jonasandre/movidesk-cli/releases/tag/v1.0.0
