# Phase 3: Age and Source Conditions - Research

**Researched:** 2026-03-18
**Domain:** Go pure-function matchers, custom duration parsing, config schema migration
**Confidence:** HIGH

## Summary

Phase 3 introduces the matcher subsystem -- pure functions that evaluate bookmark conditions without I/O. The two conditions are `olderThan` (duration-based age check) and `source` (string enum match). The phase also includes a schema migration: `OlderThan` changes from `*int` (days) to `*string` (duration string like `"30d"`, `"2w"`, `"1mo"`).

The implementation is straightforward Go: a regex-based duration parser, a top-level `MatchesConditions` function with short-circuit AND semantics, and updated validation. No external dependencies are needed. The main complexity is the ripple effect of the `*int` to `*string` type change on existing tests and testdata YAML fixtures.

**Primary recommendation:** Implement in two tasks -- (1) duration parser + validation update + schema migration, (2) matcher functions + comprehensive tests.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- `olderThan` is strictly greater than -- a bookmark created exactly N units ago does NOT match
- Reference point: `time.Now()` captured once at run start and passed as `runTime time.Time` parameter to all matchers
- `olderThan: 0h` (or any zero duration) is valid -- matches all bookmarks
- Schema change from Phase 1: `OlderThan *int` (days) -> `OlderThan *string` (duration string); validation must be updated
- Duration format: `"30d"`, `"2w"`, `"6h"`, `"1mo"`, `"1y"` -- compact, human-readable
- Supported units: `h` (hours), `d` (days), `w` (weeks), `mo` (months ~30 days), `y` (years ~365 days)
- Parse with a small internal helper; validation rejects unrecognized units at config load time
- A rule with no conditions set is rejected at config validation (already implemented in Phase 1)
- Package: `internal/engine/` -- matchers live alongside `Bookmark` and `KarakeepAPI`
- Function signature: `func MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time) bool`
- Each condition check short-circuits on first mismatch (AND semantics)

### Claude's Discretion
- Internal duration parsing helper design (regex vs split-based)
- Whether months/years use exact day counts or `time.AddDate` arithmetic
- Test table structure and helper utilities

### Deferred Ideas (OUT OF SCOPE)
- UI for rule creation and bookmark impact preview
- Minutes (`m`) unit -- excluded to avoid ambiguity with months
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| COND-01 | Rules can match bookmarks older than N days (`olderThan` condition) | Duration parser converts string to `time.Duration`, `MatchesConditions` compares `runTime.Sub(b.CreatedAt) > parsedDuration`; extended to multi-unit strings per context decisions |
| COND-02 | Rules can filter by source: rss, web, api, mobile, extension, cli, import | Source matching is simple string equality `b.Source == *c.Source`; validation already enforces valid enum values |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `time` | Go 1.26 | `time.Time`, `time.Duration`, `time.Now()` | Native duration arithmetic, no deps needed |
| Go stdlib `regexp` | Go 1.26 | Duration string parsing | Simple regex for `^\d+(h|d|w|mo|y)$` pattern |
| Go stdlib `strconv` | Go 1.26 | Parse numeric portion of duration string | `strconv.Atoi` for the digit prefix |
| Go stdlib `testing` | Go 1.26 | Table-driven unit tests | Project standard -- no test framework deps |

### Supporting
No additional libraries needed. This phase is pure Go stdlib.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom duration parser | `time.ParseDuration` | stdlib only supports ns/us/ms/s/m/h -- no days/weeks/months/years, so custom is required |
| `time.AddDate` for months/years | Fixed day multiplication (30d, 365d) | `AddDate` handles calendar months properly but creates non-deterministic durations; fixed days are simpler and match user expectations for GC |

## Architecture Patterns

### Files Changed/Created
```
internal/
  engine/
    matcher.go          # NEW: MatchesConditions, matchOlderThan, matchSource
    matcher_test.go     # NEW: comprehensive table-driven tests
    duration.go         # NEW: ParseDuration helper (or inline in matcher.go)
    duration_test.go    # NEW: duration parsing tests
  config/
    config.go           # MODIFIED: OlderThan *int -> *string
    validate.go         # MODIFIED: olderThan validation updated for duration strings
    validate_test.go    # MODIFIED: update tests for new type
    config_test.go      # MODIFIED: intPtr(30) -> strPtr("30d") in helpers
    testdata/
      valid_full.yaml   # MODIFIED: olderThan: 30 -> olderThan: "30d"
      valid_minimal.yaml # MODIFIED: same
```

### Pattern 1: Duration Parsing
**What:** A `ParseDuration(s string) (time.Duration, error)` function that converts compact strings to `time.Duration`.
**When to use:** Called during config validation (fail-fast) and during matching (convert once, compare).
**Example:**
```go
// engine/duration.go
var durationRe = regexp.MustCompile(`^(\d+)(h|d|w|mo|y)$`)

func ParseDuration(s string) (time.Duration, error) {
    m := durationRe.FindStringSubmatch(s)
    if m == nil {
        return 0, fmt.Errorf("invalid duration %q (expected format: 30d, 2w, 6h, 1mo, 1y)", s)
    }
    n, _ := strconv.Atoi(m[1]) // regex guarantees digits
    switch m[2] {
    case "h":
        return time.Duration(n) * time.Hour, nil
    case "d":
        return time.Duration(n) * 24 * time.Hour, nil
    case "w":
        return time.Duration(n) * 7 * 24 * time.Hour, nil
    case "mo":
        return time.Duration(n) * 30 * 24 * time.Hour, nil
    case "y":
        return time.Duration(n) * 365 * 24 * time.Hour, nil
    default:
        return 0, fmt.Errorf("unknown unit %q", m[2])
    }
}
```

**Design decision (Claude's discretion):** Use regex approach. It is concise, handles the full format in one pass, and the regex is simple enough to be readable. Fixed day counts (30d for months, 365d for years) are recommended over `time.AddDate` -- simpler semantics, deterministic, and appropriate for GC retention policies where calendar-exact months are overkill.

### Pattern 2: Matcher Function
**What:** A single exported `MatchesConditions` function with internal helpers.
**When to use:** Called by Phase 7 orchestrator for each bookmark.
**Example:**
```go
// engine/matcher.go
func MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time) bool {
    if c == nil {
        return true // no conditions = match all (though validation prevents this)
    }
    if c.OlderThan != nil {
        dur, _ := ParseDuration(*c.OlderThan) // validated at load time
        if runTime.Sub(b.CreatedAt) <= dur {
            return false
        }
    }
    if c.Source != nil {
        if b.Source != *c.Source {
            return false
        }
    }
    return true
}
```

### Pattern 3: Config Validation Update
**What:** Replace integer validation with duration string parsing validation.
**Example:**
```go
// In validate.go, replace the olderThan int check with:
if rule.Conditions.OlderThan != nil {
    if _, err := engine.ParseDuration(*rule.Conditions.OlderThan); err != nil {
        errs = append(errs, ValidationError{
            Field:   prefix + ".conditions.olderThan",
            Message: err.Error(),
        })
    }
}
```

**Note on import cycle:** `config` importing `engine` creates a circular dependency (engine already imports config). Two options:
1. Put `ParseDuration` in a shared package like `internal/duration/` or `internal/timeutil/`
2. Put `ParseDuration` in the `config` package itself since it is used during validation

**Recommendation:** Place `ParseDuration` in a small `internal/duration/` package. Both `config` (validation) and `engine` (matching) import it. This avoids circular deps cleanly.

### Anti-Patterns to Avoid
- **Calling `time.Now()` inside matchers:** Breaks testability and consistency. Always use injected `runTime`.
- **Parsing duration string at match time without prior validation:** The string should be validated once at config load. Matcher can assume valid format (but should still handle error gracefully).
- **Using `time.AddDate` for approximate units:** Creates inconsistency -- February months are shorter, leap years differ. Fixed multipliers are more predictable for GC.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Duration parsing | Full parser with all time units | Regex + switch on 5 known units | Only 5 units needed; stdlib `time.ParseDuration` lacks days+ |
| Time comparison | Manual epoch math | `time.Duration` and `time.Sub` | stdlib handles overflow, precision, timezone correctly |

**Key insight:** The duration parser is intentionally limited to 5 units. Don't over-engineer it into a general-purpose duration library.

## Common Pitfalls

### Pitfall 1: Import Cycle between config and engine
**What goes wrong:** `config.validate` needs to call `ParseDuration` which lives in `engine`, but `engine` imports `config` for the `Conditions` type.
**Why it happens:** Natural to put duration parsing next to the matcher.
**How to avoid:** Put `ParseDuration` in a separate small package (`internal/duration/`). Both config and engine import it.
**Warning signs:** `import cycle not allowed` compiler error.

### Pitfall 2: Off-by-one in age comparison
**What goes wrong:** Using `>=` instead of `>` causes bookmarks created exactly at the threshold to match.
**Why it happens:** Ambiguity in "older than" semantics.
**How to avoid:** Context decision is clear: strictly greater than. Use `runTime.Sub(b.CreatedAt) > dur`. Test with exact boundary values.
**Warning signs:** Test with bookmark age == threshold passes when it should fail.

### Pitfall 3: Forgetting to update Phase 1 tests and testdata
**What goes wrong:** Changing `OlderThan` from `*int` to `*string` breaks all existing tests that use `intPtr(30)`.
**Why it happens:** Type change has ripple effects across test files and YAML fixtures.
**How to avoid:** Systematically update: (1) `config.go` type, (2) `validate.go` logic, (3) `validate_test.go` helpers, (4) `config_test.go` assertions, (5) testdata YAML files.
**Warning signs:** `go test ./...` fails in `internal/config` package after type change.

### Pitfall 4: Zero duration edge case
**What goes wrong:** `olderThan: "0h"` might be rejected as invalid when it should match all bookmarks.
**Why it happens:** Natural to validate "must be positive" but context says zero is valid.
**How to avoid:** Duration validation accepts `n >= 0`. Add explicit test case for `"0h"`, `"0d"`.
**Warning signs:** Test for `olderThan: "0h"` fails validation.

### Pitfall 5: Regex not anchored or too permissive
**What goes wrong:** Duration like `"30dd"` or `"abc30d"` passes validation.
**Why it happens:** Missing `^` and `$` anchors or overly permissive character classes.
**How to avoid:** Use anchored regex: `^(\d+)(h|d|w|mo|y)$`. Test with malformed inputs.
**Warning signs:** Invalid duration strings pass validation.

### Pitfall 6: YAML type coercion for olderThan
**What goes wrong:** YAML `olderThan: 30` (without quotes) is parsed as integer, not string, causing a decode error after the type change to `*string`.
**Why it happens:** YAML auto-detects types. `30` is an integer, not a string.
**How to avoid:** YAML fixtures must quote the value: `olderThan: "30d"`. Users must also quote in their config. Document this clearly. The Go yaml decoder will reject a bare integer when target type is `*string`.
**Warning signs:** YAML decode error "cannot unmarshal int into string".

## Code Examples

### Duration Parser with Tests
```go
// internal/duration/duration.go
package duration

import (
    "fmt"
    "regexp"
    "strconv"
    "time"
)

var re = regexp.MustCompile(`^(\d+)(h|d|w|mo|y)$`)

func Parse(s string) (time.Duration, error) {
    m := re.FindStringSubmatch(s)
    if m == nil {
        return 0, fmt.Errorf("invalid duration %q (use format like 30d, 2w, 6h, 1mo, 1y)", s)
    }
    n, _ := strconv.Atoi(m[1])
    switch m[2] {
    case "h":
        return time.Duration(n) * time.Hour, nil
    case "d":
        return time.Duration(n) * 24 * time.Hour, nil
    case "w":
        return time.Duration(n) * 7 * 24 * time.Hour, nil
    case "mo":
        return time.Duration(n) * 30 * 24 * time.Hour, nil
    case "y":
        return time.Duration(n) * 365 * 24 * time.Hour, nil
    default:
        return 0, fmt.Errorf("unknown unit %q", m[2])
    }
}
```

### Matcher Function
```go
// internal/engine/matcher.go
package engine

import (
    "time"

    "github.com/lm/karaclean/internal/config"
    "github.com/lm/karaclean/internal/duration"
)

func MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time) bool {
    if c == nil {
        return true
    }
    if c.OlderThan != nil {
        dur, _ := duration.Parse(*c.OlderThan) // already validated at config load
        if runTime.Sub(b.CreatedAt) <= dur {
            return false
        }
    }
    if c.Source != nil {
        if b.Source != *c.Source {
            return false
        }
    }
    return true
}
```

### Test Pattern (table-driven with runTime injection)
```go
func TestMatchesConditions(t *testing.T) {
    runTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

    tests := []struct {
        name     string
        bookmark Bookmark
        conds    *config.Conditions
        want     bool
    }{
        {
            name: "olderThan matches",
            bookmark: Bookmark{CreatedAt: runTime.Add(-31 * 24 * time.Hour)},
            conds:    &config.Conditions{OlderThan: strPtr("30d")},
            want:     true,
        },
        {
            name: "olderThan exact boundary does not match",
            bookmark: Bookmark{CreatedAt: runTime.Add(-30 * 24 * time.Hour)},
            conds:    &config.Conditions{OlderThan: strPtr("30d")},
            want:     false,
        },
        {
            name: "source matches",
            bookmark: Bookmark{Source: "rss"},
            conds:    &config.Conditions{Source: strPtr("rss")},
            want:     true,
        },
        {
            name: "source does not match",
            bookmark: Bookmark{Source: "web"},
            conds:    &config.Conditions{Source: strPtr("rss")},
            want:     false,
        },
        {
            name: "both conditions AND -- both match",
            bookmark: Bookmark{
                CreatedAt: runTime.Add(-31 * 24 * time.Hour),
                Source:    "rss",
            },
            conds: &config.Conditions{OlderThan: strPtr("30d"), Source: strPtr("rss")},
            want:  true,
        },
        {
            name: "both conditions AND -- one fails",
            bookmark: Bookmark{
                CreatedAt: runTime.Add(-31 * 24 * time.Hour),
                Source:    "web",
            },
            conds: &config.Conditions{OlderThan: strPtr("30d"), Source: strPtr("rss")},
            want:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := MatchesConditions(tt.bookmark, tt.conds, runTime)
            if got != tt.want {
                t.Errorf("MatchesConditions() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `OlderThan *int` (days only) | `OlderThan *string` (duration string) | Phase 3 | Enables hours, weeks, months, years granularity |
| No matcher functions | `MatchesConditions` pure function | Phase 3 | Foundation for all condition/exception evaluation in Phases 4-7 |

**Deprecated/outdated:**
- `OlderThan *int`: Replaced by `*string` in this phase. All references to integer days must be migrated.

## Open Questions

1. **Large duration overflow**
   - What we know: `time.Duration` is int64 nanoseconds, max ~292 years. `olderThan: "292y"` would overflow.
   - What's unclear: Whether to explicitly cap or let natural overflow occur.
   - Recommendation: Add validation that `n` is reasonable (e.g., <= 100 for years, <= 36500 for days). Low priority -- unlikely real-world edge case.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (Go 1.26) |
| Config file | none -- Go convention, `go test` auto-discovers |
| Quick run command | `go test ./internal/engine/ ./internal/config/ ./internal/duration/` |
| Full suite command | `go test ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| COND-01 | olderThan duration matching (strict >) | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | No -- Wave 0 |
| COND-01 | Duration string parsing (all units, edge cases) | unit | `go test ./internal/duration/ -run TestParse -v` | No -- Wave 0 |
| COND-01 | olderThan validation (duration format, not int) | unit | `go test ./internal/config/ -run TestValidate -v` | Exists -- needs update |
| COND-02 | source string equality matching | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | No -- Wave 0 |
| COND-01+02 | AND composition of olderThan + source | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/engine/ ./internal/config/ ./internal/duration/`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/duration/duration.go` -- new package for `Parse` function
- [ ] `internal/duration/duration_test.go` -- covers COND-01 duration parsing
- [ ] `internal/engine/matcher.go` -- new file for `MatchesConditions`
- [ ] `internal/engine/matcher_test.go` -- covers COND-01, COND-02, and composition
- [ ] Update `internal/config/validate_test.go` -- existing tests broken by `*int` -> `*string` change
- [ ] Update `internal/config/config_test.go` -- existing tests reference `intPtr(30)`
- [ ] Update `internal/config/testdata/valid_full.yaml` -- `olderThan: 30` -> `olderThan: "30d"`
- [ ] Update `internal/config/testdata/valid_minimal.yaml` -- same

## Sources

### Primary (HIGH confidence)
- Go stdlib `time` package -- `time.Duration`, `time.Sub`, `time.Hour` constants; well-known stable API
- Go stdlib `regexp` package -- `regexp.MustCompile`, `FindStringSubmatch`; standard Go pattern
- Go stdlib `testing` package -- table-driven test pattern; standard Go convention
- Existing codebase: `internal/config/config.go`, `validate.go`, `internal/engine/bookmark.go` -- read directly

### Secondary (MEDIUM confidence)
- None needed -- this phase is pure Go stdlib with no external dependencies

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- pure Go stdlib, no external deps
- Architecture: HIGH -- function signature locked by CONTEXT.md, import cycle solution is standard Go practice
- Pitfalls: HIGH -- identified from direct codebase analysis (type change ripple, import cycle, YAML coercion)

**Research date:** 2026-03-18
**Valid until:** 2026-04-18 (stable -- Go stdlib, no moving targets)
