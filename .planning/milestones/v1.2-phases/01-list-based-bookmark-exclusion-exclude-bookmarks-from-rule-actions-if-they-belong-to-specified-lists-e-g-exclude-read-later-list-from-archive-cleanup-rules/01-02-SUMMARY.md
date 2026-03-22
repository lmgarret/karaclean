---
phase: 01-list-based-bookmark-exclusion
plan: 02
subsystem: api
tags: [karakeep, pagination, cursor, lists, validation]

requires:
  - phase: 01-list-based-bookmark-exclusion-01
    provides: "KarakeepAPI interface with ListLists/GetListBookmarks, ListInfo type, CollectListNames"
provides:
  - "ListLists API wrapper returning []ListInfo"
  - "GetListBookmarks API wrapper with cursor pagination returning []string IDs"
  - "validateListNames startup step in main.go"
affects: [01-list-based-bookmark-exclusion-03]

tech-stack:
  added: []
  patterns: [cursor-pagination-for-list-bookmarks]

key-files:
  created: []
  modified:
    - internal/karakeep/client.go
    - internal/karakeep/client_test.go
    - cmd/karaclean/main.go
    - cmd/karaclean/main_test.go

key-decisions:
  - "ListLists does not paginate (API returns all lists in one call)"
  - "GetListBookmarks clones ListBookmarks pagination pattern with cursor + limit=100"
  - "validateListNames collects ALL missing names before returning error (D-13)"

patterns-established:
  - "List API wrappers follow same error handling pattern as bookmark wrappers"

requirements-completed: [D-03, D-04, D-05, D-12, D-13]

duration: 3min
completed: 2026-03-22
---

# Phase 01 Plan 02: API Client Wrappers and Startup Validation Summary

**ListLists and GetListBookmarks API wrappers with cursor pagination, plus validateListNames startup check that fails fast on misconfigured list names**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-22T15:32:38Z
- **Completed:** 2026-03-22T15:35:29Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- ListLists wrapper maps Karakeep API List objects to engine.ListInfo (ID + Name)
- GetListBookmarks wrapper paginates with cursor and returns all bookmark IDs as []string
- validateListNames in main.go checks all configured list names exist, reports ALL missing names
- Zero overhead when no rules use inList (CollectListNames returns empty, validation skipped)

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement ListLists and GetListBookmarks client wrappers** - `1e9e38e` (test: RED), `51d61fc` (feat: GREEN)
2. **Task 2: Add ValidateListNames startup step in main.go with tests** - `72dbb2c` (test: RED), `d6cee6f` (feat: GREEN)

_Note: TDD tasks have two commits each (test then feat)_

## Files Created/Modified
- `internal/karakeep/client.go` - Replaced stub ListLists/GetListBookmarks with real implementations
- `internal/karakeep/client_test.go` - Added 7 httptest-based tests for ListLists and GetListBookmarks
- `cmd/karaclean/main.go` - Added validateListNames function and startup call between CheckAuth and notifier
- `cmd/karaclean/main_test.go` - Added 3 tests for validateListNames (happy, missing, API error)

## Decisions Made
- ListLists does not paginate -- the Karakeep API returns all lists in a single response (no NextCursor)
- GetListBookmarks clones the existing ListBookmarks pagination pattern (cursor + limit=100)
- validateListNames collects ALL missing list names before returning error, not just the first (D-13)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- API wrappers ready for plan 03 to build list set resolution and inList matching in the engine
- validateListNames ensures fail-fast on misconfigured list names at startup

## Self-Check: PASSED

All files exist, all commits found, all key content verified.

---
*Phase: 01-list-based-bookmark-exclusion*
*Completed: 2026-03-22*
