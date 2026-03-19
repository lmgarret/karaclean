# Phase 7: Run Orchestrator and Observability - Research

**Researched:** 2026-03-18
**Domain:** Go application wiring -- collect-then-act orchestration pattern, structured logging
**Confidence:** HIGH

## Summary

Phase 7 is a pure integration/wiring phase. All building blocks exist: `ListBookmarks` (pagination), `MatchesConditions` + `MatchesExceptions` (rule evaluation), and `ExecuteAction` (mutations with dry-run). The new code is a `Run()` function in `internal/engine/run.go` that composes these into a collect-then-act pipeline, plus wiring in `main.go` to invoke it.

No new external dependencies are needed. The entire phase uses Go stdlib (`context`, `time`, `log`, `fmt`) and the project's existing internal packages. The mock infrastructure in `engine_test` (`mockAPI` in `api_test.go`) is already fully capable of testing the orchestrator.

**Primary recommendation:** Implement `Run()` as a single exported function that takes `(ctx, api, rules, dryRun)` and returns `(RunSummary, error)`. Wire `main.go` to call it after auth. All testing uses the existing `mockAPI` test double.

<user_constraints>

## User Constraints (from CONTEXT.md)

### Locked Decisions
- **First-match-wins**: evaluation stops after the first rule that matches a bookmark. Rule order in YAML config determines priority.
- **Two separate non-action counters**: `no_match` (matched no rule) and `excepted` (matched conditions but protected by `unless`). Full summary: `archived`, `deleted`, `no_match`, `excepted`, `errors`.
- **Orchestrator location**: `internal/engine/run.go` with exported `Run(ctx context.Context, api KarakeepAPI, rules []config.Rule, dryRun bool) RunSummary`.
- **RunSummary**: typed struct with exported int fields (Archived, Deleted, NoMatch, Excepted, Errors) -- JSON-serializable for future UI handlers.
- **main.go wiring**: config -> auth -> Run() -> print summary -> exit.

### Claude's Discretion
- Exact log format for the run summary block
- Whether RunSummary has a String() method or is logged field by field
- Internal orchestrator loop structure (range bookmarks, range rules, break on first match)
- Error handling if ListBookmarks fails (fail-fast is appropriate)
- Whether exit code is 1 when errors > 0

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope.

</user_constraints>

<phase_requirements>

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| OBS-01 | Each run produces a structured log summary (archived: N, deleted: M, skipped: K, errors: E) | RunSummary struct with Archived/Deleted/NoMatch/Excepted/Errors fields; logged via String() method after Run() returns. Note: CONTEXT.md refined "skipped" into two counters (no_match + excepted) which provides more granularity than the requirement specifies -- this satisfies OBS-01. |

</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `log` | 1.26.1 | Run summary output | Consistent with all prior phases (log.Printf pattern) |
| Go stdlib `context` | 1.26.1 | Context propagation to API calls | Already used by KarakeepAPI interface |
| Go stdlib `time` | 1.26.1 | Capture runTime once per run | MatchesConditions requires runTime parameter |
| Go stdlib `fmt` | 1.26.1 | String formatting, Stringer implementation | Used in main.go for stderr errors |

### Supporting
No additional dependencies needed. This phase uses only existing internal packages:
- `internal/config` -- `Rule`, `Conditions`, `Exceptions` types
- `internal/engine` -- `KarakeepAPI`, `Bookmark`, `MatchesConditions`, `MatchesExceptions`, `ExecuteAction`

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `log.Printf` summary | `log/slog` structured logging | slog is more modern but project consistently uses `log.Printf`; switching mid-project creates inconsistency |
| Manual counter increment | `sync/atomic` counters | No concurrency in v1 (sequential loop), unnecessary complexity |

## Architecture Patterns

### Recommended Project Structure
```
internal/engine/
  run.go          # NEW: Run() orchestrator + RunSummary struct
  run_test.go     # NEW: Table-driven tests using existing mockAPI
  actions.go      # Existing: ExecuteAction
  api.go          # Existing: KarakeepAPI interface
  api_test.go     # Existing: mockAPI test double (reused by run_test.go)
  matcher.go      # Existing: MatchesConditions, MatchesExceptions
  bookmark.go     # Existing: Bookmark struct

cmd/karaclean/
  main.go         # MODIFIED: add engine.Run() call after auth
```

### Pattern 1: Collect-Then-Act
**What:** Paginate ALL bookmarks into memory first (collect phase), then evaluate rules and execute mutations (act phase). Never mutate while paginating.
**When to use:** Always for this orchestrator. Prevents pagination cursor corruption from concurrent mutations.
**Example:**
```go
func Run(ctx context.Context, api KarakeepAPI, rules []config.Rule, dryRun bool) (RunSummary, error) {
    runTime := time.Now()

    // Phase 1: Collect
    bookmarks, err := api.ListBookmarks(ctx)
    if err != nil {
        return RunSummary{}, fmt.Errorf("collecting bookmarks: %w", err)
    }
    log.Printf("collected %d bookmarks", len(bookmarks))

    // Phase 2: Evaluate + Act
    var summary RunSummary
    for _, b := range bookmarks {
        matched := false
        for _, rule := range rules {
            if !MatchesConditions(b, rule.Conditions, runTime) {
                continue
            }
            if MatchesExceptions(b, rule.Unless) {
                summary.Excepted++
                matched = true
                break // first-match-wins
            }
            result := ExecuteAction(ctx, api, rule.Action, b.ID, rule.Name, dryRun)
            if result.Err != nil {
                summary.Errors++
            } else {
                switch rule.Action {
                case "archive":
                    summary.Archived++
                case "delete":
                    summary.Deleted++
                }
            }
            matched = true
            break // first-match-wins
        }
        if !matched {
            summary.NoMatch++
        }
    }
    return summary, nil
}
```

### Pattern 2: RunSummary as Value Type with Stringer
**What:** RunSummary is a plain struct with exported int fields -- returned by value, JSON-serializable, implements `fmt.Stringer`.
**When to use:** Always. Small struct, no mutation after return.
**Example:**
```go
type RunSummary struct {
    Archived int `json:"archived"`
    Deleted  int `json:"deleted"`
    NoMatch  int `json:"no_match"`
    Excepted int `json:"excepted"`
    Errors   int `json:"errors"`
}

func (s RunSummary) String() string {
    return fmt.Sprintf("archived=%d deleted=%d no_match=%d excepted=%d errors=%d",
        s.Archived, s.Deleted, s.NoMatch, s.Excepted, s.Errors)
}
```

### Anti-Patterns to Avoid
- **Mutating during pagination:** Never call ArchiveBookmark/DeleteBookmark while ListBookmarks pagination is in progress. The collect-then-act pattern prevents this.
- **Combining no_match and excepted:** User explicitly requires two separate counters for visibility.
- **Returning error for per-bookmark failures:** Per-bookmark errors are logged by ExecuteAction and counted in summary.Errors. Only ListBookmarks failure returns an error from Run().
- **Evaluating all rules for a bookmark:** First-match-wins means `break` after the first matching rule. Do not continue evaluating.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bookmark pagination | Manual cursor loop in run.go | `api.ListBookmarks(ctx)` | Already handles cursor pagination internally |
| Rule matching | Duplicate condition checks | `MatchesConditions` + `MatchesExceptions` | Already tested, handle nil safely |
| Action execution | Direct API calls in orchestrator | `ExecuteAction` | Handles dry-run, error wrapping, logging |

**Key insight:** Phase 7 should contain zero domain logic -- it is pure orchestration glue. All domain logic already exists in matcher.go and actions.go.

## Common Pitfalls

### Pitfall 1: Forgetting runTime Consistency
**What goes wrong:** Calling `time.Now()` inside the loop means bookmarks evaluated later have a different reference time than earlier ones.
**Why it happens:** Seems natural to get "current" time per bookmark.
**How to avoid:** Capture `runTime := time.Now()` once before the bookmark loop and pass to all MatchesConditions calls.
**Warning signs:** Test assertions fail intermittently near time boundaries.

### Pitfall 2: Excepted vs NoMatch Counter Confusion
**What goes wrong:** A bookmark matches conditions but is excepted by `unless` -- if `matched` flag isn't set, it falls through to NoMatch.
**Why it happens:** The `matched = true` line is missing in the exception branch before `break`.
**How to avoid:** Both the "excepted" path and the "action executed" path must set `matched = true` and `break`.
**Warning signs:** `NoMatch + Excepted + Archived + Deleted + Errors != total bookmarks`.

### Pitfall 3: Run() Returning Error vs Summary.Errors
**What goes wrong:** Confusing the two error paths. Run() returns error on first ExecuteAction failure, aborting the entire run.
**Why it happens:** Natural Go instinct to `return err` on any error.
**How to avoid:** ListBookmarks failure returns `(RunSummary{}, err)`. Per-bookmark ExecuteAction failures increment `summary.Errors` and continue.
**Warning signs:** Run aborts after first action error instead of processing remaining bookmarks.

### Pitfall 4: Not Testing Zero-Rule and Zero-Bookmark Cases
**What goes wrong:** Panic or wrong counts when rules slice is empty or no bookmarks exist.
**Why it happens:** Only testing happy path with multiple bookmarks and rules.
**How to avoid:** Include edge case tests: empty rules (all bookmarks = no_match), empty bookmarks (summary is all zeros).

## Code Examples

### main.go Wiring (after existing auth step)
```go
// Step 5: Execute rules (single run)
summary, err := engine.Run(context.Background(), client, cfg.Rules, dryRun)
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}
log.Printf("run complete: %s", summary)
```

### Test Structure (table-driven with mockAPI)
```go
func TestRun(t *testing.T) {
    tests := []struct {
        name      string
        bookmarks []engine.Bookmark
        rules     []config.Rule
        dryRun    bool
        wantSum   engine.RunSummary
    }{
        {
            name:      "no bookmarks returns zero summary",
            bookmarks: nil,
            rules:     []config.Rule{{Name: "r1", Action: "archive"}},
            wantSum:   engine.RunSummary{},
        },
        {
            name:      "no rules means all bookmarks are no_match",
            bookmarks: []engine.Bookmark{{ID: "bk-1"}},
            rules:     nil,
            wantSum:   engine.RunSummary{NoMatch: 1},
        },
        // ... first-match-wins, excepted, error, dry-run cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockAPI{listBookmarksRet: tt.bookmarks}
            got, err := engine.Run(context.Background(), mock, tt.rules, tt.dryRun)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got != tt.wantSum {
                t.Errorf("got %+v, want %+v", got, tt.wantSum)
            }
        })
    }
}
```

### ListBookmarks Failure Test
```go
func TestRun_ListBookmarksError(t *testing.T) {
    mock := &mockAPI{listBookmarksErr: errors.New("api unreachable")}
    _, err := engine.Run(context.Background(), mock, nil, false)
    if err == nil {
        t.Fatal("expected error when ListBookmarks fails")
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Paginate-and-act interleaved | Collect-then-act | Design decision | Prevents pagination race conditions |
| Single "skipped" counter | no_match + excepted split | User decision | Better observability into rule effectiveness |
| `log` package | `log/slog` (Go 1.21+) | 2023 | Not adopting -- project uses `log.Printf` consistently |

## Open Questions

1. **Exit code on errors > 0**
   - What we know: User listed this as Claude's Discretion
   - Recommendation: Exit 0 on partial success. Errors are per-bookmark failures (e.g., one API timeout), not program failures. A non-zero exit could confuse container orchestrators in Phase 8. The error count in the summary provides visibility.

2. **Summary format: single-line vs multi-line**
   - What we know: User said Claude's Discretion, "multi-line log block recommended for readability"
   - Recommendation: Use `RunSummary.String()` in a single `log.Printf` call. The key=value format is grep-friendly and machine-parseable. A multi-line block adds visual noise for a five-field summary.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | None (Go convention) |
| Quick run command | `go test ./internal/engine/ -run TestRun -v` |
| Full suite command | `go test ./...` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| OBS-01 | Run produces summary with correct archived/deleted counts | unit | `go test ./internal/engine/ -run TestRun -v` | No -- Wave 0 |
| OBS-01 | First-match-wins stops evaluation after first matching rule | unit | `go test ./internal/engine/ -run TestRun -v` | No -- Wave 0 |
| OBS-01 | Excepted counter increments for exception-protected bookmarks | unit | `go test ./internal/engine/ -run TestRun -v` | No -- Wave 0 |
| OBS-01 | NoMatch counter for bookmarks matching no rule | unit | `go test ./internal/engine/ -run TestRun -v` | No -- Wave 0 |
| OBS-01 | Errors counter for failed actions (log-and-continue) | unit | `go test ./internal/engine/ -run TestRun -v` | No -- Wave 0 |
| OBS-01 | ListBookmarks failure returns error (fail-fast) | unit | `go test ./internal/engine/ -run TestRun -v` | No -- Wave 0 |
| OBS-01 | Dry-run mode passes through to ExecuteAction | unit | `go test ./internal/engine/ -run TestRun -v` | No -- Wave 0 |
| OBS-01 | main.go compiles with engine.Run() call | build | `go build ./cmd/karaclean/` | Existing |

### Sampling Rate
- **Per task commit:** `go test ./internal/engine/ -run TestRun -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/engine/run_test.go` -- table-driven tests for Run() covering all behaviors listed above
- No new framework install needed (Go stdlib testing)
- mockAPI already exists in `api_test.go` (same `engine_test` package, directly reusable)

## Sources

### Primary (HIGH confidence)
- Codebase: `internal/engine/actions.go` -- ExecuteAction signature, ActionResult struct
- Codebase: `internal/engine/api.go` -- KarakeepAPI interface (ListBookmarks, ArchiveBookmark, DeleteBookmark)
- Codebase: `internal/engine/matcher.go` -- MatchesConditions(b, conditions, runTime), MatchesExceptions(b, exceptions)
- Codebase: `internal/engine/bookmark.go` -- Bookmark domain struct
- Codebase: `internal/engine/api_test.go` -- mockAPI test double
- Codebase: `cmd/karaclean/main.go` -- current startup wiring
- Codebase: `internal/config/config.go` -- Rule, Conditions, Exceptions types
- Phase context: `07-CONTEXT.md` -- locked decisions and integration points

### Secondary (MEDIUM confidence)
- None needed -- pure application wiring with no external dependencies

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- Go stdlib only, no new dependencies
- Architecture: HIGH -- all building blocks exist and are verified in code; pattern defined in CONTEXT.md
- Pitfalls: HIGH -- derived from analyzing specific loop structure, counter semantics, and existing error patterns

**Research date:** 2026-03-18
**Valid until:** indefinite (Go stdlib, no external dependencies to version-drift)
