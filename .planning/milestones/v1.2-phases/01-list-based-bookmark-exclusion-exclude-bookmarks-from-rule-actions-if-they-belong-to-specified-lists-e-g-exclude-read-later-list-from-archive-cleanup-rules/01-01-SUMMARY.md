---
phase: 01-list-based-bookmark-exclusion
plan: 01
subsystem: config, engine
tags: [yaml, custom-unmarshal, api-interface, list-exclusion]

requires:
  - phase: v1.1-notification-system
    provides: Config/Validate structure, KarakeepAPI interface, mockAPI pattern

provides:
  - StringOrSlice type with custom UnmarshalYAML for scalar/sequence YAML
  - InList field on Conditions and Exceptions structs
  - Structural validation for empty list names
  - CollectListNames helper for extracting unique list names from config
  - Extended KarakeepAPI interface with ListLists and GetListBookmarks
  - ListInfo domain type with ID and Name fields
  - Updated mockAPI with new interface methods

affects: [01-02-api-client, 01-03-matcher-and-run]

tech-stack:
  added: []
  patterns:
    - "StringOrSlice custom UnmarshalYAML for flexible YAML input (string or list)"
    - "Stub methods with TODO markers for interface compliance before full implementation"

key-files:
  created:
    - internal/config/testdata/valid_inlist_string.yaml
    - internal/config/testdata/valid_inlist_list.yaml
  modified:
    - internal/config/config.go
    - internal/config/validate.go
    - internal/config/config_test.go
    - internal/config/validate_test.go
    - internal/engine/api.go
    - internal/engine/bookmark.go
    - internal/engine/api_test.go
    - internal/karakeep/client.go

key-decisions:
  - "StringOrSlice uses custom UnmarshalYAML with yaml.Node to work with KnownFields(true)"
  - "KarakeepClient gets stub methods returning 'not yet implemented' for interface compliance (plan 02 implements)"

patterns-established:
  - "StringOrSlice pattern: custom UnmarshalYAML on yaml.Node for flexible scalar/sequence input"

requirements-completed: [D-01, D-02, D-07, D-08, D-09, D-10, D-11]

duration: 4min
completed: 2026-03-22
---

# Phase 01 Plan 01: Config Types and API Interface Summary

**StringOrSlice custom YAML type for inList fields on Conditions/Exceptions, structural validation, CollectListNames helper, and KarakeepAPI extended with ListLists/GetListBookmarks**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-22T15:25:38Z
- **Completed:** 2026-03-22T15:29:57Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- StringOrSlice type accepting both `inList: "Read Later"` and `inList: ["A", "B"]` YAML forms
- InList structural validation rejects empty list names in conditions and unless blocks
- CollectListNames extracts deduplicated list names across all rules for pre-fetching
- KarakeepAPI interface extended with ListLists and GetListBookmarks for list membership queries
- ListInfo domain type for list ID/name pairs

## Task Commits

Each task was committed atomically (TDD: RED then GREEN):

1. **Task 1: StringOrSlice, InList fields, validation, CollectListNames**
   - RED: `935e61c` (test: add failing tests)
   - GREEN: `518d056` (feat: implement StringOrSlice, InList, validation, CollectListNames)
2. **Task 2: Extend KarakeepAPI interface**
   - RED: `970633a` (test: add failing tests for ListLists, GetListBookmarks)
   - GREEN: `140f424` (feat: extend KarakeepAPI, add ListInfo, update mockAPI)

## Files Created/Modified
- `internal/config/config.go` - StringOrSlice type, InList fields on Conditions and Exceptions
- `internal/config/validate.go` - InList nil-check, empty name validation, CollectListNames
- `internal/config/config_test.go` - Load tests for inList string and list YAML forms
- `internal/config/validate_test.go` - Validation and CollectListNames tests
- `internal/config/testdata/valid_inlist_string.yaml` - Fixture: inList as string
- `internal/config/testdata/valid_inlist_list.yaml` - Fixture: inList as list
- `internal/engine/api.go` - ListLists and GetListBookmarks on KarakeepAPI interface
- `internal/engine/bookmark.go` - ListInfo type with ID and Name
- `internal/engine/api_test.go` - mockAPI with new methods and test cases
- `internal/karakeep/client.go` - Stub implementations for interface compliance

## Decisions Made
- StringOrSlice uses custom `UnmarshalYAML` with `yaml.Node` to properly handle both scalar and sequence YAML while remaining compatible with `KnownFields(true)` strict parsing
- KarakeepClient gets stub methods returning "not yet implemented" errors to satisfy compile-time interface check; plan 02 will implement with actual API calls

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added KarakeepClient stub methods for interface compliance**
- **Found during:** Task 2 (KarakeepAPI interface extension)
- **Issue:** `var _ engine.KarakeepAPI = (*KarakeepClient)(nil)` compile-time check in client.go failed without the new methods
- **Fix:** Added stub ListLists/GetListBookmarks methods with TODO markers for plan 02
- **Files modified:** internal/karakeep/client.go
- **Verification:** `go build ./...` passes
- **Committed in:** 140f424 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary for compilation. Stubs will be replaced in plan 02. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Config types and API interface contracts are complete
- Plan 02 (API client) can implement ListLists/GetListBookmarks against real Karakeep API
- Plan 03 (matcher/run) can implement list membership checks using CollectListNames and the new interface methods

---
*Phase: 01-list-based-bookmark-exclusion*
*Completed: 2026-03-22*
