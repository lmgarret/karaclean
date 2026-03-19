# Phase 10: CI: Run Tests, Lint, and Build Docker Image - Research

**Researched:** 2026-03-19
**Domain:** GitHub Actions CI for Go projects (testing, linting, Docker build/push to GHCR)
**Confidence:** HIGH

## Summary

This phase is greenfield CI -- no `.github/` directory exists. The project has 9 test files across 5 packages, a working multi-stage Dockerfile (alpine builder to scratch), and uses Go 1.26.1. All decisions are locked in CONTEXT.md: GitHub Actions with 3 separate jobs (test, lint, docker), golangci-lint v2 with a pinned linter set, and Docker push to ghcr.io on main only.

The main technical nuance is golangci-lint v2's new configuration format. The v2 line (current: v2.11.3) uses `version: "2"` in `.golangci.yml`, merged `gosimple` into `staticcheck`, and moved formatters to a separate section. The `default: standard` preset covers errcheck, govet, ineffassign, staticcheck, and unused -- matching the user's "standard" set (minus `gosimple` which is now part of `staticcheck`). The extras (`gocyclo`, `godot`, `misspell`, `noctx`) go in the `enable` list.

**Primary recommendation:** Create two files: `.github/workflows/ci.yml` (3-job workflow) and `.golangci.yml` (v2 format config). Use pinned major versions for all actions.

<user_constraints>

## User Constraints (from CONTEXT.md)

### Locked Decisions
- GitHub Actions (`.github/workflows/`) on `ubuntu-latest` with `actions/setup-go`
- Cache go modules (`~/go/pkg/mod`) and build cache (`~/.cache/go-build`) between runs
- Use `golangci-lint` with a `.golangci.yml` config file
- Linter set: standard (errcheck, gosimple, govet, ineffassign, staticcheck, unused) + extras (gocyclo, godot, misspell, noctx)
- Build Docker image on all triggers; push to `ghcr.io` only on merge to `main`
- Tags: `ghcr.io/<owner>/karaclean:latest` + `ghcr.io/<owner>/karaclean:<git-sha>`
- Registry auth: `GITHUB_TOKEN` (auto-available)
- Triggers: `push` to `main` + `pull_request` targeting `main`
- Concurrency groups: cancel in-progress runs on same branch/PR
- Separate jobs: `test`, `lint`, `docker` -- docker depends on test+lint passing
- Tests: `go test ./... -race`

### Claude's Discretion
None -- all decisions locked.

### Deferred Ideas (OUT OF SCOPE)
None.

</user_constraints>

## Standard Stack

### Core Actions

| Action | Version | Purpose | Why Standard |
|--------|---------|---------|--------------|
| `actions/checkout` | `v6` | Clone repository | Official GitHub action |
| `actions/setup-go` | `v6` | Install Go toolchain + built-in module/build caching | Official GitHub action; built-in cache eliminates need for separate `actions/cache` |
| `golangci/golangci-lint-action` | `v9` | Run golangci-lint | Official action from golangci-lint authors; handles install + caching |
| `docker/login-action` | `v4` | Authenticate to ghcr.io | Official Docker action |
| `docker/metadata-action` | `v6` | Generate image tags (latest, sha) | Official Docker action; deterministic tag generation |
| `docker/build-push-action` | `v6` | Build and conditionally push image | Official Docker action; integrates with buildx |
| `docker/setup-buildx-action` | `v3` | Set up Docker Buildx builder | Required by build-push-action for advanced features |

### Linting Tool

| Tool | Version | Purpose |
|------|---------|---------|
| `golangci-lint` | `v2.11` | Meta-linter for Go (pin minor, not patch -- action handles binary download) |

### Key Version Notes

- **Go version:** `1.26.1` (from `go.mod`) -- use `go-version-file: go.mod` in `actions/setup-go` to auto-detect
- **golangci-lint v2 vs v1:** v2 has a **completely different config format**. Must use `version: "2"` in `.golangci.yml`. The `gosimple` linter is merged into `staticcheck` in v2.
- **actions/setup-go caching:** v5+ has built-in caching enabled by default (`cache: true`). It caches `~/go/pkg/mod` and `~/.cache/go-build` automatically. No separate `actions/cache` step needed.

## Architecture Patterns

### Project Structure (new files only)

```
.github/
  workflows/
    ci.yml          # Single workflow file with 3 jobs
.golangci.yml       # golangci-lint v2 config at repo root
```

### Pattern 1: Three-Job Workflow with Dependency

**What:** Separate `test`, `lint`, and `docker` jobs where `docker` depends on both others passing.
**When to use:** Always -- this is the locked decision.
**Example:**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

concurrency:
  group: ${{ github.workflow }}-${{ github.head_ref || github.ref }}
  cancel-in-progress: true

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
      - run: go test -race ./...

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
      - uses: golangci/golangci-lint-action@v9
        with:
          version: v2.11

  docker:
    needs: [test, lint]
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v6
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v4
        if: github.event_name == 'push' && github.ref == 'refs/heads/main'
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/metadata-action@v6
        id: meta
        with:
          images: ghcr.io/${{ github.repository }}
          tags: |
            type=raw,value=latest
            type=sha,prefix=
      - uses: docker/build-push-action@v6
        with:
          context: .
          push: ${{ github.event_name == 'push' && github.ref == 'refs/heads/main' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
```

### Pattern 2: golangci-lint v2 Config File

**What:** `.golangci.yml` with v2 format -- `version: "2"`, `default: standard`, extras in `enable`.
**Example:**

```yaml
version: "2"

linters:
  default: standard
  enable:
    - gocyclo
    - godot
    - misspell
    - noctx

  settings:
    gocyclo:
      min-complexity: 15
    godot:
      scope: toplevel
      capital: false
    errcheck:
      check-type-assertions: true
```

**Key points:**
- `default: standard` enables: errcheck, govet, ineffassign, staticcheck, unused
- `gosimple` from the user's list is now part of `staticcheck` in v2 -- no separate enable needed
- `gocyclo` min-complexity: 15 is a reasonable default (not too strict for a small project)
- `noctx` has no special settings needed

### Pattern 3: Conditional Push Logic

**What:** Build on every trigger but push only on main merge.
**How:** Use `push:` condition on login step and `push:` parameter on build-push-action.

```yaml
push: ${{ github.event_name == 'push' && github.ref == 'refs/heads/main' }}
```

This builds the image on PRs (verifying it compiles) without pushing to the registry.

### Pattern 4: Concurrency Groups

**What:** Cancel in-progress CI runs when a new push arrives on the same branch/PR.
**Example:**

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.head_ref || github.ref }}
  cancel-in-progress: true
```

`github.head_ref` is set for PRs (branch name), `github.ref` for pushes. This ensures each branch/PR gets its own concurrency group.

### Anti-Patterns to Avoid
- **Pinning actions to SHA:** Adds maintenance burden for no security benefit on official actions. Use major version tags (e.g., `@v6`).
- **Using `actions/cache` with `actions/setup-go`:** setup-go v5+ has built-in caching. Adding a separate cache step creates conflicts.
- **`enable-all` in golangci-lint:** Causes breakage on every minor update. Use `default: standard` + explicit extras.
- **Single monolithic job:** Wastes CI time when one step fails early. Separate jobs run in parallel and fail fast.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Docker tag generation | Shell script to compute tags | `docker/metadata-action` | Handles SHA, latest, semver, branch tags correctly across push/PR events |
| Go module caching | Manual `actions/cache` with key computation | `actions/setup-go` built-in cache | Auto-detects `go.sum` hash, handles restore keys |
| golangci-lint install | `go install` or `curl` in workflow | `golangci/golangci-lint-action` | Handles binary download, version resolution, result caching |
| Conditional Docker push | Custom if/else bash logic | `push:` parameter expression | Single expression handles push-vs-PR cleanly |

## Common Pitfalls

### Pitfall 1: golangci-lint v1 Config with v2 Binary
**What goes wrong:** v2 binary cannot parse v1 config format. Workflow fails with config parse error.
**Why it happens:** Many online examples still show v1 format (no `version:` field, `linters-settings:` instead of `linters.settings:`).
**How to avoid:** Always include `version: "2"` as the first line. Use `linters:` with `default:` and `settings:` (not `linters-settings:`).
**Warning signs:** Error message mentioning "configuration file version" or unexpected field names.

### Pitfall 2: Missing `permissions` for GHCR Push
**What goes wrong:** Docker push fails with 403 Forbidden.
**Why it happens:** GitHub Actions default token doesn't have `packages: write` unless explicitly granted.
**How to avoid:** Set `permissions: { contents: read, packages: write }` on the docker job.
**Warning signs:** "denied: permission_denied" in docker push output.

### Pitfall 3: `gosimple` Not Found in v2
**What goes wrong:** golangci-lint v2 errors when `gosimple` is listed as an enabled linter.
**Why it happens:** `gosimple` was merged into `staticcheck` in v2. Listing it separately causes "unknown linter" error.
**How to avoid:** Don't list `gosimple` -- it's included in `staticcheck` automatically.

### Pitfall 4: Docker Login on PR Builds
**What goes wrong:** PRs from forks fail because `GITHUB_TOKEN` doesn't have push access to the base repo's packages.
**Why it happens:** Logging in unconditionally causes a failure on fork PRs.
**How to avoid:** Gate the login step with `if: github.event_name == 'push' && github.ref == 'refs/heads/main'`.

### Pitfall 5: Concurrency Group Without `head_ref`
**What goes wrong:** All PRs share the same concurrency group, cancelling each other.
**Why it happens:** Using only `github.ref` -- PRs all resolve to `refs/pull/N/merge` pattern but `github.ref` differs per PR. However, `github.head_ref` gives the branch name which is unique per PR.
**How to avoid:** Use `${{ github.head_ref || github.ref }}` to get branch name for PRs, ref for pushes.

### Pitfall 6: Race Detector Needs CGO on Some Platforms
**What goes wrong:** `go test -race` may fail if CGO is disabled.
**Why it happens:** The race detector requires CGO. On ubuntu-latest with setup-go, CGO is enabled by default, so this is not an issue -- but be aware if switching runners.
**How to avoid:** Don't set `CGO_ENABLED=0` in the test job (only needed in Dockerfile for static binary).

## Code Examples

### Complete `.golangci.yml` (v2 format)

```yaml
# Source: https://golangci-lint.run/docs/configuration/file/
version: "2"

linters:
  default: standard
  enable:
    - gocyclo
    - godot
    - misspell
    - noctx

  settings:
    gocyclo:
      min-complexity: 15
    godot:
      scope: toplevel
    errcheck:
      check-type-assertions: true
```

### Docker Metadata for latest + SHA Tags

```yaml
# Source: https://github.com/docker/metadata-action
- uses: docker/metadata-action@v6
  id: meta
  with:
    images: ghcr.io/${{ github.repository }}
    tags: |
      type=raw,value=latest,enable=${{ github.ref == 'refs/heads/main' }}
      type=sha,prefix=
```

The `type=sha,prefix=` generates a tag like `abc1234` (short git SHA, no prefix). The `type=raw,value=latest` adds the `latest` tag, enabled only on main.

### go-version-file Usage

```yaml
# Source: https://github.com/actions/setup-go
- uses: actions/setup-go@v6
  with:
    go-version-file: go.mod  # Reads "go 1.26.1" from go.mod
```

This avoids hardcoding the Go version in the workflow -- it stays in sync with `go.mod` automatically.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| golangci-lint v1 config | v2 config with `version: "2"` | March 2025 | Must use new format; v1 configs not parseable by v2 binary |
| `gosimple` as separate linter | Merged into `staticcheck` | golangci-lint v2.0 | Don't enable `gosimple` separately |
| `actions/setup-go` + `actions/cache` | `actions/setup-go` built-in caching | setup-go v5 (2024) | Remove separate cache step |
| `docker/login-action@v3` | `docker/login-action@v4` | 2025 | Node.js runtime update |
| `actions/checkout@v4` | `actions/checkout@v6` | Jan 2026 | Node.js runtime update |

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) + `go test` |
| Config file | None (stdlib, no config needed) |
| Quick run command | `go test ./...` |
| Full suite command | `go test -race ./...` |

### Phase Requirements -> Test Map

This phase creates CI configuration files (YAML), not Go code. Validation is primarily structural.

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CI-01 | Workflow file is valid YAML | smoke | `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"` or `yq . .github/workflows/ci.yml > /dev/null` | N/A (YAML file) |
| CI-02 | golangci-lint config is valid | smoke | `golangci-lint config verify` (requires golangci-lint installed) | N/A (config file) |
| CI-03 | Existing tests still pass | regression | `go test -race ./...` | Yes (9 test files) |
| CI-04 | Linter passes on existing code | regression | `golangci-lint run ./...` | N/A |
| CI-05 | Docker image still builds | smoke | `docker build -t karaclean:test .` | N/A |

### Sampling Rate
- **Per task commit:** `go test ./... && golangci-lint run ./...`
- **Per wave merge:** Full: `go test -race ./... && golangci-lint run ./... && docker build -t karaclean:test .`
- **Phase gate:** All existing tests pass, linter passes, Docker builds successfully

### Wave 0 Gaps
- [ ] Install golangci-lint locally: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.3`
- [ ] Fix any existing lint violations before CI enforces the linter set

## Open Questions

1. **Existing lint violations**
   - What we know: The codebase was written without golangci-lint. The extras (gocyclo, godot, misspell, noctx) may flag existing code.
   - What's unclear: How many violations exist in the current codebase.
   - Recommendation: Run golangci-lint locally before committing the CI config. Fix violations in the same PR or adjust linter settings (e.g., raise gocyclo threshold).

2. **`github.repository` casing for GHCR**
   - What we know: GHCR requires lowercase image names. `github.repository` preserves the case of the GitHub repo name.
   - What's unclear: Whether the repo name contains uppercase characters.
   - Recommendation: The module path is `github.com/lm/karaclean` (all lowercase), so this should be fine. But if needed, pipe through a lowercase transform.

## Sources

### Primary (HIGH confidence)
- [golangci-lint official docs - Configuration File](https://golangci-lint.run/docs/configuration/file/) - v2 config format
- [golangci-lint official docs - Linters](https://golangci-lint.run/docs/linters/) - standard linter set
- [golangci-lint-action GitHub](https://github.com/golangci/golangci-lint-action) - action v9, requires setup-go
- [actions/setup-go GitHub](https://github.com/actions/setup-go) - v6, built-in caching
- [GitHub Docs - Publishing Docker images](https://docs.github.com/en/actions/publishing-packages/publishing-docker-images) - GHCR workflow pattern

### Secondary (MEDIUM confidence)
- [golangci-lint v2 migration guide](https://golangci-lint.run/docs/product/migration-guide/) - v1 to v2 changes
- [golangci-lint releases](https://github.com/golangci/golangci-lint/releases) - v2.11.3 current as of 2026-03-10
- [docker/metadata-action GitHub](https://github.com/docker/metadata-action) - v6, tag generation

### Tertiary (LOW confidence)
- None -- all findings verified with official sources.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all actions verified against official repos and recent releases
- Architecture: HIGH - standard Go CI pattern, well-documented by GitHub and Docker
- Pitfalls: HIGH - v2 migration issues well-documented; permissions documented by GitHub
- golangci-lint v2 config: HIGH - verified with official docs

**Research date:** 2026-03-19
**Valid until:** 2026-04-19 (stable domain, action versions change slowly)
