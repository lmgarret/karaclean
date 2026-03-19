# Phase 10: CI — Context

**Gathered:** 2026-03-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Set up GitHub Actions CI that automatically runs tests, lints Go code, and builds (and pushes) the Docker image. Covers workflow file(s), golangci-lint config, and registry integration. Release management, versioned tags, and deployment are out of scope.

</domain>

<decisions>
## Implementation Decisions

### CI Platform
- GitHub Actions (`.github/workflows/`)
- Runner: `ubuntu-latest` with `actions/setup-go` (not a Docker container job)
- Cache go modules (`~/go/pkg/mod`) and build cache (`~/.cache/go-build`) between runs

### Lint Tooling
- Use `golangci-lint` (not bare `gofmt`/`go vet`)
- Commit a `.golangci.yml` config file to pin linter set and version
- Linter set: **standard + extras**
  - Standard: `errcheck`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `unused`
  - Extras: `gocyclo` (complexity), `godot` (comment style), `misspell` (typos), `noctx` (missing context)

### Docker Build Scope
- Build image on all triggers (PRs and push to main) to verify it compiles
- Push to `ghcr.io` only on merge to `main`
- Tags on push: `ghcr.io/<owner>/karaclean:latest` + `ghcr.io/<owner>/karaclean:<git-sha>`
- Registry auth: use `GITHUB_TOKEN` (auto-available in GitHub Actions — no extra secrets needed)

### Trigger Configuration
- Triggers: `push` to `main` + `pull_request` targeting `main`
- No push-to-all-branches trigger
- Concurrency groups: cancel in-progress runs when a new push arrives on the same branch/PR (saves CI minutes)

### Job Structure
- Separate jobs for: `test`, `lint`, `docker`
- `docker` job depends on `test` and `lint` passing (fail fast)
- Tests: `go test ./...` with `-race` flag

</decisions>

<canonical_refs>
## Canonical References

No external specs — requirements are fully captured in decisions above.

### Project files to read
- `Dockerfile` — existing multi-stage build (alpine builder → scratch final); CI must build this exact file
- `go.mod` — Go version to pin in `actions/setup-go` (currently `go 1.26.1`)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `Dockerfile` — already written, multi-stage (alpine builder → scratch). CI builds this as-is.
- All `*_test.go` files across `cmd/karaclean/`, `internal/config/`, `internal/engine/`, `internal/duration/`, `internal/karakeep/` — `go test ./...` covers all packages.

### Established Patterns
- No existing CI infrastructure — fully greenfield (no `.github/`, no Makefile)
- Module path: `github.com/lm/karaclean` — used for `ghcr.io` image name derivation

### Integration Points
- `.github/workflows/ci.yml` — new file to create
- `.golangci.yml` — new file to create at repo root
- `GITHUB_TOKEN` — auto-injected by GitHub Actions for ghcr.io push; no manual secret setup needed

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard GitHub Actions patterns.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 10-ci-run-tests-lint-and-build-docker-image*
*Context gathered: 2026-03-19*
