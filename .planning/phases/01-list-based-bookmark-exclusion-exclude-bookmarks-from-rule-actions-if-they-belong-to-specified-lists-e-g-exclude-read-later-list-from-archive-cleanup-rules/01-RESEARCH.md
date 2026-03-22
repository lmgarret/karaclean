# Phase 1: List-Based Bookmark Exclusion - Research

**Researched:** 2026-03-22
**Domain:** Go YAML config extension, Karakeep list API integration, rule engine filtering
**Confidence:** HIGH

## Summary

This phase adds list-based filtering to the existing rule engine. Bookmarks can be targeted (`conditions.inList`) or protected (`unless.inList`) based on Karakeep list membership. The implementation touches four layers: config types, config validation, API client wrappers, and the rule engine (matcher + run orchestrator).

All building blocks already exist. The generated oapi-codegen client has `ListListsWithResponse()` and `GetListBookmarksWithResponse()` with cursor-based pagination identical to the existing `ListBookmarks` pattern. The `StringOrSlice` custom type was verified to work with `KnownFields(true)` -- no conflict with strict YAML parsing. The preload-by-list strategy (D-04) is straightforward: build a `map[string]map[string]bool` keyed by list name, values are sets of bookmark IDs.

**Primary recommendation:** Implement in three plans: (1) config types + StringOrSlice + structural validation, (2) API client wrappers (ListLists, GetListBookmarks) + ValidateListNames in main.go, (3) matcher integration + Run() preloading + end-to-end tests.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Lists are referenced by name only (not by ID). Example: `inList: "Read Later"`
- **D-02:** List name matching is case-sensitive, consistent with existing `hasTag` behavior
- **D-03:** If a configured list name doesn't exist in Karakeep, fail at startup with a validation error (not at runtime)
- **D-04:** Preload-by-list strategy: at run start, call `ListLists` to get all lists, then for each list name referenced in config, call `GetListBookmarks` to get bookmark IDs. Build a `map[string]bool` (set) of protected/targeted bookmark IDs. O(L) API calls where L = number of distinct referenced lists
- **D-05:** List data is fetched only when at least one rule uses `inList` (conditions or unless). Zero overhead for users who don't use this feature
- **D-06:** List preloading happens after `ListBookmarks` in the run flow, before rule evaluation
- **D-07:** `inList` field exists in both `conditions` (target bookmarks in a list) and `unless` (protect bookmarks in a list)
- **D-08:** `inList` supports single string or list of strings (StringOrSlice pattern with custom YAML unmarshaler). `inList: "Read Later"` and `inList: ["A", "B"]` both work
- **D-09:** `conditions.inList` uses OR semantics: bookmark matches if it's in ANY of the listed lists
- **D-10:** `unless.inList` uses OR semantics: bookmark is protected if it's in ANY of the listed lists (consistent with existing unless OR behavior)
- **D-11:** `config.Load()` stays pure (no network calls). Structural validation only (non-empty list names, valid YAML)
- **D-12:** A new validation step in `main.go` validates list names against the Karakeep API after client creation. Pattern: `config.Load` -> `NewKarakeepClient` -> `CheckAuth` -> `ValidateListNames` (new)
- **D-13:** `ValidateListNames` calls `ListLists` API, checks every `inList` name from config exists. Returns error listing all missing list names (not just first)

### Claude's Discretion
- StringOrSlice custom type implementation details
- Whether to add `Lists []string` to `engine.Bookmark` struct or keep list membership as a separate lookup set
- Preloaded list data structure (passed into matcher or looked up in run loop)
- How `GetListBookmarks` pagination is handled (follow existing `ListBookmarks` pagination pattern)

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go.yaml.in/yaml/v3 | 3.0.4 | YAML parsing with KnownFields | Already in use, maintained fork |

### Supporting
No new dependencies required. All functionality uses stdlib + existing generated client.

## Architecture Patterns

### StringOrSlice Custom Type (Discretion: Implementation)

**Recommendation:** Define `StringOrSlice` as `type StringOrSlice []string` in `internal/config/config.go` with a custom `UnmarshalYAML(value *yaml.Node) error` method.

**Verified behavior (HIGH confidence -- tested locally):**
- `KnownFields(true)` still rejects unknown fields on the parent struct even when a field has custom UnmarshalYAML
- String scalar (`inList: "Read Later"`) correctly unmarshals to `[]string{"Read Later"}`
- Sequence (`inList: ["A", "B"]`) correctly unmarshals to `[]string{"A", "B"}`
- Omitted field results in `nil` slice (zero value), usable as "not configured" signal

```go
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalYAML(value *yaml.Node) error {
    switch value.Kind {
    case yaml.ScalarNode:
        *s = []string{value.Value}
        return nil
    case yaml.SequenceNode:
        var items []string
        if err := value.Decode(&items); err != nil {
            return err
        }
        *s = items
        return nil
    default:
        return fmt.Errorf("expected string or list of strings")
    }
}
```

**Note on previous v1 decision:** Phase 01-01 decided "No custom UnmarshalYAML methods to preserve KnownFields strict parsing." This was conservative. Testing confirms custom UnmarshalYAML on a field type does NOT break KnownFields on the parent struct -- KnownFields operates at the struct level, not the field level. The previous concern was unfounded.

### List Membership Data Structure (Discretion: Data Structure)

**Recommendation:** Keep list membership as a separate lookup set, NOT on `engine.Bookmark`. Rationale:
1. Adding `Lists []string` to Bookmark would require modifying `toEngineBookmark()` with data that comes from a different API call (lists vs bookmarks)
2. A separate `map[string]map[string]bool` (listName -> bookmarkID set) is cleaner -- the matcher receives it as a parameter
3. This avoids coupling the Bookmark struct to list data that may not exist

**Data flow:**
```
Run() {
    bookmarks := api.ListBookmarks()

    // Only if rules reference inList
    listSets := preloadListSets(ctx, api, rules)
    // listSets: map[string]map[string]bool
    // key = list name, value = set of bookmark IDs in that list

    for _, b := range bookmarks {
        MatchesConditions(b, rule.Conditions, runTime, listSets)
        MatchesExceptions(b, rule.Unless, listSets)
    }
}
```

### Matcher Signature Change

**MatchesConditions** and **MatchesExceptions** signatures need a `listSets` parameter. Current:
```go
func MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time) bool
func MatchesExceptions(b Bookmark, ex *config.Exceptions) bool
```

New:
```go
func MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time, listSets map[string]map[string]bool) bool
func MatchesExceptions(b Bookmark, ex *config.Exceptions, listSets map[string]map[string]bool) bool
```

This is a breaking signature change. All callers (run.go, matcher_test.go) must be updated in the same plan.

### API Interface Extension

Add two methods to `engine.KarakeepAPI`:
```go
ListLists(ctx context.Context) ([]ListInfo, error)
GetListBookmarks(ctx context.Context, listID string) ([]string, error)
```

Where `ListInfo` is a new domain type:
```go
type ListInfo struct {
    ID   string
    Name string
}
```

`GetListBookmarks` returns only bookmark IDs (not full bookmarks) since we only need the ID for set membership checks.

### ValidateListNames in main.go

New step between CheckAuth and notifier creation:
```go
// Step 4.5: Validate list names if any rule uses inList
if listNames := cfg.CollectListNames(); len(listNames) > 0 {
    if err := validateListNames(ctx, client, listNames); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```

`CollectListNames()` is a config method that extracts all unique list names from all rules' conditions.InList and unless.InList. Returns empty slice if none configured (D-05 zero-overhead).

### Recommended Project Structure Changes
```
internal/config/
  config.go          # Add StringOrSlice type, InList fields to Conditions/Exceptions
  validate.go        # Add InList structural validation (non-empty strings)

internal/engine/
  api.go             # Add ListLists(), GetListBookmarks() to KarakeepAPI interface
  bookmark.go        # Add ListInfo type (NOT modify Bookmark)
  matcher.go         # Add listSets param, inList checks
  run.go             # Add preloadListSets(), update Run() signature/flow

internal/karakeep/
  client.go          # Add ListLists(), GetListBookmarks() wrapper methods

cmd/karaclean/
  main.go            # Add ValidateListNames step
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| YAML string-or-list parsing | Manual type assertion | Custom `UnmarshalYAML` on `yaml.Node` | Type-safe, handles edge cases, works with KnownFields |
| Cursor-based pagination | Custom loop | Clone existing `ListBookmarks` pattern in client.go | Proven pattern, same API shape (PaginatedBookmarks with NextCursor) |
| Set membership checks | Linear search | `map[string]bool` | O(1) lookup per bookmark per list |

## Common Pitfalls

### Pitfall 1: Empty StringOrSlice vs nil
**What goes wrong:** Treating `len(s) == 0` as "not configured" vs `s == nil`
**Why it happens:** Custom UnmarshalYAML for an empty sequence `inList: []` produces `[]string{}` (non-nil, len 0)
**How to avoid:** Use `s == nil` for "field not present in YAML." The custom unmarshaler never produces a non-nil empty slice because both branches (scalar, sequence) produce at least one element. An explicit `inList: []` would produce empty non-nil, but that is a validation error anyway (caught by structural validation).
**Warning signs:** Tests passing with nil but failing with empty slice

### Pitfall 2: Conditions empty check needs InList
**What goes wrong:** `validateConditions` currently checks all existing fields for nil to detect "empty conditions." Adding InList without updating this check means `conditions: { inList: "X" }` is rejected as "at least one condition required."
**Why it happens:** The nil-check list in `validateConditions` is hardcoded.
**How to avoid:** Add `cond.InList` to the nil-check list in `validateConditions`.

### Pitfall 3: Run() signature change ripple
**What goes wrong:** `Run()` currently takes `(ctx, api, rules, dryRun, notifications, notifier)`. Adding list preloading changes the flow but NOT the signature -- preloading happens inside Run() using the existing `api` parameter.
**Why it happens:** Temptation to pass preloaded data into Run().
**How to avoid:** Let Run() handle preloading internally. It already has access to `api` and `rules`.

### Pitfall 4: mockAPI update for tests
**What goes wrong:** Adding ListLists and GetListBookmarks to KarakeepAPI interface breaks the compile-time check `var _ engine.KarakeepAPI = (*mockAPI)(nil)` in api_test.go.
**Why it happens:** Interface extension requires all implementations to be updated.
**How to avoid:** Add the new methods to mockAPI in the same plan that extends the interface. Add corresponding test fields (`listListsRet`, `getListBookmarksRet`, etc.).

### Pitfall 5: GetListBookmarks pagination uses Cursor not NextCursor field name
**What goes wrong:** Assuming pagination fields differ from ListBookmarks.
**Why it happens:** Different API endpoints might use different pagination.
**How to avoid:** Both use `PaginatedBookmarks` struct with `NextCursor *string`. The exact same pagination loop pattern works. Verified: `GetListBookmarksResponse.JSON200` is `*PaginatedBookmarks` -- identical to ListBookmarks.

## Code Examples

### ListLists client wrapper
```go
// Source: follows existing ListBookmarks pattern in client.go
func (c *KarakeepClient) ListLists(ctx context.Context) ([]engine.ListInfo, error) {
    resp, err := c.inner.ListListsWithResponse(ctx)
    if err != nil {
        return nil, fmt.Errorf("listing lists: %w", err)
    }
    if resp.StatusCode() != http.StatusOK {
        return nil, fmt.Errorf("listing lists: unexpected status %d", resp.StatusCode())
    }
    lists := make([]engine.ListInfo, 0, len(resp.JSON200.Lists))
    for _, l := range resp.JSON200.Lists {
        lists = append(lists, engine.ListInfo{ID: l.Id, Name: l.Name})
    }
    return lists, nil
}
```

### GetListBookmarks client wrapper (paginated)
```go
// Source: follows existing ListBookmarks pagination pattern in client.go lines 54-85
func (c *KarakeepClient) GetListBookmarks(ctx context.Context, listID string) ([]string, error) {
    var ids []string
    var cursor *karakeep.Cursor

    for {
        limit := float32(100)
        resp, err := c.inner.GetListBookmarksWithResponse(ctx, listID, &karakeep.GetListBookmarksParams{
            Cursor: cursor,
            Limit:  &limit,
        })
        if err != nil {
            return nil, fmt.Errorf("getting list bookmarks: %w", err)
        }
        if resp.StatusCode() != http.StatusOK {
            return nil, fmt.Errorf("getting list bookmarks: unexpected status %d", resp.StatusCode())
        }
        for _, b := range resp.JSON200.Bookmarks {
            ids = append(ids, b.Id)
        }
        if resp.JSON200.NextCursor == nil {
            break
        }
        cursor = resp.JSON200.NextCursor
    }
    if ids == nil {
        ids = []string{}
    }
    return ids, nil
}
```

### inList condition check in matcher
```go
// inList check in MatchesConditions (OR semantics across list names)
if c.InList != nil {
    found := false
    for _, listName := range c.InList {
        if set, ok := listSets[listName]; ok && set[b.ID] {
            found = true
            break
        }
    }
    if !found {
        return false
    }
}
```

### inList exception check in matcher
```go
// inList check in MatchesExceptions (OR semantics -- any list match protects)
if ex.InList != nil {
    for _, listName := range ex.InList {
        if set, ok := listSets[listName]; ok && set[b.ID] {
            return true
        }
    }
}
```

### preloadListSets in run.go
```go
func preloadListSets(ctx context.Context, api KarakeepAPI, rules []config.Rule) (map[string]map[string]bool, error) {
    // Collect all unique list names from rules
    nameSet := make(map[string]bool)
    for _, r := range rules {
        if r.Conditions != nil {
            for _, name := range r.Conditions.InList {
                nameSet[name] = true
            }
        }
        if r.Unless != nil {
            for _, name := range r.Unless.InList {
                nameSet[name] = true
            }
        }
    }
    if len(nameSet) == 0 {
        return nil, nil // D-05: zero overhead
    }

    // Resolve list names to IDs
    lists, err := api.ListLists(ctx)
    if err != nil {
        return nil, fmt.Errorf("preloading lists: %w", err)
    }
    nameToID := make(map[string]string)
    for _, l := range lists {
        if nameSet[l.Name] {
            nameToID[l.Name] = l.ID
        }
    }

    // Fetch bookmark IDs for each list
    result := make(map[string]map[string]bool)
    for name, id := range nameToID {
        ids, err := api.GetListBookmarks(ctx, id)
        if err != nil {
            return nil, fmt.Errorf("preloading list %q: %w", name, err)
        }
        set := make(map[string]bool, len(ids))
        for _, bid := range ids {
            set[bid] = true
        }
        result[name] = set
    }
    return result, nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No list filtering | inList in conditions + unless | This phase | Users can target/protect bookmarks by list membership |

No deprecated patterns apply -- this is new functionality.

## Open Questions

1. **Should preloadListSets error be fatal or logged?**
   - What we know: ListBookmarks error is fatal in Run() (returns error). List preloading is similar -- it's needed for correct rule evaluation.
   - What's unclear: User expectation when a list API call fails mid-run.
   - Recommendation: Make it fatal (return error from Run()). Incorrect list data would cause wrong bookmarks to be deleted. Fail-safe is better. This is consistent with ListBookmarks error being fatal.

2. **Should ValidateListNames also run at the start of each cron run, or only at startup?**
   - What we know: D-03 says "fail at startup." D-12 puts it in main.go after CheckAuth.
   - What's unclear: If a list is deleted between cron runs, preloadListSets will just get an empty set for that list name.
   - Recommendation: Only at startup per D-03. If a list is deleted later, preloadListSets will not find the name-to-ID mapping, and the list will effectively be treated as empty (no bookmarks match). This is safe behavior -- it means fewer bookmarks are protected/targeted, which the user caused by deleting the list.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none -- uses `go test` |
| Quick run command | `go test ./...` |
| Full suite command | `go test -race -count=1 ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| D-01 | Lists referenced by name, not ID | unit | `go test ./internal/config/ -run TestLoad_InList -x` | No -- Wave 0 |
| D-08 | StringOrSlice accepts string or list | unit | `go test ./internal/config/ -run TestStringOrSlice -x` | No -- Wave 0 |
| D-09 | conditions.inList OR semantics | unit | `go test ./internal/engine/ -run TestMatchesConditions_InList -x` | No -- Wave 0 |
| D-10 | unless.inList OR semantics | unit | `go test ./internal/engine/ -run TestMatchesExceptions_InList -x` | No -- Wave 0 |
| D-11 | Structural validation (non-empty names) | unit | `go test ./internal/config/ -run TestValidate_InList -x` | No -- Wave 0 |
| D-13 | ValidateListNames catches missing lists | unit | `go test ./cmd/karaclean/ -run TestValidateListNames -x` | No -- Wave 0 |
| D-04 | Preload-by-list builds correct sets | unit | `go test ./internal/engine/ -run TestPreloadListSets -x` | No -- Wave 0 |
| D-05 | No preloading when no inList rules | unit | `go test ./internal/engine/ -run TestPreloadListSets_NoLists -x` | No -- Wave 0 |
| D-03 | Missing list name fails at startup | unit | `go test ./cmd/karaclean/ -run TestValidateListNames_Missing -x` | No -- Wave 0 |
| API | ListLists wrapper | unit | `go test ./internal/karakeep/ -run TestListLists -x` | No -- Wave 0 |
| API | GetListBookmarks wrapper with pagination | unit | `go test ./internal/karakeep/ -run TestGetListBookmarks -x` | No -- Wave 0 |
| E2E | Full run with inList condition + exception | integration | `go test ./internal/engine/ -run TestRun_InList -x` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./...`
- **Per wave merge:** `go test -race -count=1 ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/config/config_test.go` -- add StringOrSlice and inList loading tests
- [ ] `internal/config/validate_test.go` -- add inList validation tests
- [ ] `internal/engine/matcher_test.go` -- add inList condition/exception tests (update mockAPI)
- [ ] `internal/engine/api_test.go` -- update mockAPI with new interface methods
- [ ] `internal/engine/run_test.go` -- add preloadListSets and Run with inList tests
- [ ] `internal/config/testdata/valid_inlist_string.yaml` -- test fixture
- [ ] `internal/config/testdata/valid_inlist_list.yaml` -- test fixture

## Sources

### Primary (HIGH confidence)
- Local codebase analysis: `internal/config/config.go`, `internal/engine/matcher.go`, `internal/engine/run.go`, `internal/engine/api.go`, `internal/karakeep/client.go`, `internal/karakeep/client.gen.go`
- Local verification: StringOrSlice + KnownFields(true) compatibility tested and confirmed
- Generated API types: `ListListsResponse.JSON200.Lists []List`, `GetListBookmarksResponse.JSON200 *PaginatedBookmarks` (same pagination as ListBookmarks)

### Secondary (MEDIUM confidence)
- go.yaml.in/yaml/v3 v3.0.4: custom UnmarshalYAML with yaml.Node API (verified via local test)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- no new dependencies, all existing libraries
- Architecture: HIGH -- patterns directly mirror existing code (ListBookmarks -> GetListBookmarks, hasTag -> inList)
- Pitfalls: HIGH -- identified from direct code analysis, validated KnownFields concern

**Research date:** 2026-03-22
**Valid until:** 2026-04-22 (stable -- no external dependency changes expected)
