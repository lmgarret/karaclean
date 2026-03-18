# Architecture Research

**Domain:** Go rule-engine Docker sidecar (bookmark cleanup daemon)
**Researched:** 2026-03-18
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Entry Point (cmd/)                       │
│  ┌──────────┐  ┌──────────┐                                    │
│  │  main.go  │  │  flags   │                                    │
│  └─────┬─────┘  └─────┬────┘                                   │
│        └───────┬───────┘                                        │
├────────────────┼────────────────────────────────────────────────┤
│                ▼          Orchestration Layer                    │
│  ┌──────────────────┐  ┌──────────────────┐                     │
│  │    Scheduler      │  │   Config Loader  │                    │
│  │  (robfig/cron)    │  │   (YAML parse)   │                    │
│  └────────┬─────────┘  └────────┬─────────┘                    │
│           │                      │                              │
│           ▼                      ▼                              │
│  ┌──────────────────────────────────────────┐                   │
│  │              Run Coordinator              │                  │
│  │  (fetches bookmarks, applies rules,       │                  │
│  │   executes actions, logs results)         │                  │
│  └────────────────┬─────────────────────────┘                   │
├───────────────────┼─────────────────────────────────────────────┤
│                   ▼           Domain Layer                       │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                │
│  │   Rules    │  │  Matcher   │  │   Actions   │               │
│  │  (parsed   │  │ (evaluate  │  │ (archive,   │               │
│  │   config)  │  │  bookmark  │  │  delete)    │               │
│  └────────────┘  │  vs rule)  │  └──────┬─────┘               │
│                  └────────────┘         │                       │
├────────────────────────────────────────┼────────────────────────┤
│                                        ▼   API Client Layer     │
│  ┌─────────────────────────────────────────────────────┐        │
│  │              Karakeep API Client                      │       │
│  │  (bookmarks, highlights, lists, tags, feeds)          │       │
│  └───────────────────────┬─────────────────────────────┘        │
│                          │ HTTP + Bearer Token                   │
└──────────────────────────┼──────────────────────────────────────┘
                           ▼
                 ┌──────────────────┐
                 │  Karakeep Server  │
                 │  /api/v1/*        │
                 └──────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| **main.go** | Parse flags, load config, wire dependencies, start scheduler or single run | `cmd/karaclean/main.go` -- thin, no business logic |
| **Config Loader** | Read YAML file, validate schema, produce typed config struct | `internal/config/` -- `gopkg.in/yaml.v3` unmarshal + custom validation |
| **Scheduler** | Run the coordinator on a cron schedule, handle graceful shutdown | `internal/scheduler/` -- wraps `robfig/cron/v3` |
| **Run Coordinator** | Orchestrate a single cleanup run: fetch, match, act, report | `internal/engine/` -- the main loop |
| **Rules (model)** | Typed representation of user-defined rules from YAML | `internal/config/` -- part of config types |
| **Matcher** | Evaluate whether a bookmark matches a rule's conditions and exceptions | `internal/engine/matcher.go` -- pure functions, highly testable |
| **Actions** | Execute archive/delete via API client, respect dry-run flag | `internal/engine/actions.go` |
| **API Client** | HTTP client wrapping all Karakeep REST endpoints needed | `internal/karakeep/` -- typed methods per endpoint |

## Recommended Project Structure

```
karaclean/
├── cmd/
│   └── karaclean/
│       └── main.go              # Entry point: flags, config load, wire, run
├── internal/
│   ├── config/
│   │   ├── config.go            # Config struct definitions
│   │   ├── loader.go            # YAML loading and validation
│   │   └── config_test.go       # Config parsing edge cases
│   ├── karakeep/
│   │   ├── client.go            # Client struct, auth, base HTTP logic
│   │   ├── bookmarks.go         # List, Get, Patch, Delete bookmarks
│   │   ├── highlights.go        # Get highlights for a bookmark
│   │   ├── lists.go             # Get lists for a bookmark
│   │   ├── feeds.go             # List RSS feeds
│   │   └── client_test.go       # Tests with httptest server
│   ├── engine/
│   │   ├── runner.go            # Single-run orchestrator (fetch -> match -> act)
│   │   ├── matcher.go           # Rule evaluation: conditions + exceptions
│   │   ├── actions.go           # Archive/delete execution, dry-run support
│   │   ├── enricher.go          # Fetch supplemental data (highlights, lists)
│   │   ├── runner_test.go       # Integration-style tests with mock client
│   │   └── matcher_test.go      # Unit tests for every condition type
│   └── scheduler/
│       ├── scheduler.go         # Cron wrapper, graceful shutdown
│       └── scheduler_test.go
├── configs/
│   └── example.yaml             # Example configuration for users
├── Dockerfile
├── docker-compose.example.yml
├── go.mod
├── go.sum
└── Makefile
```

### Structure Rationale

- **cmd/karaclean/:** Single binary entry point. Thin -- just wiring. If we ever add a second command (e.g., `karaclean validate` for config checking), it lives here.
- **internal/config/:** Config is its own package because it defines types used across the codebase (rule structs, schedule, credentials). Separate from engine because config loading is a distinct concern from rule evaluation.
- **internal/karakeep/:** API client is isolated behind an interface so the engine can be tested with mocks. Organized by resource (bookmarks, highlights, lists) matching the REST API structure.
- **internal/engine/:** The core domain logic. Runner orchestrates, matcher evaluates, actions execute. No HTTP knowledge -- receives an API client interface.
- **internal/scheduler/:** Thin wrapper around robfig/cron. Separated so `main.go` can choose between scheduled mode and single-run mode without importing cron.

## Architectural Patterns

### Pattern 1: Interface-Based API Client

**What:** Define a `KarakeepAPI` interface in the engine package. The real client in `internal/karakeep/` implements it. Tests supply a mock.

**When to use:** Always -- this is the primary seam for testing the engine without hitting a real server.

**Trade-offs:** One more interface to maintain, but the testability payoff is enormous. The engine package never imports `internal/karakeep/` directly.

**Example:**
```go
// internal/engine/api.go
type KarakeepAPI interface {
    ListBookmarks(ctx context.Context, opts ListOptions) ([]Bookmark, string, error)
    GetBookmarkHighlights(ctx context.Context, id string) ([]Highlight, error)
    GetBookmarkLists(ctx context.Context, id string) ([]List, error)
    ArchiveBookmark(ctx context.Context, id string) error
    DeleteBookmark(ctx context.Context, id string) error
    ListFeeds(ctx context.Context) ([]Feed, error)
}
```

### Pattern 2: Condition Evaluator as Pure Functions

**What:** Each rule condition type (age, source, archived, favourited, tags, highlights, notes, list membership, count-per-feed) is a pure function: `func(bookmark, condition) bool`. The matcher composes them.

**When to use:** Always -- this is how you make the rule engine testable without any I/O.

**Trade-offs:** Some conditions require enriched data (highlights, lists) which means the bookmark struct must be "enriched" before matching. This is an explicit enrichment step in the runner, not hidden in the matcher.

**Example:**
```go
// internal/engine/matcher.go
func matchesCondition(b *EnrichedBookmark, c Condition) bool {
    switch c.Field {
    case "age":
        return time.Since(b.CreatedAt) > c.Duration
    case "source":
        return b.Source == c.Value
    case "favourited":
        return b.Favourited == c.Bool
    case "hasHighlights":
        return (len(b.Highlights) > 0) == c.Bool
    // ...
    }
}

func matchesRule(b *EnrichedBookmark, r Rule) bool {
    // All conditions must match
    for _, c := range r.Conditions {
        if !matchesCondition(b, c) {
            return false
        }
    }
    // No exception must match
    for _, e := range r.Exceptions {
        if matchesCondition(b, e) {
            return false
        }
    }
    return true
}
```

### Pattern 3: Enrichment Before Evaluation

**What:** The Karakeep `GET /bookmarks` endpoint returns basic fields but not highlights or list membership. An enrichment step fetches supplemental data only when rules require it.

**When to use:** When rules reference `hasHighlights`, `hasNotes`, `listMembership`, or `countPerFeed`. The runner inspects which conditions are used across all rules and only fetches what is needed.

**Trade-offs:** Adds N+1 API calls per bookmark that needs enrichment. Mitigated by: (1) only enriching when rules require it, (2) batching where possible, (3) rate limiting to be a good citizen.

**Example:**
```go
// internal/engine/enricher.go
type enrichmentNeeds struct {
    Highlights bool
    Lists      bool
    Feeds      bool // for count-per-feed, fetched once
}

func analyzeRules(rules []Rule) enrichmentNeeds {
    // Scan all conditions to determine what supplemental data is needed
}

func enrichBookmark(ctx context.Context, api KarakeepAPI, b *Bookmark, needs enrichmentNeeds) (*EnrichedBookmark, error) {
    eb := &EnrichedBookmark{Bookmark: *b}
    if needs.Highlights {
        h, err := api.GetBookmarkHighlights(ctx, b.ID)
        // ...
        eb.Highlights = h
    }
    if needs.Lists {
        l, err := api.GetBookmarkLists(ctx, b.ID)
        // ...
        eb.Lists = l
    }
    return eb, nil
}
```

### Pattern 4: Two-Phase Cleanup (Archive-then-Delete)

**What:** Rules produce two distinct action types. Archive rules run first, producing newly-archived bookmarks. Delete rules only target already-archived bookmarks with sufficient retention age. This gives users a grace period.

**When to use:** This is the core business model from PROJECT.md. Every run executes archive rules first, then delete rules.

**Trade-offs:** Means a bookmark won't be deleted on the same run it gets archived (by design). Users who want immediate deletion can set retention to 0.

## Data Flow

### Single Run Flow

```
Config (YAML)
    │
    ▼
Parse & Validate Config
    │
    ▼
Analyze Rules → determine enrichment needs
    │
    ▼
Phase 1: ARCHIVE PASS
    │
    ├─→ Fetch bookmarks (archived=false, paginated)
    │       │
    │       ▼
    │   For each bookmark:
    │       ├─→ Enrich (if needed: highlights, lists)
    │       ├─→ Match against archive rules
    │       └─→ If matched & not excepted → queue for archive
    │
    ├─→ Execute archives (or log in dry-run)
    │
Phase 2: DELETE PASS
    │
    ├─→ Fetch bookmarks (archived=true, paginated)
    │       │
    │       ▼
    │   For each bookmark:
    │       ├─→ Enrich (if needed)
    │       ├─→ Match against delete rules (including retention age)
    │       └─→ If matched & not excepted → queue for delete
    │
    ├─→ Execute deletes (or log in dry-run)
    │
    ▼
Log summary (archived: N, deleted: M, skipped: K, errors: E)
```

### Daemon Mode Flow

```
main.go
    │
    ├─→ Load config
    ├─→ Validate config
    ├─→ Create API client
    ├─→ Create runner
    │
    ├─→ if --once flag: runner.Run() then exit
    │
    └─→ else: scheduler.Start(cron_expr, runner.Run)
            │
            ├─→ robfig/cron calls runner.Run on schedule
            │
            └─→ SIGTERM/SIGINT → scheduler.Stop() → wait for
                 in-progress run → exit 0
```

### Key Data Flows

1. **Config to Rules:** YAML file is parsed into typed `Config` struct containing `[]Rule`. Each rule has `[]Condition`, `[]Exception`, and an `Action` (archive or delete). Rules are validated at startup -- fail fast on bad config.

2. **API to Enriched Bookmarks:** Raw bookmarks from `GET /bookmarks` are wrapped in `EnrichedBookmark` structs with optional highlights and list data. Enrichment is lazy based on what rules need.

3. **Matcher to Action Queue:** The matcher produces a list of `(bookmarkID, action)` pairs. The action executor processes this queue sequentially, respecting dry-run mode.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| Small instance (< 1K bookmarks) | No issues. Single paginated fetch, enrichment is fast. Runs complete in seconds. |
| Medium instance (1K-10K bookmarks) | Pagination handles this. Enrichment API calls (N+1 for highlights/lists) become the bottleneck. Consider parallel enrichment with bounded concurrency (e.g., 5 goroutines). |
| Large instance (10K+ bookmarks) | Add progress logging. Consider incremental runs (only process bookmarks newer than last run). Rate-limit API calls to avoid overloading Karakeep. |

### Scaling Priorities

1. **First bottleneck: Enrichment API calls.** If a user has 5000 bookmarks and rules check highlights, that is 5000 additional API calls. Fix: bounded-concurrency goroutine pool, and only enrich bookmarks that pass basic conditions first (filter by age/source before enriching).
2. **Second bottleneck: Memory for large bookmark sets.** Fix: process bookmarks in pages rather than loading all into memory. The cursor-based pagination from Karakeep naturally supports this -- process each page before fetching the next.

## Anti-Patterns

### Anti-Pattern 1: God Struct for Bookmark

**What people do:** Put every possible field (highlights, lists, feed info, reading progress) into a single Bookmark struct that is always fully populated.

**Why it's wrong:** Forces unnecessary API calls. Makes it unclear which fields are populated. Leads to nil pointer panics when a field was not enriched.

**Do this instead:** Use a base `Bookmark` struct for API response data and a separate `EnrichedBookmark` that wraps it with optional enrichment fields. Make enrichment explicit in the type system.

### Anti-Pattern 2: Evaluating Rules Inside the API Client

**What people do:** Put rule-matching logic in the API fetch loop to "save a pass over the data."

**Why it's wrong:** Couples API client to domain logic. Makes the API client untestable in isolation. Makes rule logic untestable without HTTP mocks.

**Do this instead:** Keep the API client dumb (fetch/mutate only). Keep the matcher pure (no I/O). The runner coordinates between them.

### Anti-Pattern 3: Relying on Karakeep Filtering Instead of Local Matching

**What people do:** Try to push all conditions to the API query parameters to avoid fetching unnecessary bookmarks.

**Why it's wrong:** The Karakeep API only supports `archived` and `favourited` filters. Everything else (age, source, tags, highlights) must be evaluated locally. Building a partial-push/partial-local system adds complexity for marginal gain.

**Do this instead:** Use API filters for the obvious cases (archived=false for archive pass, archived=true for delete pass). Do all other matching locally. The simplicity is worth the extra data transfer.

### Anti-Pattern 4: No Graceful Shutdown

**What people do:** Just `os.Exit(0)` or let the container be killed mid-run.

**Why it's wrong:** A half-completed run may have archived some bookmarks but not others, leaving the system in an inconsistent state. More importantly, a DELETE in progress could leave the API in a bad state.

**Do this instead:** Listen for SIGTERM/SIGINT. Signal the scheduler to stop accepting new runs. Wait for the current run to complete (with a timeout). Then exit.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Karakeep REST API | HTTP client with Bearer token auth, cursor pagination | Base URL from config. All endpoints under `/api/v1/`. Token passed as `Authorization: Bearer <token>` header. |
| Docker runtime | Container receives config via volume mount, env vars for overrides | Config path: env `KARACLEAN_CONFIG` or default `/config/karaclean.yaml`. API key can come from env `KARACLEAN_API_KEY` to avoid putting secrets in YAML. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| main -> config | Direct function call: `config.Load(path)` returns `(*Config, error)` | Fail fast: if config is invalid, exit 1 with clear error message before starting scheduler. |
| main -> scheduler | `scheduler.New(cronExpr, runFunc)` then `.Start()` / `.Stop()` | Scheduler owns the cron lifecycle. Runner is injected as a function. |
| scheduler -> engine | `runner.Run(ctx)` called per tick | Context carries cancellation for graceful shutdown. |
| engine -> karakeep | Via `KarakeepAPI` interface | Engine never imports the karakeep package. Interface defined in engine package. Dependency injected from main. |
| engine -> config | Engine receives `[]Rule` from config, not the full config struct | Engine does not know about cron schedule, API credentials, or file paths. |

## Build Order (Suggested Implementation Phases)

The components have clear dependency ordering:

```
Phase 1: Config + API Client (independent, can be built in parallel)
    ├── internal/config/    -- YAML types, loader, validation
    └── internal/karakeep/  -- HTTP client, all needed endpoints

Phase 2: Engine Core (depends on Phase 1 types)
    ├── internal/engine/matcher.go   -- rule evaluation (pure logic)
    ├── internal/engine/enricher.go  -- supplemental data fetching
    └── internal/engine/actions.go   -- archive/delete execution

Phase 3: Runner + Scheduler (depends on Phase 2)
    ├── internal/engine/runner.go    -- orchestrates a single run
    └── internal/scheduler/          -- cron wrapper

Phase 4: Entry Point + Docker (depends on everything)
    ├── cmd/karaclean/main.go        -- wire everything together
    ├── Dockerfile                   -- multi-stage build
    └── docker-compose.example.yml
```

**Why this order:**
- Config and API client define the types that the engine operates on.
- The matcher is pure logic with no dependencies beyond types -- easiest to test, highest value to get right.
- The runner depends on the matcher and API client interface.
- The scheduler is trivial once the runner exists.
- main.go is pure wiring -- last to build, first to change if architecture shifts.

## Sources

- [Go official module layout guidance](https://go.dev/doc/modules/layout)
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout) -- widely referenced, though not official
- [robfig/cron v3](https://pkg.go.dev/github.com/robfig/cron/v3) -- standard Go cron library
- [Martin Fowler on Rules Engines](https://martinfowler.com/bliki/RulesEngine.html)
- [Rules Engine Design Pattern](https://deviq.com/design-patterns/rules-engine-pattern/)
- [Go REST API client best practices](https://medium.com/@cep21/go-client-library-best-practices-83d877d604ca)
- [No-nonsense Go project layout](https://laurentsv.com/blog/2024/10/19/no-nonsense-go-package-layout.html)
- Karakeep OpenAPI spec: `karakeep-upstream/packages/open-api/karakeep-openapi-spec.json`

---
*Architecture research for: Go rule-engine Docker sidecar (Karaclean)*
*Researched: 2026-03-18*
