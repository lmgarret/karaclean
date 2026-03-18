---
phase: 01-config-loading-and-validation
plan: 02
subsystem: config
tags: [validation, yaml, go, error-reporting]

# Dependency graph
requires:
  - phase: 01-01
    provides: Config/Rule/Conditions/Exceptions structs, Load() with strict YAML parsing
provides:
  - ValidationError and ValidationErrors types with YAML-matching field paths
  - Config.Validate() method checking action enum, source enum, required fields, non-empty conditions, positive olderThan
  - Load() integration calling Validate() after decode
affects: [02-api-client, 03-age-source-conditions]

# Tech tracking
tech-stack:
  added: []
  patterns: [collected-error-reporting, yaml-field-path-errors, table-driven-tests]

key-files:
  created:
    - internal/config/validate.go
    - internal/config/validate_test.go
  modified:
    - internal/config/config.go

key-decisions:
  - "Validate() returns []ValidationError slice, not error interface, for caller flexibility"
  - "ValidationErrors wraps slice and implements error interface for Load() return"
  - "Source enum values from Karakeep API: rss, web, api, mobile, extension, cli, import"

patterns-established:
  - "Collected validation: all errors gathered before returning, not fail-fast"
  - "YAML-matching field paths: rules[N].field format for user-facing errors"
  - "Table-driven tests: single TestValidate function with named subtests"

requirements-completed: [CONF-02]

# Metrics
duration: 3min
completed: 2026-03-18
---

# Phase 1 Plan 2: Config Validation Summary

**Semantic config validation with collected error reporting using YAML-matching field paths (rules[N].action format)**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-18T10:47:09Z
- **Completed:** 2026-03-18T10:49:57Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- ValidationError and ValidationErrors types provide structured error reporting with YAML-matching field paths
- Validate() checks all semantic rules: action enum (archive/delete), source enum (7 values), required fields, non-empty conditions, positive olderThan
- All validation errors collected and reported at once with "config validation failed:" format
- Load() integrates validation after decode -- invalid configs return ValidationErrors
- 17 table-driven subtests plus integration and regression tests

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement ValidationError types and Config.Validate() method**
   - `61b66e5` (test: RED phase - failing validation tests)
   - `17381e9` (feat: GREEN phase - validate.go + config.go integration)
2. **Task 2: Write comprehensive validation tests**
   - `95cdf83` (test: table-driven tests, integration, regression)

## Files Created/Modified
- `internal/config/validate.go` - ValidationError/ValidationErrors types, Validate() method, action/source enums
- `internal/config/validate_test.go` - 17 table-driven subtests, ValidationErrors.Error() test, Load integration test, regression test
- `internal/config/config.go` - Added Validate() call after YAML decode in Load()

## Decisions Made
- Validate() returns []ValidationError slice (not error interface) for caller flexibility -- callers can inspect individual errors
- ValidationErrors wraps the slice and implements error for Load() return compatibility
- Source enum values taken from Karakeep API: rss, web, api, mobile, extension, cli, import

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 1 complete: config loading with strict parsing (Plan 01) and semantic validation (Plan 02) are both implemented and tested
- Config types and Load() function ready for consumption by Phase 2 (API Client)
- All 20+ tests passing across both plans

---
*Phase: 01-config-loading-and-validation*
*Completed: 2026-03-18*
