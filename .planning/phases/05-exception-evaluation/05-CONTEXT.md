# Phase 5: Exception Evaluation - Context

**Gathered:** 2026-03-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Implement exception evaluation: a function that evaluates the `unless` clause on a rule and returns whether a matched bookmark should be skipped. The `Exceptions` struct is already defined in `internal/config/config.go`. This phase adds the exception matcher function, validation for exception fields, and tests. No schema changes needed.

</domain>

<decisions>
## Implementation Decisions

### OR semantics (per requirements)
- Multiple exception clauses on a single rule use **OR semantics** — any single exception firing causes the bookmark to be skipped
- Short-circuit on first exception that fires (consistent with conditions short-circuiting on first mismatch)

### Function shape
- Claude's Discretion: naming (`MatchesExceptions`, `IsProtected`, etc.) and return polarity
- Suggested integration point in Phase 7: `if MatchesExceptions(b, rule.Unless) { skip }` — returning true means "exception fired, skip this bookmark"

### `hasNote` detection
- `Bookmark.Note` is a plain `string`; `hasNote: true` matches when `strings.TrimSpace(b.Note) != ""`
- Whitespace-only strings are treated as no note (defensive, avoids surprising behavior from API returning `"   "`)

### Carrying forward from prior phases
- Pointer fields for optional exception fields (`nil` = absent = skip check)
- Tests required alongside all implementation (project-level requirement)
- Same validation pattern: empty `HasTag` string in exceptions is a user error — reject at config load with a clear message

### Claude's Discretion
- Whether to add validation rejecting `hasNote: false` (semantically odd as an exception clause)
- Whether to validate `favourited: false` in exceptions (skip if NOT starred — unusual but not impossible)
- Exact function name and return polarity
- Test table structure

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §EXCP-01 — `unless: favourited`
- `.planning/REQUIREMENTS.md` §EXCP-02 — `unless: hasTag`
- `.planning/REQUIREMENTS.md` §EXCP-03 — `unless: hasNote`
- `.planning/REQUIREMENTS.md` §EXCP-04 — `unless: archived` / `unless: notArchived`

### Existing codebase (must read before planning)
- `internal/config/config.go` — `Exceptions` struct already defined: `Favourited *bool`, `HasTag *string`, `HasNote *bool`, `Archived *bool`
- `internal/engine/bookmark.go` — `Bookmark.Note string` (empty string = no note), `Archived bool`, `Favourited bool`, `Tags []string`
- `internal/engine/matcher.go` — `MatchesConditions` implementation; exception matcher follows the same pointer-nil pattern
- `internal/config/validate.go` — collected-error validation pattern; add `HasTag` empty-string check for exceptions block
- `internal/engine/matcher_test.go` — existing table-driven test structure to extend

### Prior phase context
- `.planning/phases/04-status-and-tag-conditions/04-CONTEXT.md` — tag matching semantics, pointer-nil pattern, AND short-circuit

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/engine/matcher.go`: `MatchesConditions` — exception matcher follows identical pointer-nil pattern but OR semantics (return true on first match)
- `internal/config/validate.go`: `ValidationError` type and collected-error pattern — reuse for `HasTag` empty-string validation in exceptions block
- `internal/engine/matcher_test.go`: table-driven test structure — extend with exception test cases

### Established Patterns
- Pointer fields: `nil` = field absent = skip check
- `MatchesConditions` short-circuits on first FALSE (AND); exception matcher short-circuits on first TRUE (OR)
- `Bookmark.Note` is a plain string; empty string means no note set

### Integration Points
- `internal/engine/matcher.go` — add exception evaluation function here (alongside `MatchesConditions`)
- `internal/config/validate.go` — add empty-string guard for `exceptions.hasTag` (parallel to `conditions.hasTag`)
- `internal/engine/matcher_test.go` — extend with exception cases
- Phase 7 (Run Orchestrator) will call exception matcher after `MatchesConditions` returns true

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches. User deferred all areas to Claude's discretion.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 05-exception-evaluation*
*Context gathered: 2026-03-18*
