# Phase 7: Run Orchestrator and Observability - Context

**Gathered:** 2026-03-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Wire the complete rule evaluation run: paginate all bookmarks into memory, evaluate all rules against each bookmark (collect-then-act ordering), execute mutations, and produce a structured run summary. This is the glue phase that connects all previously built components into a working CLI. Cron scheduling and daemon mode are Phase 8.

</domain>

<decisions>
## Implementation Decisions

### Multi-rule matching semantics
- **First-match-wins**: evaluation stops after the first rule that matches a bookmark
- Remaining rules are not evaluated for that bookmark (silent skip — no log output)
- Rule order in YAML config determines priority; user controls this intentionally
- Avoids double-action edge cases (archive then delete in same run causing 404)

### Run summary counters
- Two separate counters for non-actioned bookmarks (not combined into a single "skipped"):
  - `no_match` — bookmarks that matched no rule's conditions
  - `excepted` — bookmarks that matched a rule's conditions but were protected by an `unless` clause
- Full summary: `archived`, `deleted`, `no_match`, `excepted`, `errors`
- Summary format: Claude's Discretion — multi-line log block recommended for readability; use `log.Printf` consistent with established pattern

### Orchestrator location
- `internal/engine/run.go` — exported `Run(ctx context.Context, api KarakeepAPI, rules []config.Rule, dryRun bool) RunSummary` function
- `RunSummary` is a typed struct (Archived, Deleted, NoMatch, Excepted, Errors int fields) — serializable to JSON for future UI handlers with no refactoring
- Testable via mock `KarakeepAPI`, consistent with established interface-driven design

### main.go wiring
- Phase 7 wires `main.go` to call `engine.Run()` on startup: config → auth → Run() → print summary → exit
- Enables manual invocation and single-run testing before Phase 8 adds the cron loop
- Phase 8 wraps this in a scheduler; Phase 7 establishes the single-run path

### Claude's Discretion
- Exact log format for the run summary block (e.g., `=== Run Summary ===` header vs inline line)
- Whether `RunSummary` has a `String()` method or is logged field by field
- Internal orchestrator loop structure (range bookmarks, range rules, break on first match)
- Error handling if `ListBookmarks` fails (fail-fast is appropriate: no bookmarks = no run)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §OBS-01 — structured log summary: archived N, deleted M, skipped K, errors E

### Existing engine (must read before planning)
- `internal/engine/actions.go` — `ExecuteAction(ctx, api, action, bookmarkID, ruleName, dryRun) ActionResult`; `ActionResult.Err` is non-nil on failure
- `internal/engine/api.go` — `KarakeepAPI` interface; `ListBookmarks`, `ArchiveBookmark`, `DeleteBookmark`
- `internal/engine/matcher.go` — `MatchesConditions(b, conditions, runTime)` and `MatchesExceptions(b, exceptions)` — call both per bookmark per rule
- `internal/engine/bookmark.go` — `Bookmark` domain struct
- `cmd/karaclean/main.go` — current startup wiring (config → auth → stub); Phase 7 adds `engine.Run()` call after auth

### Config types
- `internal/config/config.go` — `Rule` struct (Name, Conditions, Unless, Action); `Config.Rules []Rule`

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `engine.MatchesConditions` + `engine.MatchesExceptions` — already handle nil conditions/exceptions safely; combine as: match if `MatchesConditions && !MatchesExceptions`
- `engine.ExecuteAction` — returns `ActionResult` with `Err` field; orchestrator accumulates error count from non-nil `Err`
- `engine.KarakeepAPI.ListBookmarks` — already handles cursor pagination internally; returns `[]Bookmark`

### Established Patterns
- Interface-driven with mock-based tests (engine package tests use mock `KarakeepAPI`)
- `log.Printf` for operational output; `fmt.Fprintf(os.Stderr, ...)` + `os.Exit(1)` for startup failures
- Log-and-continue for per-item errors (Phase 6); fail-fast for startup/collection errors
- `time.Now()` captured once per run and passed to `MatchesConditions` as `runTime` for consistency

### Integration Points
- `internal/engine/run.go` (new) — `Run()` calls `ListBookmarks`, loops bookmarks×rules (first-match), calls `ExecuteAction`, accumulates `RunSummary`
- `cmd/karaclean/main.go` — after `client.CheckAuth(ctx)`, call `engine.Run(ctx, client, cfg.Rules, dryRun)` and log the summary; then exit 0 (or exit 1 if errors > 0 is desired — Claude's Discretion)

</code_context>

<specifics>
## Specific Ideas

- `RunSummary` struct should be JSON-serializable (exported fields, no unexported types) — user has UI plans; `engine.Run()` stays callable from a future HTTP handler with no refactoring
- The two-counter approach (`no_match` + `excepted`) gives users visibility into whether their rules are actually matching anything vs being suppressed by exceptions

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 07-run-orchestrator-and-observability*
*Context gathered: 2026-03-18*
