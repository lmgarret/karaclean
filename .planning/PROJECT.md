# Karaclean

## What This Is

Karaclean is a Docker sidecar for the Karakeep bookmark manager that automatically archives and deletes bookmarks based on user-defined YAML rules. Users write declarative rules with conditions (age, source, status, tags), exceptions (protect starred/noted/tagged bookmarks), and actions (archive or delete). It runs on a cron schedule as a production Docker sidecar, enforcing cleanup rules via Karakeep's HTTP API.

**v1.0 shipped 2026-03-19.** 11,168 lines of Go, 10 phases, 20 plans. All 20 v1 requirements satisfied.

## Core Value

Users can define flexible, declarative cleanup rules that keep their Karakeep instance lean without ever touching bookmarks they care about.

## Requirements

### Validated

- ✓ User can configure rules in a YAML file with conditions and actions — v1.0
- ✓ Rules can match on: age (`olderThan`), source (rss/web/api/mobile/extension/cli/import), archived status, favourited status, tags (`hasTag`/`lacksTag`) — v1.0
- ✓ Rules support exception conditions (`unless favourited`, `unless hasTag`, `unless hasNote`, `unless archived/notArchived`) — v1.0
- ✓ Actions: archive (PATCH API) and delete (DELETE API) — v1.0
- ✓ Dry-run mode previews all changes without executing mutations — v1.0
- ✓ Cron schedule defined in YAML config — v1.0
- ✓ Startup API token validation before any rules execute — v1.0
- ✓ Works as a Docker Compose sidecar or standalone container (scratch image, static binary) — v1.0
- ✓ Single-user (one API key per instance) — v1.0
- ✓ Config validation rejects unknown fields at startup (strict YAML parsing) — v1.0
- ✓ GitHub Actions CI: test with -race, golangci-lint v2, Docker build/push — v1.0

### Active

- [ ] Per-run deletion cap — halt if a single run would delete more than N bookmarks (SAFE-01)
- [ ] `--validate` CLI flag — validate config without running (TOOL-01)
- [ ] RSS feed-scoped rules — target specific feeds by ID or name (RULE-01)
- [ ] AND/OR logical combinators for multi-condition rules (RULE-02)
- [ ] Count-based retention: `keepNewest: N` per feed (RULE-03)
- [ ] `go mod tidy` — fix 4 direct deps marked as `// indirect`

### Out of Scope

- Web UI — deferred; YAML config first
- Multi-user support — single API key per container; additive if needed
- Direct database access — HTTP API only to stay decoupled from Karakeep internals
- Reading progress / highlights / list membership conditions — require N+1 API calls; defer to v2+
- Push notifications / alerting — out of scope for v1

## Context

- **Status:** v1.0 shipped. Fully functional bookmark GC sidecar.
- **Tech stack:** Go 1.24+, oapi-codegen for Karakeep client, robfig/cron v3 for scheduling, go.yaml.in/yaml/v3 for config
- **LOC:** ~11,168 Go | 121 files in v1.0
- **Submodule:** `karakeep-upstream/` contains Karakeep's source for API reference; do not modify it
- **Karakeep API:** REST at `/api/v1`, Bearer token auth, cursor-based pagination
- **Archive action:** `PATCH /v1/bookmarks/{id}` with `{ "archived": true }`
- **Delete action:** `DELETE /v1/bookmarks/{id}`
- **List bookmarks:** `GET /v1/bookmarks` — cursor pagination; highlights/lists require separate calls
- **Known debt:** 4 direct deps annotated `// indirect` in go.mod (cosmetic, fix with `go mod tidy`)

## Constraints

- **Tech stack:** Go — single binary, minimal container footprint, no runtime dependencies
- **Integration:** Karakeep HTTP API only — no direct DB access
- **Config:** YAML file mounted into the container at a configurable path
- **Deployment:** Docker container; docker-compose or standalone `docker run`
- **Tests:** Every phase includes unit tests (user requirement)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go runtime | Single binary, tiny image, no runtime deps | ✓ Good — scratch image works cleanly |
| HTTP API over direct DB | Decoupled from Karakeep internals, survives upgrades | ✓ Good — no Karakeep coupling issues |
| YAML config (UI later) | Fastest to ship, power-user friendly, UI is additive | ✓ Good |
| OpenAPI codegen (oapi-codegen) | Karakeep has complete OpenAPI spec — avoid hand-writing HTTP client boilerplate | ✓ Good — generated client clean, wrapper thin |
| `KarakeepClient` wrapper name | oapi-codegen generates `Client` in same package — collision required rename | ✓ Resolved correctly |
| Pointer types for optional fields | Distinguish nil from zero-value in config (`*int`, `*string`, `*bool`) | ✓ Good |
| Strictly-greater-than `olderThan` | Exact boundary does not match — deterministic semantics | ✓ Good |
| Fixed day counts (mo=30d, y=365d) | Deterministic for GC retention, appropriate for this use case | ✓ Good |
| Collect-then-act in engine.Run() | Prevents pagination race conditions; safe for delete operations | ✓ Good |
| `_ "time/tzdata"` in main.go | Embeds tzdata into binary for scratch image — no OS tzdata dependency | ✓ Essential for scratch |
| Tests required every phase | User requirement — all phases include unit tests alongside implementation | ✓ Complied throughout |
| golangci-lint v2 (not v1) | v2 is current; `gosimple` merged into `staticcheck`; `version: "2"` format | ✓ Correct — v1 format would have caused CI failures |
| Single-user per instance | Simplifies auth and rule scoping; multi-user is additive | ✓ Good |

---
*Last updated: 2026-03-19 after v1.0 milestone*
