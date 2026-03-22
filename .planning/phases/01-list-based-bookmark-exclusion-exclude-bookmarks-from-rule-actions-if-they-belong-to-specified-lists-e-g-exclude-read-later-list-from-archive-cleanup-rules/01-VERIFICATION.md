---
phase: 01-list-based-bookmark-exclusion
verified: 2026-03-22T15:41:18Z
status: passed
score: 14/14 must-haves verified
re_verification: false
---

# Phase 01: List-Based Bookmark Exclusion Verification Report

**Phase Goal:** Exclude bookmarks from rule actions if they belong to specified lists (e.g., exclude "Read Later" list from archive/cleanup rules)
**Verified:** 2026-03-22T15:41:18Z
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| #  | Truth | Status | Evidence |
|----|-------|--------|----------|
| 1  | `inList` field accepted in both `conditions` and `unless` YAML blocks | VERIFIED | `Conditions.InList StringOrSlice` and `Exceptions.InList StringOrSlice` in `internal/config/config.go:65-83` |
| 2  | `inList` accepts a single string or a list of strings | VERIFIED | `StringOrSlice.UnmarshalYAML` handles `yaml.ScalarNode` and `yaml.SequenceNode` in `config.go:26-41`; fixtures `valid_inlist_string.yaml` and `valid_inlist_list.yaml` exist |
| 3  | Empty list names are rejected during structural validation | VERIFIED | `validateConditions` iterates `cond.InList` checking for `""` at `validate.go:186-193`; `validateRule` checks `rule.Unless.InList` at `validate.go:123-130`; error message `"list name must not be empty"` |
| 4  | `KarakeepAPI` interface declares `ListLists` and `GetListBookmarks` methods | VERIFIED | Both methods present in `internal/engine/api.go:22-28` |
| 5  | `mockAPI` compiles with new interface methods | VERIFIED | `func (m *mockAPI) ListLists` and `func (m *mockAPI) GetListBookmarks` in `internal/engine/api_test.go:47-56`; compile-time check `var _ engine.KarakeepAPI = (*mockAPI)(nil)` at line 59 |
| 6  | `ListLists` API wrapper returns `[]ListInfo` with ID and Name | VERIFIED | `internal/karakeep/client.go:115-128` maps `l.Id` and `l.Name` to `engine.ListInfo`; real implementation, not stub |
| 7  | `GetListBookmarks` API wrapper paginates and returns all bookmark IDs | VERIFIED | `internal/karakeep/client.go:131-158` uses cursor pagination loop identical to `ListBookmarks` pattern |
| 8  | `ValidateListNames` at startup rejects config if a referenced list name does not exist in Karakeep | VERIFIED | `validateListNames` in `cmd/karaclean/main.go:152-171` called from main at line 74-79 between `CheckAuth` and notifier creation |
| 9  | `ValidateListNames` is skipped when no rules use `inList` (D-05 zero overhead) | VERIFIED | `main.go:74`: guard `if listNames := cfg.CollectListNames(); len(listNames) > 0` — validation only runs when non-empty |
| 10 | `conditions.inList` matches bookmark if it belongs to ANY listed list (OR semantics) | VERIFIED | `MatchesConditions` in `matcher.go:43-54`: iterates `c.InList`, sets `found=true` on first set membership hit; test `TestMatchesConditions_InList/inList_OR_semantics_bookmark_in_second_list` |
| 11 | `unless.inList` protects bookmark if it belongs to ANY listed list (OR semantics) | VERIFIED | `MatchesExceptions` in `matcher.go:110-116`: iterates `ex.InList`, returns `true` on first hit; test `TestMatchesExceptions_InList` |
| 12 | List name matching is case-sensitive (D-02) | VERIFIED | Plain map key lookup in both matchers (no case folding); tests `case_sensitive_no_match` at `matcher_test.go:465` and `matcher_test.go:525` verify `"read later"` does not match key `"Read Later"` |
| 13 | List data is preloaded before rule evaluation and only when rules use `inList` (D-05) | VERIFIED | `PreloadListSets` in `run.go:35-77` returns `nil` immediately when `nameSet` is empty; called in `Run()` at line 98 before rule evaluation loop at line 114 |
| 14 | `Run()` correctly passes `listSets` to matcher functions | VERIFIED | `run.go:117` passes `listSets` to `MatchesConditions`; `run.go:120` passes `listSets` to `MatchesExceptions` |

**Score:** 14/14 truths verified

---

### Required Artifacts

| Artifact | Provides | Status | Details |
|----------|----------|--------|---------|
| `internal/config/config.go` | `StringOrSlice` type, `InList` fields on `Conditions` and `Exceptions` | VERIFIED | `type StringOrSlice []string` at line 24; `InList StringOrSlice` on both structs at lines 72, 82 |
| `internal/config/validate.go` | `InList` structural validation and `CollectListNames` method | VERIFIED | `func (c *Config) CollectListNames` at line 259; empty-name loops in `validateConditions` (line 186) and `validateRule` (line 123) |
| `internal/engine/api.go` | Extended `KarakeepAPI` interface with `ListLists`, `GetListBookmarks` | VERIFIED | Both methods declared at lines 22-28 |
| `internal/engine/bookmark.go` | `ListInfo` domain type with `ID` and `Name` fields | VERIFIED | `type ListInfo struct` with `ID string` and `Name string` at lines 19-22 |
| `internal/engine/api_test.go` | Updated `mockAPI` with new interface methods and test cases | VERIFIED | `ListLists` and `GetListBookmarks` methods; tests at lines 134-194 |
| `internal/karakeep/client.go` | `ListLists` and `GetListBookmarks` wrapper methods (real implementation) | VERIFIED | Real implementations with pagination at lines 115-158; no stub/TODO |
| `cmd/karaclean/main.go` | `validateListNames` step between `CheckAuth` and notifier creation | VERIFIED | Function at line 152; called at line 74 after `CheckAuth` (line 68), before notifier (line 82) |
| `cmd/karaclean/main_test.go` | Tests for `validateListNames` happy and error paths | VERIFIED | `TestValidateListNames`, `TestValidateListNames_Missing`, `TestValidateListNames_APIError` at lines 99-140 |
| `internal/engine/matcher.go` | `inList` checks in `MatchesConditions` and `MatchesExceptions` with `listSets` parameter | VERIFIED | Both function signatures include `listSets map[string]map[string]bool`; inList checks at lines 43-54 and 110-116 |
| `internal/engine/matcher_test.go` | `TestMatchesConditions_InList` and `TestMatchesExceptions_InList` with case-sensitivity tests | VERIFIED | Both test functions present; `case_sensitive_no_match` subtests at lines 465 and 525 |
| `internal/engine/run.go` | `PreloadListSets` function and updated `Run()` flow | VERIFIED | `func PreloadListSets` at line 35 (exported); wired into `Run()` at line 98 |
| `internal/engine/run_test.go` | Tests for `PreloadListSets` and `Run` with `inList` | VERIFIED | `TestPreloadListSets_NoInList`, `TestPreloadListSets_ResolvesAndFetches`, `TestPreloadListSets_ListListsError`, `TestPreloadListSets_GetListBookmarksError`, `TestRun_InListCondition`, `TestRun_InListException` |
| `internal/config/testdata/valid_inlist_string.yaml` | Fixture: `inList` as string | VERIFIED | File exists |
| `internal/config/testdata/valid_inlist_list.yaml` | Fixture: `inList` as list | VERIFIED | File exists |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/config/config.go` | `internal/config/validate.go` | `cond.InList` used in nil-check | VERIFIED | `cond.InList == nil` in `validateConditions` at `validate.go:147` |
| `internal/engine/api_test.go` | `internal/engine/api.go` | compile-time interface check | VERIFIED | `var _ engine.KarakeepAPI = (*mockAPI)(nil)` at `api_test.go:59` |
| `internal/karakeep/client.go` | `internal/engine/api.go` | implements `KarakeepAPI` interface | VERIFIED | `var _ engine.KarakeepAPI = (*KarakeepClient)(nil)` at `client.go:19`; `ListLists` and `GetListBookmarks` fully implemented |
| `cmd/karaclean/main.go` | `internal/config/validate.go` | calls `CollectListNames` | VERIFIED | `cfg.CollectListNames()` at `main.go:74` |
| `internal/engine/run.go` | `internal/engine/matcher.go` | passes `listSets` to `MatchesConditions` and `MatchesExceptions` | VERIFIED | `MatchesConditions(b, rule.Conditions, runTime, listSets)` at `run.go:117`; `MatchesExceptions(b, rule.Unless, listSets)` at `run.go:120` |
| `internal/engine/run.go` | `internal/engine/api.go` | calls `ListLists` and `GetListBookmarks` via `KarakeepAPI` | VERIFIED | `api.ListLists(ctx)` at `run.go:53` inside `PreloadListSets`; `api.GetListBookmarks(ctx, id)` at `run.go:66` |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| D-01 | 01-01 | Lists referenced by name only, not by ID | SATISFIED | Config stores `InList StringOrSlice` (names); `PreloadListSets` resolves names to IDs via `nameToID` map lookup |
| D-02 | 01-01, 01-03 | List name matching is case-sensitive | SATISFIED | Plain `map[string]bool` key lookup in `MatchesConditions` and `MatchesExceptions`; `case_sensitive_no_match` tests in `matcher_test.go:465,525` |
| D-03 | 01-02 | Fail at startup if configured list name doesn't exist in Karakeep | SATISFIED | `validateListNames` called in `main.go:74-79` after `CheckAuth`, exits on missing names |
| D-04 | 01-02, 01-03 | Preload-by-list strategy: `ListLists` then per-list `GetListBookmarks` | SATISFIED | `PreloadListSets` in `run.go:35-77` calls `api.ListLists` then `api.GetListBookmarks` per referenced list |
| D-05 | 01-02, 01-03 | Zero overhead when no rules use `inList` | SATISFIED | `PreloadListSets` returns `nil` when `nameSet` is empty (`run.go:49-51`); main skips `validateListNames` when `CollectListNames` returns empty |
| D-06 | 01-03 | List preloading happens after `ListBookmarks`, before rule evaluation | SATISFIED | `Run()` order: `ListBookmarks` at line 91, `PreloadListSets` at line 98, rule loop at line 116 |
| D-07 | 01-01 | `inList` field exists in both `conditions` and `unless` | SATISFIED | `Conditions.InList` at `config.go:72`; `Exceptions.InList` at `config.go:82` |
| D-08 | 01-01 | `inList` supports single string or list of strings (`StringOrSlice`) | SATISFIED | `StringOrSlice.UnmarshalYAML` handles both scalar and sequence YAML nodes |
| D-09 | 01-01, 01-03 | `conditions.inList` uses OR semantics | SATISFIED | `MatchesConditions` sets `found=true` on ANY matching list (`matcher.go:43-54`); test `inList_OR_semantics_bookmark_in_second_list` |
| D-10 | 01-01, 01-03 | `unless.inList` uses OR semantics | SATISFIED | `MatchesExceptions` returns `true` on first matching list (`matcher.go:110-116`); test `TestMatchesExceptions_InList/inList_OR_single_match` |
| D-11 | 01-01 | `config.Load()` stays pure (no network calls); structural validation only | SATISFIED | `Load` calls `cfg.Validate()` which calls `validateConditions` — only string non-empty checks, no network |
| D-12 | 01-02 | Validation step in `main.go`: `Load` → `NewKarakeepClient` → `CheckAuth` → `ValidateListNames` | SATISFIED | `main.go` follows this exact order at lines 30, 61, 68, 74 |
| D-13 | 01-02 | `ValidateListNames` reports ALL missing names, not just first | SATISFIED | `validateListNames` accumulates `var missing []string` before returning; `TestValidateListNames_Missing` verifies both "No Such List" and "Also Missing" appear in error |

All 13 requirement IDs (D-01 through D-13) are accounted for. No orphaned requirements.

---

### Anti-Patterns Found

None. Scan of all phase-modified files revealed:
- No `TODO`, `FIXME`, `PLACEHOLDER`, or stub comments
- No `return nil` or static-data stubs in `ListLists`/`GetListBookmarks` — both are real implementations
- No form handlers that only call `preventDefault`
- `PreloadListSets` is exported (not `preloadListSets` as the plan originally specified) — this is a deliberate improvement for testability documented in the summary as a decision, not a gap

---

### Human Verification Required

None. All observable behaviors are verifiable programmatically:
- Config parsing, struct fields, validation logic, and `CollectListNames` are all code-inspectable
- Matcher logic is deterministic and fully covered by unit tests
- Startup flow ordering is statically readable in `main.go`
- Test suite passes including race detector (`go test -race ./...`)

---

### Test Suite Results

```
ok  github.com/lm/karaclean/cmd/karaclean       (TestValidateListNames, TestValidateListNames_Missing, TestValidateListNames_APIError)
ok  github.com/lm/karaclean/internal/config     (TestLoad_InList*, TestValidate_InList*, TestCollectListNames*)
ok  github.com/lm/karaclean/internal/engine     (TestMatchesConditions_InList, TestMatchesExceptions_InList, TestPreloadListSets*, TestRun_InListCondition, TestRun_InListException)
ok  github.com/lm/karaclean/internal/karakeep   (httptest-based tests for ListLists and GetListBookmarks)
```

`go test -race -count=1 ./...` — all packages pass, race detector clean.

---

## Gaps Summary

No gaps. All 14 observable truths verified, all 14 artifacts exist and are substantive (not stubs), all 6 key links are wired, all 13 requirements are satisfied, and the full test suite passes including the race detector.

---

_Verified: 2026-03-22T15:41:18Z_
_Verifier: Claude (gsd-verifier)_
