# Phase 4: Status and Tag Conditions - Research

**Researched:** 2026-03-18
**Domain:** Go condition matching -- boolean fields and string set membership
**Confidence:** HIGH

## Summary

Phase 4 extends the existing `MatchesConditions` function in `internal/engine/matcher.go` with four new condition checks: `archived` (bool), `favourited` (bool), `hasTag` (string set membership), and `lacksTag` (inverse set membership). All data types, config struct fields, and the matching function signature already exist from prior phases -- this is purely additive logic.

The implementation follows the exact same pattern used for `olderThan` and `source` in Phase 3: nil-check the pointer field, compare against the bookmark, return false on mismatch. Validation adds two empty-string guards for `hasTag` and `lacksTag` in `validate.go`. No new packages, no schema changes, no new files.

**Primary recommendation:** Add four `if c.X != nil` blocks to `MatchesConditions` following the existing pattern, add two validation checks in `Validate()`, and extend both test tables with comprehensive cases.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Tag matching is **case-sensitive** -- exact string equality, no case-folding
- Set membership is **ANY** semantics -- `hasTag: "X"` matches if any element in `Bookmark.Tags` equals `"X"`; `lacksTag` is the inverse
- `hasTag` and `lacksTag` values must be **non-empty strings** -- reject at config validation
- **No contradiction check** -- `hasTag: X` + `lacksTag: X` on same rule is allowed (rule simply never matches)
- AND semantics carried forward from Phase 3 -- all non-nil conditions must pass
- Same function signature: `MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time) bool`
- Tests required alongside all implementation

### Claude's Discretion
- Test table structure and helper utilities
- Order of condition checks within `MatchesConditions` (e.g., cheapest checks first)

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| COND-03 | Rules can match on archived status (`archived: true/false`) | Bool pointer comparison in matcher; `Bookmark.Archived bool` and `Conditions.Archived *bool` already exist |
| COND-04 | Rules can match on favourited status (`favourited: true/false`) | Identical pattern to COND-03; `Bookmark.Favourited bool` and `Conditions.Favourited *bool` already exist |
| COND-05 | Rules can match bookmarks that have a specific tag (`hasTag`) | Linear scan of `Bookmark.Tags []string` with exact match; `Conditions.HasTag *string` exists |
| COND-06 | Rules can match bookmarks that lack a specific tag (`lacksTag`) | Inverse of COND-05; `Conditions.LacksTag *string` exists |
</phase_requirements>

## Standard Stack

No new dependencies. This phase uses only Go standard library and existing project packages.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib | 1.24+ | Boolean comparison, string iteration | No external deps needed for simple logic |

### Supporting
Already in project from prior phases -- no additions needed.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Linear scan for tag matching | `map[string]struct{}` set | Bookmarks typically have <10 tags; linear scan is simpler and faster for small N; no pre-processing needed |
| Case-sensitive match | `strings.EqualFold` | User decision: case-sensitive is locked |

## Architecture Patterns

### Existing Project Structure (no changes)
```
internal/
  config/
    config.go        # Conditions struct (already has all fields)
    validate.go      # Add hasTag/lacksTag empty-string checks
    validate_test.go # Extend validation test table
  engine/
    bookmark.go      # Bookmark type (already has all fields)
    matcher.go       # Add 4 condition check blocks
    matcher_test.go  # Extend matcher test table
```

### Pattern 1: Nil-Pointer Condition Check (established in Phase 3)
**What:** Each optional condition is a pointer field. Nil means "skip this check." Non-nil means "must match."
**When to use:** Every condition in `MatchesConditions`.
**Example (existing pattern from matcher.go):**
```go
if c.Source != nil {
    if b.Source != *c.Source {
        return false
    }
}
```

### Pattern 2: Boolean Condition Check
**What:** Compare dereferenced `*bool` against bookmark's bool field.
**When to use:** `archived` and `favourited` conditions.
**Example:**
```go
if c.Archived != nil {
    if b.Archived != *c.Archived {
        return false
    }
}
```

### Pattern 3: Tag Set Membership
**What:** Linear scan over `[]string` looking for exact match.
**When to use:** `hasTag` condition.
**Example:**
```go
if c.HasTag != nil {
    found := false
    for _, tag := range b.Tags {
        if tag == *c.HasTag {
            found = true
            break
        }
    }
    if !found {
        return false
    }
}
```

### Pattern 4: Tag Set Non-Membership (Inverse)
**What:** Linear scan over `[]string`; match fails if tag IS found.
**When to use:** `lacksTag` condition.
**Example:**
```go
if c.LacksTag != nil {
    for _, tag := range b.Tags {
        if tag == *c.LacksTag {
            return false
        }
    }
}
```

### Recommended Check Order in MatchesConditions
Cheapest checks first for short-circuit efficiency:
1. `Archived` -- single bool comparison (O(1))
2. `Favourited` -- single bool comparison (O(1))
3. `Source` -- single string comparison (O(1), already exists)
4. `HasTag` -- linear scan over tags slice (O(n))
5. `LacksTag` -- linear scan over tags slice (O(n))
6. `OlderThan` -- duration parse + time comparison (already exists, slightly more expensive)

Note: The existing function has `OlderThan` first, then `Source`. Reordering is Claude's discretion per CONTEXT.md. Recommendation: insert the four new checks between `Source` and the final `return true`, keeping existing checks in their current positions to minimize diff noise. Alternatively, reorder all six for optimal short-circuit. Either approach is valid.

### Anti-Patterns to Avoid
- **Extracting a helper `containsTag` function too early:** With only two call sites (`hasTag` and `lacksTag`), inlining keeps the logic visible and avoids indirection. A helper can be extracted later if more tag operations appear.
- **Using `slices.Contains` from Go 1.21+:** While available, the project has used manual loops throughout. Stay consistent with the codebase style unless the team decides to adopt `slices` broadly.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| N/A | N/A | N/A | This phase is pure conditional logic -- no complex problems to solve |

**Key insight:** This phase is intentionally simple. The architecture decisions (pointer fields, AND semantics, short-circuit) were all made in Phases 1 and 3. Phase 4 just fills in the remaining condition slots.

## Common Pitfalls

### Pitfall 1: Forgetting nil Tags Slice
**What goes wrong:** If `Bookmark.Tags` is nil (no tags), `hasTag` should not match and `lacksTag` should always match. A nil slice in Go has length 0 and `range` over it produces zero iterations, so the correct behavior falls out naturally. But a developer might add a nil guard that accidentally inverts the logic.
**Why it happens:** Defensive coding instinct.
**How to avoid:** Trust Go's range-over-nil-slice behavior. Write explicit test cases for bookmarks with nil/empty tags.
**Warning signs:** Seeing `if b.Tags == nil` guards before the range loop.

### Pitfall 2: Validation Not Covering Empty String
**What goes wrong:** `hasTag: ""` or `lacksTag: ""` passes validation, then silently matches nothing (no tag has an empty name) or everything (no tag equals empty string, so `lacksTag` always passes).
**Why it happens:** Forgetting that `*string` being non-nil doesn't mean the string is non-empty.
**How to avoid:** Add explicit `if *c.HasTag == ""` checks in `Validate()`. Already a locked decision.
**Warning signs:** Missing validation test cases for empty string values.

### Pitfall 3: Accidentally Using Pointer Equality
**What goes wrong:** Writing `b.Archived != c.Archived` instead of `b.Archived != *c.Archived` for bool conditions.
**Why it happens:** `c.Archived` is `*bool`, `b.Archived` is `bool` -- Go compiler will catch this (type mismatch), so this is actually prevented at compile time. Not a real risk, just mentioning for completeness.

### Pitfall 4: Validation Field Path Typos
**What goes wrong:** Validation error messages use wrong field path (e.g., `conditions.hastag` instead of `conditions.hasTag`).
**Why it happens:** Manual string construction.
**How to avoid:** Follow the exact pattern from existing validation: `prefix + ".conditions.hasTag"`.

## Code Examples

Verified patterns from the existing codebase:

### Validation Pattern (from validate.go line 88-93)
```go
// Existing source validation pattern to follow:
if rule.Conditions.Source != nil && !contains(validSources, *rule.Conditions.Source) {
    errs = append(errs, ValidationError{
        Field:   prefix + ".conditions.source",
        Message: fmt.Sprintf("invalid value %q (...)", *rule.Conditions.Source),
    })
}
```

### New Validation for HasTag/LacksTag
```go
// Add after existing olderThan validation block:
if rule.Conditions.HasTag != nil && *rule.Conditions.HasTag == "" {
    errs = append(errs, ValidationError{
        Field:   prefix + ".conditions.hasTag",
        Message: "must not be empty",
    })
}
if rule.Conditions.LacksTag != nil && *rule.Conditions.LacksTag == "" {
    errs = append(errs, ValidationError{
        Field:   prefix + ".conditions.lacksTag",
        Message: "must not be empty",
    })
}
```

### Test Helper Already Available
```go
// In config_test.go (config package):
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

// In matcher_test.go (engine package):
func strPtr(s string) *string { return &s }
// NOTE: boolPtr does NOT exist yet in matcher_test.go -- must add it
```

### Test Case Structure (existing pattern from matcher_test.go)
```go
{
    name:     "archived true matches archived bookmark",
    bookmark: engine.Bookmark{ID: "x", CreatedAt: runTime, Archived: true},
    conds:    &config.Conditions{Archived: boolPtr(true)},
    want:     true,
},
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| N/A | N/A | N/A | No relevant changes -- Go bool/string comparison is timeless |

This phase involves no evolving APIs or libraries. Pure Go logic.

## Open Questions

None. All decisions are locked, all types exist, and the implementation pattern is fully established.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | N/A (Go convention) |
| Quick run command | `go test ./internal/engine/ -run TestMatchesConditions -v` |
| Full suite command | `go test ./...` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| COND-03 | archived condition matches/rejects | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | Exists (extend) |
| COND-04 | favourited condition matches/rejects | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | Exists (extend) |
| COND-05 | hasTag condition matches/rejects | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | Exists (extend) |
| COND-06 | lacksTag condition matches/rejects | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | Exists (extend) |
| COND-03-06 | all conditions combined AND | unit | `go test ./internal/engine/ -run TestMatchesConditions -v` | Exists (extend) |
| VAL-hasTag | empty string hasTag rejected | unit | `go test ./internal/config/ -run TestValidate -v` | Exists (extend) |
| VAL-lacksTag | empty string lacksTag rejected | unit | `go test ./internal/config/ -run TestValidate -v` | Exists (extend) |

### Sampling Rate
- **Per task commit:** `go test ./internal/engine/ ./internal/config/ -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] Add `boolPtr` helper to `internal/engine/matcher_test.go` -- needed for archived/favourited test cases (already exists in config test package but packages are separate)

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/engine/matcher.go`, `internal/engine/bookmark.go`, `internal/config/config.go`, `internal/config/validate.go` -- all read and verified
- Existing tests: `internal/engine/matcher_test.go`, `internal/config/validate_test.go` -- all read and verified passing
- Phase 4 CONTEXT.md -- locked decisions and canonical references

### Secondary (MEDIUM confidence)
None needed -- this phase is entirely internal logic with no external dependencies.

### Tertiary (LOW confidence)
None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new deps, pure Go stdlib
- Architecture: HIGH -- extending established patterns with identical structure
- Pitfalls: HIGH -- verified against actual codebase; nil-slice and empty-string are the main risks, both well-understood

**Research date:** 2026-03-18
**Valid until:** Indefinite -- this research covers Go fundamentals and established project patterns that will not change
