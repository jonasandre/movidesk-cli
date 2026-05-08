# Releasing movidesk-cli

This document is for maintainers cutting a release. Users do not need it.

## Pre-flight, once

1. **Create the Homebrew tap repo** at
   `https://github.com/jonasandre/homebrew-movidesk`. Empty `main` branch
   is fine — GoReleaser will push the formula into `Formula/` automatically.
2. **Create a GitHub Personal Access Token** scoped `repo` on the tap
   repository. Save it as the `HOMEBREW_TAP_TOKEN` secret on the
   `movidesk-cli` repo (Settings → Secrets and variables → Actions).
3. Verify that branch protection on `main` does not block the release
   workflow's `softprops/action-gh-release` step.

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
