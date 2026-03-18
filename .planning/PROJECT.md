# Karaclean

## What This Is

Karaclean is a Docker sidecar for the Karakeep bookmark manager that automatically archives and deletes bookmarks based on user-defined rules. It targets users with high-volume RSS feeds who need automated garbage collection to prevent database bloat. Users define rules in a YAML config file; Karaclean runs on a cron schedule and enforces them via Karakeep's HTTP API.

## Core Value

Users can define flexible, declarative cleanup rules that keep their Karakeep instance lean without ever touching bookmarks they care about.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] User can configure rules in a YAML file with conditions and actions
- [ ] Rules can match on: age, source (rss/web/etc), archived status, favourited status, tags, presence of highlights, presence of notes, list membership, reading progress, count per feed
- [ ] Rules support exception conditions (e.g. delete archived, unless favourited or has highlights)
- [ ] Actions include: archive (set archived=true) and delete (permanent)
- [ ] Two-phase cleanup: auto-archive first, then auto-delete archived items after a retention period
- [ ] Cron schedule is defined in the YAML config
- [ ] Sidecar authenticates to Karakeep via API bearer token
- [ ] Works as a Docker Compose sidecar or standalone container
- [ ] Single-user (one API key per instance)
- [ ] Dry-run mode to preview what would be archived/deleted without making changes

### Out of Scope

- Web UI — deferred to a later milestone; start with YAML
- Multi-user support — single API key per container instance for now
- Direct database access — HTTP API only to avoid coupling to Karakeep internals
- Push notifications / alerting — out of scope for v1

## Context

- **Submodule:** `karakeep-upstream/` contains Karakeep's source for API reference; do not modify it
- **Karakeep API:** REST at `/api/v1`, Bearer token auth, cursor-based pagination
- **Relevant bookmark fields via API:** `id`, `createdAt`, `archived`, `favourited`, `source` (rss/web/api/mobile/extension/cli/import), `tags[]`, `note`, `summary`, highlights endpoint, lists endpoint, reading progress fields (`readingProgressPercent`)
- **Archive action:** `PATCH /v1/bookmarks/{id}` with `{ "archived": true }`
- **Delete action:** `DELETE /v1/bookmarks/{id}`
- **List bookmarks:** `GET /v1/bookmarks` — supports `archived`/`favourited` filters; cursor pagination; highlights and list membership require separate calls
- **Feeds API:** `GET /v1/feeds` — lists RSS feeds; bookmarks from feeds have `source: "rss"`
- High-volume RSS use case: tens to hundreds of new entries per day per feed, easily filling the DB within weeks

## Constraints

- **Tech stack**: Go — single binary, minimal container footprint, no runtime dependencies
- **Integration**: Karakeep HTTP API only — no direct DB access
- **Config**: YAML file mounted into the container at a configurable path
- **Deployment**: Docker container; docker-compose or standalone `docker run`

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|------------|
| Go runtime | Single binary, tiny image, no runtime deps | — Pending |
| HTTP API over direct DB | Decoupled from Karakeep internals, survives upgrades | — Pending |
| YAML config (UI later) | Fastest to ship, power-user friendly, UI is additive | — Pending |
| Archive-then-delete pattern | Mirrors Karakeep's native archive feature; gives users a grace period | — Pending |
| Single-user per instance | Simplifies auth and rule scoping; multi-user is additive | — Pending |

---
*Last updated: 2026-03-18 after initialization*
