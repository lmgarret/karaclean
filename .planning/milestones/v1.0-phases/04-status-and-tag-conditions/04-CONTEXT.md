# Phase 4: Status and Tag Conditions - Context

**Gathered:** 2026-03-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Extend `MatchesConditions` in `internal/engine/matcher.go` with four new condition checks: `archived`, `favourited`, `hasTag`, and `lacksTag`. Config types (`Conditions` struct) and `Bookmark` fields are already in place from prior phases — this phase is matcher logic, validation additions, and tests only. No schema changes needed.

</domain>

<decisions>
## Implementation Decisions

### Tag matching semantics
- **Case-sensitive** — tags matched exactly as stored; `"read-later"` ≠ `"Read-Later"`
- **Set membership (ANY)** — `hasTag: "X"` matches if any tag in `Bookmark.Tags` equals `"X"`; same inverse for `lacksTag`
- Mirrors expected Karakeep behavior; simpler and more predictable than case-folding

### Validation additions
- `hasTag` and `lacksTag` values must be **non-empty strings** — empty string is a user error, reject at config load with a clear validation message (consistent with Phase 1 fail-fast philosophy)
- **No contradiction check** — `hasTag: X` + `lacksTag: X` on the same rule is not caught; the rule simply never matches (AND semantics). Not worth the validation complexity for an unlikely edge case.

### Carrying forward from Phase 3
- AND semantics — all non-nil conditions must pass; short-circuit on first mismatch
- Extend `MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time) bool` — same function, same signature
- Tests required alongside all implementation (project-level requirement)

### Claude's Discretion
- Test table structure and helper utilities
- Order of condition checks within `MatchesConditions` (e.g., cheapest checks first)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §COND-03 — `archived: true/false` condition
- `.planning/REQUIREMENTS.md` §COND-04 — `favourited: true/false` condition
- `.planning/REQUIREMENTS.md` §COND-05 — `hasTag` condition
- `.planning/REQUIREMENTS.md` §COND-06 — `lacksTag` condition

### Existing codebase (must read before planning)
- `internal/engine/matcher.go` — current `MatchesConditions` implementation; Phase 4 extends this function
- `internal/engine/bookmark.go` — `Bookmark` type: `Archived bool`, `Favourited bool`, `Tags []string` — these are the fields matched in this phase
- `internal/config/config.go` — `Conditions` struct already has `Archived *bool`, `Favourited *bool`, `HasTag *string`, `LacksTag *string` — no schema changes needed
- `internal/config/validate.go` — add non-empty string validation for `HasTag` and `LacksTag`
- `internal/engine/matcher_test.go` — existing test file; Phase 4 adds cases here

### Prior phase context
- `.planning/phases/03-age-and-source-conditions/03-CONTEXT.md` — matcher API shape, AND semantics, runTime injection pattern

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/engine/matcher.go`: existing `MatchesConditions` — Phase 4 adds four `if c.X != nil { ... }` blocks following the same pattern
- `internal/config/validate.go`: `ValidationError` type and collected-error pattern — reuse for `hasTag`/`lacksTag` empty-string errors
- `internal/engine/matcher_test.go`: existing table-driven test structure — extend with new condition cases

### Established Patterns
- Pointer fields for optional conditions (`nil` = absent, skip check)
- Short-circuit: `return false` on first condition mismatch, `return true` at end
- Tag list from `Bookmark.Tags []string` — iterate and compare with `==` (case-sensitive)

### Integration Points
- `internal/engine/matcher.go` — the only file changed for matcher logic
- `internal/config/validate.go` — add empty-string guards for `HasTag` and `LacksTag`
- `internal/engine/matcher_test.go` — extend test table with archived, favourited, hasTag, lacksTag, and combined cases

</code_context>

<specifics>
## Specific Ideas

No specific references — standard set-membership semantics for tag matching.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 04-status-and-tag-conditions*
*Context gathered: 2026-03-18*
