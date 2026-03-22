# Phase 1: List-Based Bookmark Exclusion - Context

**Gathered:** 2026-03-22
**Status:** Ready for planning

<domain>
## Phase Boundary

Add list-based filtering to rules: `conditions.inList` to target bookmarks in specific lists, and `unless.inList` to protect bookmarks in specific lists from a rule's action. Users reference lists by name in YAML config. List data is preloaded from the Karakeep API only when rules use list-based fields.

</domain>

<decisions>
## Implementation Decisions

### List Identification
- **D-01:** Lists are referenced by name only (not by ID). Example: `inList: "Read Later"`
- **D-02:** List name matching is case-sensitive, consistent with existing `hasTag` behavior
- **D-03:** If a configured list name doesn't exist in Karakeep, fail at startup with a validation error (not at runtime)

### API Data Strategy
- **D-04:** Preload-by-list strategy: at run start, call `ListLists` to get all lists, then for each list name referenced in config, call `GetListBookmarks` to get bookmark IDs. Build a `map[string]bool` (set) of protected/targeted bookmark IDs. O(L) API calls where L = number of distinct referenced lists
- **D-05:** List data is fetched only when at least one rule uses `inList` (conditions or unless). Zero overhead for users who don't use this feature
- **D-06:** List preloading happens after `ListBookmarks` in the run flow, before rule evaluation

### Config Field Design
- **D-07:** `inList` field exists in both `conditions` (target bookmarks in a list) and `unless` (protect bookmarks in a list)
- **D-08:** `inList` supports single string or list of strings (StringOrSlice pattern with custom YAML unmarshaler). `inList: "Read Later"` and `inList: ["A", "B"]` both work
- **D-09:** `conditions.inList` uses OR semantics: bookmark matches if it's in ANY of the listed lists
- **D-10:** `unless.inList` uses OR semantics: bookmark is protected if it's in ANY of the listed lists (consistent with existing unless OR behavior)

### Validation Design
- **D-11:** `config.Load()` stays pure (no network calls). Structural validation only (non-empty list names, valid YAML)
- **D-12:** A new validation step in `main.go` validates list names against the Karakeep API after client creation. Pattern: `config.Load` → `NewKarakeepClient` → `CheckAuth` → `ValidateListNames` (new)
- **D-13:** `ValidateListNames` calls `ListLists` API, checks every `inList` name from config exists. Returns error listing all missing list names (not just first)

### Claude's Discretion
- StringOrSlice custom type implementation details
- Whether to add `Lists []string` to `engine.Bookmark` struct or keep list membership as a separate lookup set
- Preloaded list data structure (passed into matcher or looked up in run loop)
- How `GetListBookmarks` pagination is handled (follow existing `ListBookmarks` pagination pattern)

</decisions>

<specifics>
## Specific Ideas

- Example config the user wants to support:
  ```yaml
  rules:
    - name: clean-old-archived
      conditions:
        archived: true
        olderThan: 30d
      unless:
        inList: "Read Later"
      action: delete
  ```
- Also supports targeting by list:
  ```yaml
  rules:
    - name: clean-rss-list
      conditions:
        inList: "RSS Feeds"
      unless:
        inList: "Read Later"
      action: delete
  ```

</specifics>

<canonical_refs>
## Canonical References

No external specs — requirements are fully captured in decisions above.

### Existing patterns to follow
- `internal/config/config.go` — Conditions/Exceptions struct patterns, pointer types for optional fields
- `internal/engine/matcher.go` — MatchesConditions/MatchesExceptions evaluation patterns
- `internal/karakeep/client.go` — ListBookmarks pagination pattern (reuse for GetListBookmarks)
- `internal/engine/run.go` — Run() orchestration flow where list preloading should be inserted
- `internal/engine/api.go` — KarakeepAPI interface where new list methods should be added

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `KarakeepClient` already has generated methods: `ListListsWithResponse()`, `GetListBookmarksWithResponse()`, `GetBookmarkListsWithResponse()` — just need wrapper methods
- `ListBookmarks` pagination pattern in `client.go` lines 54-85 — reusable for `GetListBookmarks` pagination
- `toEngineBookmark()` mapping pattern — reference for any new mapping

### Established Patterns
- Conditions use AND semantics (all non-nil must match) — `inList` fits as another AND clause
- Exceptions use OR semantics (any match protects) — `inList` fits as another OR clause
- Pointer types for optional config fields (`*string`, `*bool`) — `inList` needs custom type (StringOrSlice)
- Config validation returns `[]ValidationError` slice for all errors at once
- `KnownFields(true)` in YAML decoder rejects unknown fields — new fields must be in struct

### Integration Points
- `config.Conditions` struct — add `InList` field
- `config.Exceptions` struct — add `InList` field
- `config.Validate()` — add structural validation for `InList` (non-empty names)
- `engine.KarakeepAPI` interface — add `ListLists()` and `GetListBookmarks()` methods
- `karakeep.KarakeepClient` — implement new interface methods
- `engine.MatchesConditions()` — add inList check using preloaded set
- `engine.MatchesExceptions()` — add inList check using preloaded set
- `engine.Run()` — add list preloading step between ListBookmarks and rule evaluation
- `main.go` — add `ValidateListNames` call after `CheckAuth`

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-list-based-bookmark-exclusion*
*Context gathered: 2026-03-22*
