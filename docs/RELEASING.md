# Releasing movidesk-cli

This document is for maintainers cutting a release. Users do not need it.

## Pre-flight, once

### 1. Create the Homebrew tap repo

`https://github.com/jonasandre/homebrew-movidesk`. Empty `main` branch is
fine — GoReleaser pushes the formula into `Formula/` on every release.

### 2. Mint the Homebrew tap token

Use a **fine-grained Personal Access Token** (preferred over classic):
smaller blast radius, mandatory expiration, granular permissions.

Path: GitHub → **Settings** → **Developer settings** →
**Personal access tokens** → **Fine-grained tokens** → **Generate new token**.

| Field | Value |
|---|---|
| Token name | `goreleaser-homebrew-movidesk` |
| Resource owner | `jonasandre` |
| Repository access | **Only select repositories** → `homebrew-movidesk` |
| Expiration | 90 days (set a calendar reminder to rotate) |

Repository permissions — leave everything **No access** except:

| Permission | Access | Why |
|---|---|---|
| **Contents** | Read and write | GoReleaser commits the formula file |
| **Metadata** | Read-only | Granted automatically when any other permission is selected |

Account permissions: none.

Save the token as a repository secret on `movidesk-cli`:

GitHub → `movidesk-cli` repo → **Settings** → **Secrets and variables** →
**Actions** → **New repository secret**. Name it `HOMEBREW_TAP_TOKEN`,
paste the `github_pat_…` value.

### 3. Confirm release workflow access

Verify branch protection on `main` does not block the release workflow's
GoReleaser step. The workflow needs `contents: write` on the
`movidesk-cli` repo (already granted in `.github/workflows/release.yml`)
and the fine-grained token above for the tap repo.

### 4. Calendar a rotation reminder

The fine-grained token expires; set a reminder ~1 week before the
expiration date to mint a fresh one and update `HOMEBREW_TAP_TOKEN`. If
the token expires mid-release, GitHub Releases still publish — only the
brew tap update fails. Re-run the release workflow once the token is
refreshed to retry the tap push.

## Cutting a release

```bash
# 1. Make sure main is green and your tree is clean.
git switch main
git pull --ff-only
git status   # must be clean

# 2. Run the local checks before tagging.
make lint test
make docs        # regenerates docs/cli/*.md (commit if anything changed)

# 3. Update CHANGELOG.md: move the [Unreleased] section under the new
#    version heading and add a fresh empty [Unreleased] block at the top.
${EDITOR:-vi} CHANGELOG.md

# 4. Commit + tag (signed if you sign).
git commit -am "release: vX.Y.Z"
git tag -s vX.Y.Z -m "vX.Y.Z"
git push --follow-tags
```

The `release` workflow at `.github/workflows/release.yml` triggers on the
tag, runs `goreleaser release --clean`, and:

- Builds darwin/linux/windows × amd64/arm64 from `cmd/movidesk-cli`.
- Bundles `LICENSE`, `README.md`, `CHANGELOG.md` in each archive.
- Publishes a **draft** GitHub release with auto-generated notes.
- Pushes a Homebrew formula to `jonasandre/homebrew-movidesk/Formula/`.

## After the workflow finishes

1. Review the draft release on the GitHub Releases page. The body comes
   from your conventional-commit history grouped by `feat:` / `fix:` /
   `chore:` etc. (see `.goreleaser.yaml` `changelog.groups`).
2. Smoke-test on at least one platform:
   ```bash
   brew uninstall movidesk-cli || true
   brew untap jonasandre/movidesk || true
   brew tap jonasandre/movidesk
   brew install movidesk-cli
   movidesk-cli --version
   movidesk-cli --help
   ```
3. **Publish** the draft release once you're satisfied.

## Local snapshots

To validate the GoReleaser config without tagging:

```bash
make release-check       # equivalent to `goreleaser check`
make release-snapshot    # builds artifacts under ./dist (no upload)
ls dist/
```

`release-snapshot` requires GoReleaser locally:
`go install github.com/goreleaser/goreleaser/v2@latest`.

## Rolling back a release

If something is broken after the GitHub Release is published:

```bash
gh release delete vX.Y.Z --cleanup-tag
git push --delete origin vX.Y.Z
# Optionally: revert the offending commit on main and cut vX.Y.Z+1 fresh.
```

The Homebrew formula will still point at the deleted release until the
next release overwrites it. Push a manual revert PR to the tap repo if
the broken version is critical.

## Versioning policy

- `v0.x.y` — pre-1.0, breaking changes can ship in any minor.
- `v1.x.y` — Semantic Versioning, breaking changes only in major bumps.
- The CLI's flags and config schema are part of the public surface;
  removing or renaming a flag is a breaking change.
- The Go SDK under `internal/movidesk/...` is **not** a public API;
  importing it from outside this module is unsupported.
