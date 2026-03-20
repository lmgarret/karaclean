---
phase: 01-notification-system-send-per-rule-action-summaries-to-configurable-channels-slack-ntfy-telegram-etc-with-global-default-channel-and-per-rule-channel-override
plan: 01
subsystem: config
tags: [shoutrrr, notifications, yaml, validation, ntfy, slack, telegram]

requires:
  - phase: 10-ci
    provides: "Existing Config/Rule structs and Validate() framework"
provides:
  - "Notifications and NotificationChannel types in config.go"
  - "Notify *string field on Rule for per-rule channel override"
  - "validateNotifications function with Shoutrrr URL and channel ref validation"
affects: [01-02-notification-engine, 01-03-run-integration]

tech-stack:
  added: [github.com/nicholas-fedor/shoutrrr v0.14.0]
  patterns: [shoutrrr-url-validation-at-startup, opt-in-nil-notifications]

key-files:
  created: [internal/config/testdata/valid_notifications.yaml]
  modified: [internal/config/config.go, internal/config/validate.go, internal/config/config_test.go, internal/config/validate_test.go, go.mod, go.sum]

key-decisions:
  - "Used ntfy URLs in testdata instead of Slack placeholders (Shoutrrr validates URL format at CreateSender time, fake Slack tokens fail)"
  - "Notifications is *Notifications (nil = opt-in disabled, no errors)"
  - "Notify is *string on Rule (nil = no per-rule override)"

patterns-established:
  - "Shoutrrr URL validation via CreateSender at config load time (fail-fast)"
  - "Notification channel reference validation pattern: default and per-rule notify must reference defined channels"

requirements-completed: [NOTIF-CFG, NOTIF-VAL]

duration: 3min
completed: 2026-03-20
---

# Phase 01 Plan 01: Notification Config & Validation Summary

**Config structs extended with Notifications/NotificationChannel types and Shoutrrr URL validation at startup via validateNotifications**

## Performance

- **Duration:** 3 min
- **Started:** 2026-03-20T11:14:59Z
- **Completed:** 2026-03-20T11:17:59Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Notifications and NotificationChannel types added to config with YAML tags
- Per-rule Notify *string field for channel override routing
- validateNotifications validates Shoutrrr URLs, channel references, and orphan notify fields
- 10+ new test cases covering parsing and validation edge cases
- Shoutrrr v0.14.0 dependency added

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Notifications types and Notify field** (TDD)
   - `42db2d3` (test: failing tests for notification config parsing)
   - `60ed84f` (feat: Notifications/NotificationChannel types and Shoutrrr dep)
2. **Task 2: Add validateNotifications with tests** (TDD)
   - `005f039` (test: failing tests for notification validation)
   - `3d07974` (feat: validateNotifications with Shoutrrr URL and channel ref checks)

## Files Created/Modified
- `internal/config/config.go` - Added Notifications, NotificationChannel structs; Notify field on Rule
- `internal/config/validate.go` - Added validateNotifications function with shoutrrr.CreateSender validation
- `internal/config/config_test.go` - TestLoad_ValidNotifications and TestLoad_NoNotifications
- `internal/config/validate_test.go` - TestValidateNotifications with 8 test cases
- `internal/config/testdata/valid_notifications.yaml` - Test fixture with channels, default, per-rule notify
- `go.mod` / `go.sum` - Added shoutrrr v0.14.0 dependency

## Decisions Made
- Used ntfy URLs in testdata instead of Slack placeholders -- Shoutrrr validates URL format at CreateSender time, fake Slack webhook tokens fail initialization
- Notifications is *Notifications (nil = opt-in disabled, produces no validation errors)
- Notify is *string on Rule (nil = no per-rule channel override)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed testdata Slack URL causing validation failure**
- **Found during:** Task 2 (validateNotifications implementation)
- **Issue:** valid_notifications.yaml used placeholder Slack webhook URL `slack://hook:TOKEN-A-TOKEN-B-TOKEN-C@webhook` which fails Shoutrrr CreateSender validation at config load time
- **Fix:** Changed slack-team channel URL to `ntfy://ntfy.sh/karaclean-slack-team` and updated corresponding test assertion
- **Files modified:** internal/config/testdata/valid_notifications.yaml, internal/config/config_test.go
- **Verification:** `go test ./...` passes
- **Committed in:** 3d07974 (part of Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Necessary fix -- plan's Slack placeholder was syntactically invalid for Shoutrrr. No scope creep.

## Issues Encountered
None beyond the testdata URL fix documented above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Config types ready for Plan 02 (notification engine) to build NotificationSender
- validateNotifications ensures all channel references are valid before engine starts
- shoutrrr.CreateSender pattern established for reuse in engine's Send implementation

---
*Phase: 01-notification-system*
*Completed: 2026-03-20*
