# Karaclean

Go Docker sidecar for Karakeep bookmark cleanup. Declarative YAML rules, cron-scheduled, safety-first design.

## Pre-Commit Checklist

1. **Documentation**: When features are added or modified, update documentation (README.md, docs/karaclean.example.yaml) in the same commit or PR. Do not defer docs to a follow-up.

2. **Lint**: Run `~/go/bin/golangci-lint run ./...` before committing. Fix all lint errors. Config is in `.golangci.yml` (v2 format, standard linters plus gocyclo, godot, misspell, noctx).

3. **Tests**: Run `go test -race ./...` before committing. Fix all test failures.

## Project Conventions

- Config parsing uses `go.yaml.in/yaml/v3` (maintained fork, NOT gopkg.in)
- Pointer types for optional config fields (nil vs zero-value distinction)
- `KnownFields(true)` for strict YAML parsing -- no custom UnmarshalYAML unless required for compatibility
- Case-sensitive matching throughout
- `Validate()` returns `[]ValidationError` for caller flexibility

## Tech Stack

- Go (module: github.com/lmgarret/karaclean)
- golangci-lint v2.11 (binary at ~/go/bin/golangci-lint)
- Docker multi-stage build with scratch final image, multi-arch (amd64 + arm64)
- CI: GitHub Actions (golangci-lint-action@v9, go test -race)

## Releases

- Tag-driven: pushing a `vX.Y.Z` git tag runs `.github/workflows/release.yml`.
- Docker tags published per release: `X.Y.Z`, `X.Y`, `X`, and `latest` (= newest
  release). Main-branch builds publish `edge` + commit `<sha>` (see `ci.yml`).
- Changelog is AI-generated from commits via `nosovj/llm-release-action`; requires an
  `ANTHROPIC_API_KEY` repository secret. No `CHANGELOG.md` is committed -- the changelog
  lives in GitHub Releases.
- Version is stamped into the binary via `-ldflags` (`main.version/commit/date`),
  exposed by `--version`. Keep the README Image Tags table in sync with tag changes.
