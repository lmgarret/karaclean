# Phase 3: Age and Source Conditions - Context

**Gathered:** 2026-03-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Implement pure matcher functions that evaluate `olderThan` and `source` conditions against `engine.Bookmark` structs. No I/O, no side effects. This phase establishes the matcher foundation — Phases 4-8 extend and call it. Also includes a schema change: `OlderThan` changes from `*int` (days only) to `*string` (duration string supporting multiple units).

</domain>

<decisions>
## Implementation Decisions

### Age threshold semantics
- `olderThan` is **strictly greater than** — a bookmark created exactly N units ago does NOT match
- Reference point: `time.Now()` captured **once at run start** and passed as `runTime time.Time` parameter to all matchers — consistent within a run, no clock drift
- `olderThan: 0h` (or any zero duration) is valid — matches all bookmarks (consistent semantics, not rejected)
- Schema change from Phase 1: `OlderThan *int` (days) → `OlderThan *string` (duration string); validation must be updated accordingly

### Duration string format
- Format: `"30d"`, `"2w"`, `"6h"`, `"1mo"`, `"1y"` — compact, human-readable
- Supported units:
  - `h` = hours
  - `d` = days
  - `w` = weeks
  - `mo` = months (~30 days — `mo` chosen over `m` to avoid minutes/months ambiguity)
  - `y` = years (~365 days)
- Parse with a small internal helper; validation rejects unrecognized units at config load time

### Nil/empty conditions behavior
- A rule with **no conditions set** is rejected at config validation — at least one condition field must be non-nil
- Consistent with the fail-fast, descriptive-error philosophy from Phases 1-2
- Prevents accidental "match everything" rules (e.g., a bare `action: delete` with no conditions)

### Matcher API shape
- **Package:** `internal/engine/` — matchers live alongside `Bookmark` and `KarakeepAPI`; no new package needed
- **Function signature:** standalone pure functions, not methods on Rule (avoids config/engine circular dependency)
  ```go
  // engine/matcher.go
  func MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time) bool
  ```
- Each condition check short-circuits on first mismatch (AND semantics — all must pass)
- Phase 4 adds its conditions to the same function; Phase 5 adds a parallel `MatchesExceptions(b Bookmark, e *config.Exceptions) bool`
- Phase 7 orchestrator calls `MatchesConditions` and (later) `MatchesExceptions` per bookmark before applying action

### Claude's Discretion
- Internal duration parsing helper design (regex vs split-based)
- Whether months/years use exact day counts or `time.AddDate` arithmetic
- Test table structure and helper utilities

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §COND-01 — "Rules can match bookmarks older than N days (`olderThan` condition)" — **note:** Phase 3 extends this to multi-unit duration strings
- `.planning/REQUIREMENTS.md` §COND-02 — "Rules can filter by source: rss, web, api, mobile, extension, cli, import"

### Existing codebase (must read before planning)
- `internal/engine/bookmark.go` — `Bookmark` type: `CreatedAt time.Time`, `Source string`; these are the fields matched in this phase
- `internal/engine/api.go` — `KarakeepAPI` interface; new file `matcher.go` will be added alongside
- `internal/config/config.go` — `Conditions` struct: `OlderThan *int` must change to `OlderThan *string`; `Source *string` stays as-is
- `internal/config/validate.go` — validation of `OlderThan` (currently validates as int) must be updated to parse and validate duration strings; also add >= 1 condition check per rule

### Project foundation
- `.planning/PROJECT.md` — Tech stack (Go, no runtime deps), key decisions (fail-fast, tests required every phase)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/config/config.go`: `Conditions` struct with pointer fields — nil = not set, already established pattern for optional conditions
- `internal/config/validate.go`: `ValidationErrors` / `ValidationError` types — reuse same error pattern for new duration validation errors
- `internal/engine/bookmark.go`: `Bookmark.CreatedAt time.Time` and `Bookmark.Source string` — directly usable in matcher logic

### Established Patterns
- Fail-fast with collected errors — all validation errors reported at once, not one at a time
- `fmt.Errorf("context: %w", err)` error wrapping
- Pointer types for optional fields (nil = absent)
- `go.yaml.in/yaml/v3` for YAML parsing (already in go.mod)
- Tests required alongside all implementation (project-level requirement)

### Integration Points
- `internal/config/validate.go` — `OlderThan` validation must be updated from int range check to duration string parse+validate; rule-level "at least one condition" check added here
- `internal/engine/matcher.go` (new file) — exports `MatchesConditions`; imported by Phase 7 orchestrator
- `internal/config/config.go` — `OlderThan *int` → `OlderThan *string` is a breaking change; Phase 1 tests referencing int values must be updated

</code_context>

<specifics>
## Specific Ideas

- Duration string format inspired by familiar Docker/cron tooling: `30d`, `2w`, `6h`, `1mo`, `1y`
- `runTime` injected as parameter enables deterministic, clock-independent unit tests (no `time.Now()` calls inside matchers)
- Future UI rule preview: pure functions with injected time make it trivial to call matchers from a future HTTP handler without side effects

</specifics>

<deferred>
## Deferred Ideas

- UI for rule creation and bookmark impact preview — mentioned as future plan; matcher design (pure functions, injected runTime) is compatible with this without further changes now
- Minutes (`m`) unit — excluded to avoid ambiguity with months; sub-hour precision not needed for bookmark GC

</deferred>

---

*Phase: 03-age-and-source-conditions*
*Context gathered: 2026-03-18*
