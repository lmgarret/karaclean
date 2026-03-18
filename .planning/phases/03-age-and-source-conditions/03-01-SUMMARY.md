---
phase: 03-age-and-source-conditions
plan: 01
subsystem: config
tags: [duration, parsing, regex, yaml, validation]

# Dependency graph
requires:
  - phase: 01-config-loading
    provides: Config struct with OlderThan *int, Validate(), testdata YAML fixtures
provides:
  - "internal/duration package with Parse function for compact duration strings"
  - "OlderThan *string type in config.Conditions"
  - "Duration-based validation in validate.go using duration.Parse"
affects: [03-02 matcher, 07 orchestrator]

# Tech tracking
tech-stack:
  added: []
  patterns: [regex-based duration parsing, separate shared package to avoid import cycles]

key-files:
  created:
    - internal/duration/duration.go
    - internal/duration/duration_test.go
  modified:
    - internal/config/config.go
    - internal/config/validate.go
    - internal/config/validate_test.go
    - internal/config/config_test.go
    - internal/config/testdata/valid_full.yaml
    - internal/config/testdata/valid_minimal.yaml

key-decisions:
  - "Duration parser in internal/duration/ (not engine) to avoid import cycle between config and engine"
  - "Zero durations (0h, 0d) accepted as valid per user decision -- matches all bookmarks"
  - "Fixed day counts: mo=30d, y=365d (deterministic, appropriate for GC retention)"

patterns-established:
  - "Shared utility packages in internal/ for cross-package dependencies"
  - "Regex-based parsing with compile-time MustCompile for validation"

requirements-completed: [COND-01]

# Metrics
duration: 2min
completed: 2026-03-18
---

# Phase 03 Plan 01: Duration Parser and OlderThan Migration Summary

**Regex-based duration parser (h/d/w/mo/y) with OlderThan *int to *string migration across config, validation, tests, and YAML fixtures**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-18T14:00:12Z
- **Completed:** 2026-03-18T14:02:44Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- Created internal/duration package with Parse function supporting 5 units (h, d, w, mo, y)
- Migrated OlderThan from *int to *string across entire config subsystem
- Added 4 new validation test cases (weeks, months, invalid format, invalid unit)
- Zero duration now accepted as valid; all 35+ tests pass including full suite

## Task Commits

Each task was committed atomically:

1. **Task 1: Create duration parser package with tests** - `f335f57` (test+feat, TDD)
2. **Task 2: Migrate OlderThan from *int to *string** - `6880c18` (feat)

## Files Created/Modified
- `internal/duration/duration.go` - Parse function with regex for compact duration strings
- `internal/duration/duration_test.go` - 14 table-driven tests covering all units, zero, and invalid inputs
- `internal/config/config.go` - OlderThan *int changed to *string
- `internal/config/validate.go` - Replaced integer check with duration.Parse validation
- `internal/config/validate_test.go` - Migrated intPtr to strPtr, added weeks/months/invalid test cases
- `internal/config/config_test.go` - Updated assertions for string comparison, WrongType test for duration error
- `internal/config/testdata/valid_full.yaml` - olderThan values changed to "30d", "90d"
- `internal/config/testdata/valid_minimal.yaml` - olderThan value changed to "30d"

## Decisions Made
- Duration parser placed in internal/duration/ (shared package) to avoid import cycle between config and engine
- Fixed day multiplication for months (30d) and years (365d) rather than time.AddDate
- Zero durations accepted as valid per user decision in CONTEXT.md

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Duration parser ready for use by matcher in Plan 02
- OlderThan is now *string throughout, matcher can call duration.Parse on validated values
- No blockers for Plan 02 (matcher functions)

---
*Phase: 03-age-and-source-conditions*
*Completed: 2026-03-18*
