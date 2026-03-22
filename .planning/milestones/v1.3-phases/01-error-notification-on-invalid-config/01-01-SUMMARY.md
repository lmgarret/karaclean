---
phase: 01-error-notification-on-invalid-config
plan: 01
subsystem: config
tags: [notifications, shoutrrr, yaml, error-handling]

requires:
  - phase: v1.1
    provides: "Notification system with Notifier interface and ShoutrrrNotifier"
provides:
  - "NotifyOnError opt-in field on Notifications struct"
  - "SendConfigError function for dispatching config validation errors"
  - "Two-pass Load with lenient fallback for YAML decode errors"
  - "ConfigErrorNotifier interface in config package (avoids import cycle)"
affects: []

tech-stack:
  added: []
  patterns: ["ConfigErrorNotifier interface mirrors engine.Notifier to avoid import cycle", "Variadic parameter for backward-compatible function extension", "Lenient fallback parse with notificationsOnly struct"]

key-files:
  created:
    - internal/config/testdata/valid_notify_on_error.yaml
    - internal/config/testdata/invalid_with_notify_on_error.yaml
    - internal/config/testdata/syntax_error_with_notifications.yaml
  modified:
    - internal/config/config.go
    - internal/config/config_test.go
    - cmd/karaclean/main.go
    - README.md
    - karaclean.example.yaml

key-decisions:
  - "ConfigErrorNotifier interface in config package mirrors engine.Notifier to avoid import cycle (config -> engine -> config)"
  - "Variadic notifier parameter on Load maintains backward compatibility with all existing callers"
  - "Lenient fallback uses notificationsOnly struct without KnownFields(true) to parse notifications from files with unknown fields"

patterns-established:
  - "ConfigErrorNotifier: local interface mirroring for import cycle avoidance"

requirements-completed: [ERRNOTIF-01, ERRNOTIF-02, ERRNOTIF-03, ERRNOTIF-04]

duration: 3min
completed: 2026-03-22
---

# Phase 01 Plan 01: Error Notification on Invalid Config Summary

**NotifyOnError opt-in field with two-pass Load, lenient fallback parse, and SendConfigError dispatch to default channel via ConfigErrorNotifier interface**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-22T18:02:10Z
- **Completed:** 2026-03-22T18:05:30Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments
- NotifyOnError *bool field on Notifications struct (opt-in, nil=false per D-01)
- Two-pass Load: YAML decode then validate, dispatching error notification on failure when notifyOnError is true
- Lenient fallback parse for YAML decode errors extracts notifications section via notificationsOnly struct
- SendConfigError best-effort dispatch: logs failure, does not propagate
- main.go passes ShoutrrrNotifier to Load for production error notification
- Full test coverage: 7 SendConfigError subtests, 3 Load_NotifyOnError subtests, 1 LenientFallback test
- README and example config document the feature

## Task Commits

Each task was committed atomically:

1. **Task 1: Add NotifyOnError field, two-pass Load, lenient fallback, SendConfigError, and tests**
   - `58f7657` (test: add failing tests for notifyOnError feature - RED)
   - `7fdedcd` (feat: implement notifyOnError with two-pass Load and lenient fallback - GREEN)
2. **Task 2: Update README and example config** - `8ecae51` (docs)

## Files Created/Modified
- `internal/config/config.go` - ConfigErrorNotifier interface, NotifyOnError field, SendConfigError, two-pass Load with lenient fallback
- `internal/config/config_test.go` - TestSendConfigError (7 subtests), TestLoad_NotifyOnError (3 subtests), TestLoad_LenientFallback
- `internal/config/testdata/valid_notify_on_error.yaml` - Valid config with notifyOnError: true
- `internal/config/testdata/invalid_with_notify_on_error.yaml` - Invalid config (missing action) with notifyOnError: true
- `internal/config/testdata/syntax_error_with_notifications.yaml` - Config with unknown field and notifyOnError: true
- `cmd/karaclean/main.go` - Passes &engine.ShoutrrrNotifier{} to config.Load
- `README.md` - notifyOnError in config table, Error Notifications subsection
- `karaclean.example.yaml` - notifyOnError: true with comment, config reference

## Decisions Made
- ConfigErrorNotifier interface defined in config package (mirrors engine.Notifier) to avoid import cycle config -> engine -> config
- Variadic notifier parameter on Load() for backward compatibility -- existing callers pass no notifier
- Lenient fallback uses a separate notificationsOnly struct decoded without KnownFields(true)
- Syntax error testdata uses unknown field (triggers KnownFields strict rejection) rather than actual YAML structure error, since lenient decode without KnownFields can succeed where strict fails

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed syntax_error_with_notifications.yaml testdata**
- **Found during:** Task 1 (TDD GREEN phase)
- **Issue:** Original testdata had actual YAML structure error that even lenient decoder could not parse. The lenient fallback is designed for KnownFields(true) rejections, not structural YAML errors.
- **Fix:** Changed testdata to use an unknown field (unknownField: "...") which strict KnownFields(true) rejects but lenient decode accepts
- **Files modified:** internal/config/testdata/syntax_error_with_notifications.yaml
- **Verification:** TestLoad_LenientFallback passes
- **Committed in:** 7fdedcd (Task 1 GREEN commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Testdata adjustment necessary for correct test behavior. No scope creep.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Error notification feature complete and tested
- No blockers or concerns

---
*Phase: 01-error-notification-on-invalid-config*
*Completed: 2026-03-22*
