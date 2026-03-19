# Phase 6: Actions and Dry-Run - Context

**Gathered:** 2026-03-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Implement the two mutation actions (`archive` and `delete`) that apply a matched rule's outcome to a bookmark via the Karakeep API. Add dry-run mode that suppresses all mutations and logs intended actions instead. No run orchestration or pagination — that is Phase 7. This phase delivers the action execution layer that Phase 7's orchestrator will call.

</domain>

<decisions>
## Implementation Decisions

### Dry-run activation
- All three mechanisms activate dry-run — any one is sufficient:
  1. `--dry-run` CLI flag
  2. `dryRun: true` in the YAML config file (new top-level field alongside `timezone`, `schedule`, `rules`)
  3. `KARACLEAN_DRY_RUN=true` environment variable
- Precedence: flag > env var > config field (flag wins, then env var, then config)
- Consistent with `KARAKEEP_URL` / `KARAKEEP_API_KEY` pattern already established for env-based config

### Action error handling
- **Log and continue** — if an archive or delete API call fails for a specific bookmark, log the error with bookmark ID and rule name, then continue processing remaining bookmarks
- One bad API response (transient error, 404, etc.) must not abort the entire cleanup run
- Log format: `ERROR archive failed: bookmark <id> (rule: <name>): <error>`
- Error count is available for Phase 7's run summary (`errors: E`)

### Dry-run output format
- Claude's Discretion — user did not specify; log one line per intended action with bookmark ID, action type, and rule name
- Must clearly distinguish dry-run lines (e.g., `DRY-RUN archive: bookmark <id> (rule: <name>)`) so users can grep or scan easily

### API interface extension
- Claude's Discretion — add `ArchiveBookmark(ctx, id)` and `DeleteBookmark(ctx, id)` to `KarakeepAPI` interface; implement in `KarakeepClient`
- Generated client already has `UpdateBookmark` (PATCH with `archived: true`) and `DeleteBookmark` (DELETE) — thin wrapper methods follow the same pattern as `CheckAuth` and `ListBookmarks`

### Claude's Discretion
- Whether action execution is a standalone function (`ExecuteAction`) or small per-action functions
- Where action logic lives (`internal/engine/actions.go` or alongside matcher)
- Exact log level and format for dry-run lines
- Whether `dryRun` field in config triggers a startup log warning ("dry-run mode enabled — no mutations will be executed")

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §ACTN-01 — archive action via Karakeep PATCH API
- `.planning/REQUIREMENTS.md` §ACTN-02 — delete action via Karakeep DELETE API
- `.planning/REQUIREMENTS.md` §ACTN-03 — dry-run mode: no mutations, all actions logged

### Existing codebase (must read before planning)
- `internal/engine/api.go` — `KarakeepAPI` interface; Phase 6 adds `ArchiveBookmark` and `DeleteBookmark` methods (comment in file already anticipates this)
- `internal/karakeep/client.go` — `KarakeepClient` wrapper; new methods implement the extended interface, following `CheckAuth`/`ListBookmarks` patterns
- `internal/config/config.go` — `Config` struct; add `DryRun bool` field with `yaml:"dryRun"` tag
- `cmd/karaclean/main.go` — startup wiring; add `--dry-run` flag and `KARACLEAN_DRY_RUN` env var resolution with flag > env > config precedence

### Karakeep API (for endpoint details)
- `karakeep-upstream/packages/open-api/karakeep-openapi-spec.json` — `PATCH /bookmarks/{id}` (UpdateBookmark with `archived: true`) and `DELETE /bookmarks/{id}` endpoints

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/karakeep/client.go`: `CheckAuth` and `ListBookmarks` — pattern for wrapping generated client methods with error handling and status code checks
- `internal/engine/matcher.go`: `MatchesConditions` and `MatchesExceptions` — action execution will be called after both return the right values (Phase 7 wires this up; Phase 6 just provides the action functions)
- `internal/config/config.go`: `Config` struct with pointer-optional fields — `DryRun bool` is a plain bool (zero value = false = live mode)

### Established Patterns
- Fail-fast for startup errors (missing env vars, auth failure)
- Log-and-continue for per-item operational errors (this phase introduces it for action failures)
- `fmt.Errorf("context: %w", err)` error wrapping
- Interface-driven design: `KarakeepAPI` injected into engine functions, enabling mock-based tests without real HTTP

### Integration Points
- `internal/engine/api.go` — extend `KarakeepAPI` interface with `ArchiveBookmark(ctx context.Context, id string) error` and `DeleteBookmark(ctx context.Context, id string) error`
- `internal/karakeep/client.go` — implement new interface methods; `var _ engine.KarakeepAPI = (*KarakeepClient)(nil)` compile-time check will enforce this
- `internal/config/config.go` — add `DryRun bool` field
- `cmd/karaclean/main.go` — flag + env var + config field dry-run resolution; pass dry-run bool into action executor
- Phase 7 orchestrator (not yet built) will call action functions after matcher evaluation

</code_context>

<specifics>
## Specific Ideas

- Dry-run lines must be clearly labelled (e.g., `DRY-RUN` prefix) so users can distinguish preview output from live execution output at a glance
- Log-and-continue error model means Phase 7's run summary will include an `errors: E` count — action layer should surface the error count upward

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 06-actions-and-dry-run*
*Context gathered: 2026-03-18*
