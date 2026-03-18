# Project Research Summary

**Project:** Karaclean
**Domain:** Go Docker sidecar — bookmark garbage collection rule engine for Karakeep
**Researched:** 2026-03-18
**Confidence:** HIGH

## Executive Summary

Karaclean is a single-purpose Docker sidecar daemon that applies user-defined YAML rules to a Karakeep bookmark collection, archiving and deleting items based on age, source, tags, and status conditions. This is a niche but well-understood problem class: retention automation. Analogues exist in RSS readers (FreshRSS, Miniflux) and email filter systems, and the patterns from those domains translate directly. The recommended implementation is a pure Go binary with minimal external dependencies — two third-party libraries (go.yaml.in/yaml/v3 for config parsing and netresearch/go-cron for scheduling), everything else from the standard library. The result is a sub-10 MiB Docker image built on a scratch base that runs entirely without a runtime.

The architecture is a layered rule engine: a Config Loader parses and validates YAML, a Karakeep API Client handles all REST interactions behind an interface, an Engine evaluates rules against enriched bookmark data, and a Scheduler drives periodic runs with graceful shutdown. The critical design insight is that collect-then-act must be the data flow pattern — paginate all bookmarks into memory first, evaluate rules, then execute mutations. Interleaving reads and writes with cursor-based pagination causes bookmarks to be skipped or double-processed. All seven critical pitfalls identified in research are addressable with known techniques; none require novel engineering.

The primary risk is accidental mass deletion: Karakeep's `DELETE` is permanent with no undo. This risk is fully mitigated by three mandatory Phase 1 features: dry-run mode (skip all mutations, log intent), archive-before-delete as the default workflow (a grace period between archive and delete), and per-run deletion caps (halt if a single run would delete more than a configurable threshold). Config validation with strict unknown-field rejection is equally critical, because Go's YAML parser silently ignores misspelled keys, which can turn a protection condition into a no-op. Ship these safety features before any destructive capability reaches users.

## Key Findings

### Recommended Stack

The stack is intentionally minimal. Go's standard library covers HTTP client, JSON, logging, signal handling, and flag parsing. Only two external libraries are needed: `go.yaml.in/yaml/v3` (v3.0.4, the maintained successor to the now-archived `gopkg.in/yaml.v3`) and `github.com/netresearch/go-cron` (v0.13.1, a maintained fork of the abandoned `robfig/cron` with fixes for TZ parsing panics and Go 1.25+ support). The minimum Go version is 1.25 driven by the cron library; the recommended build toolchain is Go 1.26.x. The Docker image uses a multi-stage build with a `scratch` runtime stage — copy the static binary, CA certificates, and timezone data only.

**Core technologies:**
- Go 1.25+ (build with 1.26.x): language runtime — single static binary, zero runtime deps, tiny Docker image
- `go.yaml.in/yaml/v3` v3.0.4: YAML config parsing — maintained successor to archived gopkg.in/yaml.v3
- `github.com/netresearch/go-cron` v0.13.1: cron scheduling — maintained fork fixing known panic bugs in robfig/cron
- `log/slog` (stdlib): structured logging — JSON/text output built-in since Go 1.21, no external logger needed
- `net/http` (stdlib): Karakeep API client — production-grade HTTP client, no framework needed for a REST consumer
- `encoding/json` (stdlib): API payload encode/decode — API responses are small, stdlib is sufficient

**What to avoid:**
- `gopkg.in/yaml.v3` — archived and unmaintained since April 2025
- `robfig/cron/v3` — unmaintained since 2020 with known panic bugs in TZ parsing
- `go.yaml.in/yaml/v4` — still in RC (v4.0.0-rc.4), not production-ready
- Alpine base image — scratch is smaller and has no unnecessary attack surface for a static binary

### Expected Features

The feature landscape is clear. Every retention system (FreshRSS, Miniflux, email filters) converges on the same table stakes: age-based conditions, source-based conditions, protection signals (favourited, tags), and two-phase archive-then-delete. No direct competitor exists as a Karakeep sidecar; Karaclean has an opportunity to deliver per-feed retention policies that even FreshRSS and Miniflux lack.

**Must have (table stakes — v1):**
- YAML config parsing with strict validation (unknown fields rejected, fail fast on startup)
- `olderThan` age condition — primary filter for garbage collection
- `source` condition (rss, web, extension, etc.) — target high-volume RSS bookmarks
- `archived` and `favourited` status conditions — enable two-phase workflow
- `hasTag` / `lacksTag` conditions — tags are embedded in list response (no extra API calls)
- `unless` exception conditions — protection clauses; safety-critical for user trust
- Archive action (`PATCH /v1/bookmarks/{id}` with `archived: true`)
- Delete action (`DELETE /v1/bookmarks/{id}`)
- Dry-run mode — mandatory for a deletion tool; must block all mutations including archives
- Per-run deletion cap — halt if single run would delete more than N bookmarks
- Cron scheduling with explicit timezone support
- Structured JSON logging + per-run summary (archived: N, deleted: M, skipped: K, errors: E)

**Should have (differentiators — v1.x after validation):**
- RSS feed-scoped rules (`rssFeedId`) — per-feed retention policies; unique competitive differentiator
- AND/OR logical combinators — complex multi-condition rules
- Count-based retention (`keepNewest: N`) — cap bookmarks per feed rather than age-based
- Bookmark type conditions (`type: link/text/asset`)
- `--validate` CLI flag — config validation without running rules
- Note presence condition (`hasNote: true`) — cheap engagement signal (field on bookmark object)
- Tag-based actions (add tag before archive/delete for audit trail)
- Rule priority / `stopAfterMatch` semantics

**Defer (v2+):**
- Highlight presence conditions — requires per-bookmark API call (N+1 cost)
- List membership conditions — same N+1 cost as highlights
- Web UI — violates sidecar simplicity; only if YAML proves too high-friction
- Reading progress conditions — blocked by Karakeep (tRPC-only, not in REST API)

### Architecture Approach

The architecture follows a clean layered pattern: Entry Point wires dependencies and handles flags; Config Loader parses and validates YAML; Scheduler (cron wrapper) drives periodic runs; Run Coordinator orchestrates a single run using a two-phase collect-then-act loop; Matcher evaluates pure condition functions against enriched bookmark data; API Client exposes a typed interface that the engine never imports directly (dependency injected). The key architectural insight is the `KarakeepAPI` interface defined in the engine package — this is the primary test seam, enabling the entire engine to be tested with mocks without any real HTTP calls.

**Major components:**
1. `cmd/karaclean/main.go` — thin entry point: parse flags, load config, wire dependencies, start scheduler or single run
2. `internal/config/` — YAML types, loader with `KnownFields(true)` strict parsing, semantic validation
3. `internal/karakeep/` — HTTP client behind `KarakeepAPI` interface; organized by resource (bookmarks, highlights, feeds)
4. `internal/engine/` — matcher (pure functions), enricher (lazy supplemental data fetching), actions (archive/delete with dry-run support), runner (orchestrates single run)
5. `internal/scheduler/` — thin cron wrapper with graceful shutdown on SIGTERM/SIGINT

**Build order:** Config + API Client (independent, parallizable) → Engine Core (matcher, enricher, actions) → Runner + Scheduler → Entry Point + Docker.

### Critical Pitfalls

1. **Irreversible hard deletes** — Karakeep DELETE is permanent, no undo. Prevention: mandatory dry-run before first live run, archive-before-delete as default workflow, configurable per-run deletion cap (default 50). Ship these in Phase 1 before any destructive capability.

2. **Pagination mutation race** — Paginating and mutating concurrently causes bookmarks to be skipped or double-processed due to cursor invalidation. Prevention: collect-then-act pattern is mandatory from day one. Paginate all bookmarks into memory, evaluate rules, then execute mutations as a separate phase.

3. **Silent YAML misconfiguration** — Go YAML parsers silently ignore unknown keys, turning a misspelled protection condition into a no-op and potentially matching everything. Prevention: use `KnownFields(true)` strict parsing on `yaml.Decoder`, reject configs with unknown fields, fail loudly at startup.

4. **Rule conflict resolution ambiguity** — A bookmark matching multiple rules needs a deterministic resolution strategy. Prevention: exception-first evaluation (protection always wins), then most-conservative-action-wins (keep > archive > delete). Must be designed before the first rule is evaluated; hard to change without breaking user configs.

5. **Cron timezone and DST edge cases** — Docker defaults to UTC; users think in local time. Prevention: require explicit `timezone` field in YAML config, pass to cron via `cron.WithLocation()`, default to UTC with a startup log message if unspecified.

## Implications for Roadmap

Based on research, the architecture research itself suggests a clean four-phase build order driven by component dependencies. The pitfall research reinforces that safety features must be foundational, not retrofitted.

### Phase 1: Foundation and Safety Infrastructure
**Rationale:** Config loading and the API client define the types that everything else operates on. Safety features (dry-run, deletion caps, strict config validation, auth checking) must exist before any destructive capability reaches users — they are table stakes, not polish.
**Delivers:** Working YAML config loader with strict validation; Karakeep API client with full type coverage; dry-run mode; auth startup check; per-run deletion cap; structured logging scaffold.
**Addresses:** YAML config validation (table stakes), structured logging (table stakes), dry-run mode (table stakes), auth failure handling.
**Avoids:** Silent YAML misconfiguration (Pitfall 4), irreversible hard deletes (Pitfall 1), auth failures mid-run (Pitfall 7).

### Phase 2: Rule Engine Core
**Rationale:** With types and API client in place, the matcher and enricher can be built as pure, heavily-tested logic. The conflict resolution strategy must be designed here before any action is ever applied to real data.
**Delivers:** Condition evaluator (age, source, archived, favourited, tags, `unless` exceptions), enricher with lazy supplemental data fetching, rule conflict resolution (exception-first + most-conservative-action-wins).
**Addresses:** Age condition, source condition, archived/favourited conditions, tag conditions, exception conditions (all table stakes).
**Avoids:** Rule conflict resolution ambiguity (Pitfall 3), N+1 enrichment (Pitfall 6 — lazy enrichment from the start).

### Phase 3: Actions, Runner, and Two-Phase Cleanup
**Rationale:** Actions depend on the matcher (to know what to act on) and the API client (to perform mutations). The runner orchestrates the full collect-then-act loop that avoids the pagination race condition.
**Delivers:** Archive action, delete action with deletion cap enforcement, collect-then-act run loop, two-phase archive-pass + delete-pass execution, per-run summary reporting.
**Addresses:** Archive action (table stakes), delete action (table stakes), two-phase workflow (table stakes), run summary (table stakes).
**Avoids:** Pagination mutation race (Pitfall 2), irreversible hard deletes (Pitfall 1 — caps enforced here).

### Phase 4: Scheduler, Docker, and Release
**Rationale:** The scheduler is trivial once the runner exists. Docker packaging is the final step. This phase makes the tool production-deployable.
**Delivers:** Cron scheduler with explicit timezone support, graceful SIGTERM/SIGINT shutdown, multi-stage Dockerfile with scratch base, docker-compose example, example YAML config.
**Addresses:** Cron scheduling (table stakes), structured deployment.
**Avoids:** Cron timezone and DST edge cases (Pitfall 5), overlapping concurrent runs.

### Phase 5: Differentiators and Polish (v1.x)
**Rationale:** After validating the MVP with real users, add the features that differentiate Karaclean from generic retention tools. Feed-scoped rules are the highest-value differentiator and should lead this phase.
**Delivers:** RSS feed-scoped rules, AND/OR logical combinators, count-based retention (`keepNewest: N`), bookmark type conditions, `--validate` CLI flag, note presence condition, tag-based actions.
**Addresses:** All P2 features from FEATURES.md.
**Uses:** Existing Karakeep API Client feeds endpoint; extends the matcher with new condition types.

### Phase Ordering Rationale

- Config and API Client come first because they define the types the entire engine operates on. Building them first enables parallel development of engine components.
- Safety features (dry-run, caps, strict validation) are Phase 1 because they must exist before any destructive capability. Retrofitting safety onto a working delete tool is harder and riskier than building it in from the start.
- The rule engine core (Phase 2) is built before the runner (Phase 3) because the matcher is pure logic — the most testable component, highest value to get right before introducing I/O.
- The scheduler (Phase 4) is trivial once the runner exists; it is last because it is a thin wrapper that can be replaced without touching business logic.
- Differentiators (Phase 5) are post-MVP because the core value proposition (reliable, safe bookmark cleanup) must be proven before investing in feed-scoped rules or count-based retention.

### Research Flags

Phases with well-documented patterns (skip research-phase during planning):
- **Phase 1:** Config loading and HTTP client patterns are standard Go. Stack is fully specified with exact versions.
- **Phase 3:** Two-phase collect-then-act and archive/delete API calls are straightforward given the OpenAPI spec.
- **Phase 4:** Multi-stage Docker builds and cron scheduling with timezone are well-documented.

Phases likely needing deeper research during planning:
- **Phase 2:** Rule conflict resolution semantics and the enricher's lazy evaluation logic have subtle edge cases. Worth a planning research pass on "Go rules engine patterns" to validate the most-conservative-action approach against alternatives before coding.
- **Phase 5:** Feed-scoped rules require understanding the `GET /v1/feeds` endpoint response shape and how `rssFeedId` filters interact with pagination. Quick API validation pass before implementation.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All library versions verified on pkg.go.dev and GitHub as of 2026-03-18. Deprecation status of gopkg.in/yaml.v3 and robfig/cron confirmed from primary sources. |
| Features | HIGH | Derived from Karakeep source code directly (OpenAPI spec, shared types, tRPC routers). API constraints verified against live source. Competitor analysis from official docs. |
| Architecture | HIGH | Standard Go project layout and rules engine patterns are well-established. Build order derived from component dependency graph, not speculation. |
| Pitfalls | HIGH | Critical pitfalls sourced from Karakeep source code analysis (confirmed hard delete, cursor pagination behavior, page size limits). YAML silent-drop behavior is a known Go ecosystem footgun with documented mitigations. |

**Overall confidence:** HIGH

### Gaps to Address

- **Karakeep API rate limiting on reads:** Source code confirms rate limiting exists on bookmark creation, but read/delete rate limits were not found. If a large collection triggers undocumented rate limits during enrichment, the bounded-concurrency semaphore (Phase 5) may need to be moved to Phase 2 or 3. Monitor during development.
- **`go.yaml.in/yaml/v4` timeline:** v4.0.0-rc.4 is the current release. If v4 reaches stable before Phase 1 is complete, evaluate migration. If not, stay on v3.0.4 — migration is straightforward when v4 stabilizes.
- **Karakeep API version stability:** All endpoints are under `/api/v1`. The hardcoded API version is an acceptable tradeoff for now. Track Karakeep upstream releases for any v2 signals before Phase 5.
- **Per-run deletion cap default:** Research recommends a default of 50 deletions per run, but this is a judgment call. Validate with real users during MVP testing. Too low causes incomplete cleanup on first run; too high loses the safety benefit.

## Sources

### Primary (HIGH confidence)
- Karakeep OpenAPI source: `karakeep-upstream/packages/open-api/lib/bookmarks.ts` — bookmark fields, filter params, pagination
- Karakeep shared types: `karakeep-upstream/packages/shared/types/bookmarks.ts` — bookmark type definitions
- Karakeep tRPC routers: `karakeep-upstream/packages/trpc/routers/bookmarks.ts` — confirmed hard delete, cursor pagination, readingProgress is tRPC-only
- `go.yaml.in/yaml/v3` v3.0.4 on pkg.go.dev — verified version and maintainer status
- `go.yaml.in/yaml/v4` v4.0.0-rc.4 on pkg.go.dev — verified pre-release status
- `github.com/netresearch/go-cron` v0.13.1 on GitHub and pkg.go.dev — verified version and Go 1.25+ requirement
- `github.com/robfig/cron` on GitHub — confirmed unmaintained since 2020
- Go slog official blog post (go.dev) — stdlib structured logging since Go 1.21
- Go release history (go.dev) — Go 1.26.1 confirmed current stable

### Secondary (MEDIUM confidence)
- FreshRSS configuration docs — retention feature comparison
- Miniflux configuration parameters — cleanup frequency and starred-entry behavior
- Miniflux per-feed retention policy issue #770 — confirmed feature gap (not supported)
- Go REST API client best practices — interface-based client design
- golang-standards/project-layout — cmd/ and internal/ conventions
- Alpine vs distroless vs scratch analysis — Docker image strategy
- Handling Timezone Issues in Cron Jobs (2025 guide) — DST skip/double-fire behavior
- REST API pagination and race condition patterns — collect-then-act rationale

### Tertiary (LOW confidence)
- Go YAML silent zero-value bug (buildsoftwaresystems.com) — mapstructure ErrorUnused mitigation (needs validation against KnownFields approach)
- Go ecosystem trends 2025 (JetBrains) — stdlib-first philosophy confirmation

---
*Research completed: 2026-03-18*
*Ready for roadmap: yes*
